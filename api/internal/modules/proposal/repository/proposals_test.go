package repository_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	userID          = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"
	consignorID     = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a71"
	otherConsignor  = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a72"
	storeID         = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a73"
	otherStore      = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a74"
	breadID         = "0192f0c4-0000-7000-8000-000000000101"
	archivedID      = "0192f0c4-0000-7000-8000-000000000102"
	foreignID       = "0192f0c4-0000-7000-8000-000000000103"
	breadCategoryID = "01a1161e-0a30-752d-ab9b-7397976f0dba"
	storedHash      = "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"
)

func base() time.Time {
	return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
}

type world struct {
	t         *testing.T
	pool      *database.Pool
	queries   *sqlc.Queries
	proposals *repository.Proposals
}

func newWorld(t *testing.T) world {
	t.Helper()
	pool := dbtest.MigratedPool(t)
	scene := world{
		t: t, pool: pool, queries: sqlc.New(pool.Querier(t.Context())), proposals: repository.NewProposals(pool),
	}
	scene.must(scene.queries.CreateUser(t.Context(), sqlc.CreateUserParams{
		ID: userID, Email: "anton@contoh.id", PasswordHash: storedHash, PepperVersion: "v1",
		FullName: "Anton", Whatsapp: "6281234567890", PrivacyAcceptedAt: base(),
	}))
	for id, kind := range map[string]string{
		consignorID: "CONSIGNOR", otherConsignor: "CONSIGNOR", storeID: "STORE", otherStore: "STORE",
	} {
		_, err := scene.queries.CreateOrganization(t.Context(), sqlc.CreateOrganizationParams{
			ID: id, Type: kind, Name: "Usaha " + id[len(id)-2:], Slug: "usaha-" + id[len(id)-2:],
			Whatsapp: "6281234567890", CityCode: "33.74", CreatedAt: base(),
		})
		scene.must(err)
	}
	scene.product(breadID, consignorID)
	scene.product(archivedID, consignorID)
	scene.product(foreignID, otherConsignor)
	_, err := scene.queries.SetProductArchived(t.Context(), sqlc.SetProductArchivedParams{
		ID: archivedID, ConsignorOrgID: consignorID, IsArchived: true, UpdatedAt: base(),
	})
	scene.must(err)
	return scene
}

func (w world) must(err error) {
	w.t.Helper()
	if err != nil {
		w.t.Fatalf("setup: %v", err)
	}
}

func (w world) product(id, consignor string) {
	w.t.Helper()
	_, err := w.queries.SaveProduct(w.t.Context(), sqlc.SaveProductParams{
		ID: id, ConsignorOrgID: consignor, Sku: "SKU-" + id[len(id)-3:], Name: "Roti " + id[len(id)-3:],
		CategoryID: breadCategoryID, Unit: "pcs", SuggestedRetailPrice: 2000, SuggestedDepositPrice: 1750,
		ChangedAt: base(),
	})
	w.must(err)
}

func invitation(index int, sender string, productIDs ...string) domain.Invitation {
	items := make([]domain.Item, 0, len(productIDs))
	for _, productID := range productIDs {
		terms, _ := pricing.NewTerms(pricing.TermsInput{Model: "FIXED_DEPOSIT", RetailPrice: 2000, DepositPrice: 1750})
		items = append(items, domain.Item{ProductID: productID, Action: domain.Add, Terms: terms})
	}
	draft := domain.RevisionDraft{
		Terms: terms.General{Schedule: terms.ScheduleWeekly, PaymentTermsDays: 7, Rhythm: terms.RhythmDaily},
		Items: items, Message: "Salam kenal.",
	}
	senderType := actor.Consignor
	if sender == storeID || sender == otherStore {
		senderType = actor.Store
	}
	created, _ := draft.Invite(domain.InvitationParams{
		IDs: domain.InvitationIDs{
			ProposalID: fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", index),
			RevisionID: fmt.Sprintf("0192f0c4-0000-7000-9000-%012d", index), Code: "ABCDEFGHJ" + string("KLMNP"[index]),
		},
		Sender:   domain.Sender{UserID: userID, OrganizationID: sender, OrganizationType: senderType},
		Sequence: index, At: base(),
	})
	return created
}

func (w world) invite(created domain.Invitation) {
	w.t.Helper()
	if err := w.proposals.Open(w.t.Context(), created.Opening()); err != nil {
		w.t.Fatalf("invite: %v", err)
	}
}

func storeClaim(created domain.Invitation, store string) domain.Claim {
	return domain.Claim{
		ProposalID: created.IDs.ProposalID, ConsignorID: created.ConsignorID(), StoreID: store, ClaimerID: store,
		At: base().Add(time.Hour),
	}
}

func TestCreateInvitation_FRPRP02_SenderReadsTheProductsAndTerms(t *testing.T) {
	scene := newWorld(t)
	created := invitation(1, consignorID, breadID)
	scene.invite(created)

	got, err := scene.proposals.GetProposal(t.Context(), domain.ProposalKey{
		OrganizationID: consignorID, ProposalID: created.IDs.ProposalID,
	})

	if err != nil || got.Number != "AJ-202610-0001" || got.InviteCode != created.IDs.Code ||
		got.Store.OrganizationID != "" ||
		len(got.Revision.Items) != 1 || got.Revision.Items[0].SKU != "SKU-101" ||
		got.Revision.Items[0].Terms.DepositPrice != 1750 || got.Revision.Terms.Schedule != terms.ScheduleWeekly {
		t.Fatalf("got = %+v, err = %v", got, err)
	}
}

