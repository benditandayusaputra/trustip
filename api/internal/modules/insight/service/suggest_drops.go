package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/insight/domain"
)

type DropHistoryReader interface {
	ListDropHistories(ctx context.Context, query domain.HistoryQuery) ([]domain.ProductHistory, error)
}

type Clock interface {
	Now() time.Time
}

type SuggestDrops struct {
	histories DropHistoryReader
	clock     Clock
}

func NewSuggestDrops(histories DropHistoryReader, clock Clock) *SuggestDrops {
	return &SuggestDrops{histories: histories, clock: clock}
}

func (uc *SuggestDrops) Execute(
	ctx context.Context, signedIn actor.Actor, partnershipID string,
) ([]domain.DropSuggestion, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return nil, fmt.Errorf("check drop suggestion reader: %w", err)
	}
	now := uc.clock.Now()
	query := domain.NewHistoryQuery(signedIn.Membership.OrganizationID, partnershipID, now)
	products, err := uc.histories.ListDropHistories(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list drop histories of %s: %w", partnershipID, err)
	}
	return domain.SuggestDrops(products, now), nil
}
