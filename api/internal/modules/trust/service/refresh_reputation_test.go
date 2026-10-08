package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/trust/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/trust/service"
)

var errDown = errors.New("store down")

type fakeReputation struct {
	windows []domain.Window
	saved   []domain.Stats
	savedAt time.Time
	fail    error
}

func (f *fakeReputation) ListFacts(_ context.Context, window domain.Window) ([]domain.Facts, error) {
	f.windows = append(f.windows, window)
	return []domain.Facts{{OrganizationID: "store-1", DueSettlements: 4, OnTimeSettlements: 3}}, f.fail
}

func (f *fakeReputation) SaveStats(_ context.Context, stats []domain.Stats, computedAt time.Time) (int, error) {
	f.saved, f.savedAt = stats, computedAt
	return len(stats), nil
}

func TestRefreshReputation_BRTRS02_StoresTheDailySummary(t *testing.T) {
	store := &fakeReputation{}
	now := time.Date(2026, 11, 10, 17, 5, 0, 0, time.UTC)

	refreshed, err := service.NewRefreshReputation(store).Execute(t.Context(), now)

	if err != nil || refreshed != 1 || store.saved[0].OnTimeBPS != 7500 || !store.savedAt.Equal(now) {
		t.Fatalf("refreshed = %d, err = %v, saved = %+v", refreshed, err, store.saved)
	}
	if store.windows[0] != domain.WindowAt(now) {
		t.Fatalf("window = %+v", store.windows[0])
	}
}

func TestRefreshReputation_FailureIsWrapped(t *testing.T) {
	now := time.Date(2026, 11, 10, 17, 5, 0, 0, time.UTC)

	_, err := service.NewRefreshReputation(&fakeReputation{fail: errDown}).Execute(t.Context(), now)

	if !errors.Is(err, errDown) {
		t.Fatalf("err = %v, want wrapped store failure", err)
	}
}
