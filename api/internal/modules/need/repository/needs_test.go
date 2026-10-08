package repository_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	storeID    = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a73"
	userID     = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"
	snacksID   = "01a1161e-0a49-71f7-a49a-3cf81042e3b0"
	storedHash = "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"
)

func base() time.Time { return time.Date(2026, 10, 20, 3, 0, 0, 0, time.UTC) }

func newNeeds(t *testing.T) *repository.Needs {
	t.Helper()
	pool := dbtest.MigratedPool(t)
	queries := sqlc.New(pool.Querier(t.Context()))
	must(t, queries.CreateUser(t.Context(), sqlc.CreateUserParams{
		ID: userID, Email: "sri@contoh.id", PasswordHash: storedHash, PepperVersion: "v1", FullName: "Bu Sri",
		Whatsapp: "6281234567890", PrivacyAcceptedAt: base(),
	}))
	_, err := queries.CreateOrganization(t.Context(), sqlc.CreateOrganizationParams{
		ID: storeID, Type: "STORE", Name: "Warung Bu Sri", Slug: "warung-bu-sri", Whatsapp: "6281234567890",
		CityCode: "33.74", CreatedAt: base(),
	})
	must(t, err)
	return repository.NewNeeds(pool)
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func posting(index int) domain.Posting {
	draft := domain.Draft{
		Title: "Butuh camilan kantin", CategoryID: snacksID,
		MaxRetailPrice: domain.Amount{Value: 5000, IsGiven: true},
	}
	id := fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", index)
	return draft.Post(domain.Author{NeedID: id, StoreID: storeID, UserID: userID}, base())
}

func TestCreateNeed_BRNED01_DatabaseRefusesTheSixthOpenNeed(t *testing.T) {
	needs := newNeeds(t)
	for index := range domain.MaxOpenPerStore {
		must(t, needs.CreateNeed(t.Context(), posting(index)))
	}

	err := needs.CreateNeed(t.Context(), posting(domain.MaxOpenPerStore))

	if !errors.Is(err, domain.ErrOpenLimit) {
		t.Fatalf("err = %v", err)
	}
}

func TestSaveChange_FRNED06_ClosingFreesASlotAndStaleChangesFail(t *testing.T) {
	needs := newNeeds(t)
	for index := range domain.MaxOpenPerStore {
		must(t, needs.CreateNeed(t.Context(), posting(index)))
	}
	key := domain.NeedKey{StoreID: storeID, NeedID: posting(0).ID}
	need, err := needs.LockNeed(t.Context(), key)
	must(t, err)
	closed, _ := need.Close(base())

	must(t, needs.SaveChange(t.Context(), closed))
	stale := needs.SaveChange(t.Context(), closed)
	sixth := needs.CreateNeed(t.Context(), posting(domain.MaxOpenPerStore))

	if !errors.Is(stale, domain.ErrChangedMeanwhile) || sixth != nil {
		t.Fatalf("stale = %v, sixth = %v", stale, sixth)
	}
}

func TestExpireOpen_BRNED03_MarksNeedsPastTheirShowingPeriod(t *testing.T) {
	needs := newNeeds(t)
	must(t, needs.CreateNeed(t.Context(), posting(0)))

	early, earlyErr := needs.ExpireOpen(t.Context(), base().Add(domain.ShowingPeriod-time.Second))
	late, lateErr := needs.ExpireOpen(t.Context(), base().Add(domain.ShowingPeriod))
	stored, err := needs.GetNeed(t.Context(), domain.NeedKey{StoreID: storeID, NeedID: posting(0).ID})

	if earlyErr != nil || lateErr != nil || err != nil || early != 0 || late != 1 || stored.Status != domain.Expired ||
		!stored.Draft.MaxRetailPrice.IsGiven || stored.CategoryName == "" {
		t.Fatalf("early = %d, late = %d, stored = %+v", early, late, stored)
	}
}

func TestListNeeds_FiltersByStatusInTheStore(t *testing.T) {
	needs := newNeeds(t)
	must(t, needs.CreateNeed(t.Context(), posting(0)))

	open, openErr := needs.ListNeeds(t.Context(), domain.ListQuery{StoreID: storeID, Status: domain.Open, Limit: 20})
	closed, closedErr := needs.ListNeeds(t.Context(), domain.ListQuery{
		StoreID: storeID, Status: domain.Closed, Limit: 20,
	})

	if openErr != nil || closedErr != nil || len(open) != 1 || len(closed) != 0 {
		t.Fatalf("open = %d, closed = %d", len(open), len(closed))
	}
}
