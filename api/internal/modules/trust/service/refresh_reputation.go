package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/trust/domain"
)

type ReputationStore interface {
	ListFacts(ctx context.Context, window domain.Window) ([]domain.Facts, error)
	SaveStats(ctx context.Context, stats []domain.Stats, computedAt time.Time) (int, error)
}

type RefreshReputation struct {
	store ReputationStore
}

func NewRefreshReputation(store ReputationStore) *RefreshReputation {
	return &RefreshReputation{store: store}
}

func (uc *RefreshReputation) Execute(ctx context.Context, now time.Time) (int, error) {
	facts, err := uc.store.ListFacts(ctx, domain.WindowAt(now))
	if err != nil {
		return 0, fmt.Errorf("list reputation facts: %w", err)
	}
	saved, err := uc.store.SaveStats(ctx, domain.Summarize(facts), now)
	if err != nil {
		return 0, fmt.Errorf("save reputation summary: %w", err)
	}
	return saved, nil
}