func TestCreateInvitation_BRPRP01_ArchivedOrForeignProductsAreRejected(t *testing.T) {
	for _, productID := range []string{archivedID, foreignID} {
		scene := newWorld(t)
		created := invitation(1, consignorID, breadID, productID)

		err := scene.pool.WithinTx(t.Context(), func(ctx context.Context) error {
			return scene.proposals.Open(ctx, created.Opening())
		})

		if !errors.Is(err, domain.ErrUnknownProduct) {
			t.Fatalf("%s: err = %v", productID, err)
		}
	}
}

func TestNextNumber_FRPRP07_CountsPerOrganizationAndMonth(t *testing.T) {
	scene := newWorld(t)
	key := domain.SequenceKey{OrganizationID: consignorID, Prefix: "AJ", Period: "202610"}

	first, firstErr := scene.proposals.NextNumber(t.Context(), key)
	second, secondErr := scene.proposals.NextNumber(t.Context(), key)
	key.Period = "202611"
	fresh, freshErr := scene.proposals.NextNumber(t.Context(), key)

	if errors.Join(firstErr, secondErr, freshErr) != nil || first != 1 || second != 2 || fresh != 1 {
		t.Fatalf("first = %d, second = %d, fresh = %d", first, second, fresh)
	}
}

func TestClaimInvitation_BRPRP07_StoreBecomesAPartyOnce(t *testing.T) {
	scene := newWorld(t)
	created := invitation(1, consignorID, breadID)
	scene.invite(created)

	err := scene.proposals.ClaimInvitation(t.Context(), storeClaim(created, storeID))
	again := scene.proposals.ClaimInvitation(t.Context(), storeClaim(created, otherStore))
	view, viewErr := scene.proposals.GetInvitation(t.Context(), created.IDs.Code)
	got, getErr := scene.proposals.GetProposal(t.Context(), domain.ProposalKey{
		OrganizationID: storeID, ProposalID: created.IDs.ProposalID,
	})

	if err != nil || !errors.Is(again, domain.ErrClaimRaced) || viewErr != nil || view.StoreID != storeID ||
		view.ItemCount != 1 || getErr != nil || got.AwaitingOrgID != storeID {
		t.Fatalf("err = %v, again = %v, view = %+v, got = %+v %v", err, again, view, got, getErr)
	}
}

func TestClaimInvitation_RCC74_ConcurrentClaimsOnlyOneWins(t *testing.T) {
	scene := newWorld(t)
	created := invitation(1, consignorID, breadID)
	scene.invite(created)
	var group sync.WaitGroup
	outcomes := make(chan error, 2)
	for _, store := range []string{storeID, otherStore} {
		group.Go(func() {
			outcomes <- scene.proposals.ClaimInvitation(t.Context(), storeClaim(created, store))
		})
	}
	group.Wait()
	close(outcomes)

	won, lost := 0, 0
	for err := range outcomes {
		switch {
		case err == nil:
			won++
		case errors.Is(err, domain.ErrClaimRaced):
			lost++
		}
	}
	if won != 1 || lost != 1 {
		t.Fatalf("won = %d, lost = %d", won, lost)
	}
}

func TestClaimInvitation_BRPRP04_SecondOpenProposalForThePairIsRefused(t *testing.T) {
	scene := newWorld(t)
	first, second := invitation(1, consignorID, breadID), invitation(2, consignorID)
	scene.invite(first)
	scene.invite(second)
	scene.must(scene.proposals.ClaimInvitation(t.Context(), storeClaim(first, storeID)))

	err := scene.proposals.ClaimInvitation(t.Context(), storeClaim(second, storeID))

	if !errors.Is(err, domain.ErrAlreadyOpen) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetProposal_RSEC38_StoreBeforeTheClaimGetsNotFound(t *testing.T) {
	scene := newWorld(t)
	created := invitation(1, consignorID, breadID)
	scene.invite(created)

	_, err := scene.proposals.GetProposal(t.Context(), domain.ProposalKey{
		OrganizationID: storeID, ProposalID: created.IDs.ProposalID,
	})

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateInvitation_FRPRP07_StoreInvitationWithoutProducts(t *testing.T) {
	scene := newWorld(t)
	created := invitation(1, storeID)
	scene.invite(created)

	view, err := scene.proposals.GetInvitation(t.Context(), created.IDs.Code)

	if err != nil || view.Sender.Type != actor.Store || view.ConsignorID != "" || view.ItemCount != 0 {
		t.Fatalf("view = %+v, err = %v", view, err)
	}
}

func TestHasActivePartnership_FalseWithoutPartnership(t *testing.T) {
	scene := newWorld(t)

	pair := domain.Pair{ConsignorID: consignorID, StoreID: storeID}

	found, err := scene.proposals.HasActivePartnership(t.Context(), pair)

	if err != nil || found {
		t.Fatalf("found = %v, err = %v", found, err)
	}
}
