package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
)

type GetBlockers struct {
	deps EndDeps
}

func NewGetBlockers(deps EndDeps) *GetBlockers {
	return &GetBlockers{deps: deps}
}

func (uc *GetBlockers) Execute(
	ctx context.Context, signedIn actor.Actor, partnershipID string,
) ([]domain.Blocker, error) {
	if err := policy.Require(signedIn, policy.EndPartnerships); err != nil {
		return nil, fmt.Errorf("check partnership ender: %w", err)
	}
	var blockers []domain.Blocker
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		key := endKeyOf(keyOf(signedIn, partnershipID), uc.deps.Clock.Now())
		state, err := uc.deps.Partnerships.LockForEnding(ctx, key)
		if err != nil {
			return fmt.Errorf("lock partnership: %w", err)
		}
		blockers = state.Blockers
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list blockers of %s: %w", partnershipID, err)
	}
	return blockers, nil
}
