package repository_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	userID        = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"
	consignorID   = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a71"
	storeID       = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a73"
	partnershipID = "0192f0c4-0000-7000-8000-000000000501"
	storedHash    = "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"
)

func base() time.Time {
	return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
}

func privateKey() ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, ed25519.SeedSize))
}

func sign(digest []byte) domain.Signature {
	return domain.Signature{KeyID: "v1", Bytes: ed25519.Sign(privateKey(), digest)}
}

type world struct {
	t       *testing.T
	pool    *database.Pool
	queries *sqlc.Queries
	seals   *repository.Seals
}

func newWorld(t *testing.T) world {
	t.Helper()
	pool := dbtest.MigratedPool(t)
	scene := world{t: t, pool: pool, queries: sqlc.New(pool.Querier(t.Context())), seals: repository.NewSeals(pool)}
	scene.must(scene.queries.CreateUser(t.Context(), sqlc.CreateUserParams{
		ID: userID, Email: "anton@contoh.id", PasswordHash: storedHash, PepperVersion: "v1",
		FullName: "Anton", Whatsapp: "6281234567890", PrivacyAcceptedAt: base(),
	}))
	for id, kind := range map[string]string{consignorID: "CONSIGNOR", storeID: "STORE"} {
		_, err := scene.queries.CreateOrganization(t.Context(), sqlc.CreateOrganizationParams{
			ID: id, Type: kind, Name: "Usaha " + kind, Slug: "usaha-" + id[len(id)-2:], Whatsapp: "6281234567890",
			CityCode: "33.74", CreatedAt: base(),
		})
		scene.must(err)
	}
	scene.must(scene.queries.CreatePartnership(t.Context(), sqlc.CreatePartnershipParams{
		ID: partnershipID, ConsignorOrgID: consignorID, StoreOrgID: storeID, StartedAt: base(),
	}))
	return scene
}

func (w world) must(err error) {
	w.t.Helper()
	if err != nil {
		w.t.Fatalf("setup: %v", err)
	}
}

func request(index int) domain.Request {
	return domain.Request{
		PartnershipID: partnershipID, Event: domain.AgreementAccepted,
		Document: domain.Document{
			Type: "AGREEMENT", ID: fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", index),
			Number: fmt.Sprintf("KS-%d", index),
		},
		Content: map[string]any{"version": index}, ActorUserID: userID, ActorOrgID: storeID,
		At: base().Add(time.Duration(index) * time.Minute),
	}
}

func (w world) seal(ctx context.Context, index int) domain.Seal {
	w.t.Helper()
	seal, err := w.trySeal(ctx, index)
	if err != nil {
		w.t.Fatalf("seal %d: %v", index, err)
	}
	return seal
}

func (w world) trySeal(ctx context.Context, index int) (domain.Seal, error) {
	head, err := w.seals.LockHead(ctx, partnershipID)
	if err != nil {
		return domain.Seal{}, fmt.Errorf("lock: %w", err)
	}
	entry, err := domain.NewEntry(request(index), head)
	if err != nil {
		return domain.Seal{}, fmt.Errorf("entry: %w", err)
	}
	seal, err := domain.Sealed(fmt.Sprintf("0192f0c4-0000-7000-9000-%012d", index), entry, sign)
	if err != nil {
		return domain.Seal{}, fmt.Errorf("sealed: %w", err)
	}
	if err := w.seals.Append(ctx, seal); err != nil {
		return domain.Seal{}, fmt.Errorf("append: %w", err)
	}
	return seal, nil
}

func (w world) chain() []domain.Seal {
	w.t.Helper()
	seals, err := w.seals.ListChain(w.t.Context(), partnershipID)
	if err != nil {
		w.t.Fatalf("list: %v", err)
	}
	return seals
}

func TestSeal_RSEC30_StoredChainVerifiesAndIsSigned(t *testing.T) {
	scene := newWorld(t)
	scene.seal(t.Context(), 1)
	scene.seal(t.Context(), 2)

	stored := scene.chain()

	if len(stored) != 2 || domain.VerifyChain(stored) != nil {
		t.Fatalf("stored = %+v, verify = %v", stored, domain.VerifyChain(stored))
	}
	for _, seal := range stored {
		if !ed25519.Verify(privateKey().Public().(ed25519.PublicKey), seal.EntryHash, seal.Signature.Bytes) {
			t.Fatalf("signature of %d does not verify", seal.Entry.Seq)
		}
	}
}

func TestSeal_RSEC30_ConcurrentEventsDoNotForkTheChain(t *testing.T) {
	scene := newWorld(t)
	var group sync.WaitGroup
	failures := make(chan error, 5)
	for index := range 5 {
		group.Go(func() {
			failures <- scene.pool.WithinTx(t.Context(), func(ctx context.Context) error {
				_, err := scene.trySeal(ctx, index+1)
				return err
			})
		})
	}
	group.Wait()
	close(failures)
	for err := range failures {
		scene.must(err)
	}

	stored := scene.chain()

	if len(stored) != 5 || domain.VerifyChain(stored) != nil {
		t.Fatalf("stored = %d, verify = %v", len(stored), domain.VerifyChain(stored))
	}
}

func TestSeal_BRSIG03_RolledBackEventLeavesNoSeal(t *testing.T) {
	scene := newWorld(t)
	eventFailed := errors.New("event failed")

	err := scene.pool.WithinTx(t.Context(), func(ctx context.Context) error {
		scene.seal(ctx, 1)
		return eventFailed
	})
	after := scene.seal(t.Context(), 2)

	if !errors.Is(err, eventFailed) || after.Entry.Seq != 1 || len(scene.chain()) != 1 {
		t.Fatalf("err = %v, after = %+v", err, after)
	}
}

func TestSeals_RSEC31_UpdateAndDeleteAreRejected(t *testing.T) {
	scene := newWorld(t)
	sealed := scene.seal(t.Context(), 1)

	_, updateErr := scene.queries.TamperSealNumber(t.Context(), sqlc.TamperSealNumberParams{
		ID: sealed.ID, EntityNumber: "KS-9",
	})
	_, deleteErr := scene.queries.DeleteSealsOfPartnership(t.Context(), partnershipID)

	stored := scene.chain()
	if updateErr == nil || deleteErr == nil || len(stored) != 1 || stored[0].Entry.Document.Number != "KS-1" {
		t.Fatalf("update = %v, delete = %v", updateErr, deleteErr)
	}
}

func TestSeals_BRSIG04_DemoPartnershipSealsCanBeReset(t *testing.T) {
	scene := newWorld(t)
	scene.seal(t.Context(), 1)
	for _, id := range []string{consignorID, storeID} {
		scene.must(scene.queries.SetOrganizationDemo(t.Context(), sqlc.SetOrganizationDemoParams{ID: id, IsDemo: true}))
	}

	deleted, err := scene.queries.DeleteSealsOfPartnership(t.Context(), partnershipID)

	if err != nil || deleted != 1 {
		t.Fatalf("deleted = %d, err = %v", deleted, err)
	}
}
