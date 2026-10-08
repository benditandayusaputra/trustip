package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
)

type transition func(need domain.Need, now time.Time) (domain.Change, error)

type ChangeNeed struct {
	deps   Deps
	apply  transition
	action audit.Action
}

func NewExtendNeed(deps Deps) *ChangeNeed {
	return &ChangeNeed{deps: deps, apply: domain.Need.Extend, action: audit.NeedExtended}
}

func NewFulfillNeed(deps Deps) *ChangeNeed {
	return &ChangeNeed{deps: deps, apply: domain.Need.Fulfill, action: audit.NeedFulfilled}
}

func NewCloseNeed(deps Deps) *ChangeNeed {
	return &ChangeNeed{deps: deps, apply: domain.Need.Close, action: audit.NeedClosed}
}

func (uc *ChangeNeed) Execute(ctx context.Context, signedIn actor.Actor, needID string) (domain.Need, error) {
	if err := policy.Require(signedIn, policy.ManageNeeds); err != nil {
		return domain.Need{}, fmt.Errorf("check need owner: %w", err)
	}
	now := uc.deps.Clock.Now()
	var need domain.Need
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.change(ctx, keyOf(signedIn, needID), now); err != nil {
			return err
		}
		var err error
		need, err = finishNeedEvent(ctx, uc.deps, needEvent{
			signedIn: signedIn, needID: needID, action: uc.action, at: now,
		})
		return err
	})
	if err != nil {
		return domain.Need{}, fmt.Errorf("change need %s: %w", needID, err)
	}
	return need, nil
}

func (uc *ChangeNeed) change(ctx context.Context, key domain.NeedKey, now time.Time) error {
	locked, err := uc.deps.Writer.LockNeed(ctx, key)
	if err != nil {
		return fmt.Errorf("lock need: %w", err)
	}
	change, err := uc.apply(locked, now)
	if err != nil {
		return err
	}
	err = uc.deps.Writer.SaveChange(ctx, change)
	switch {
	case errors.Is(err, domain.ErrOpenLimit):
		return apperror.New(apperror.CodeNeedLimitReached)
	case errors.Is(err, domain.ErrChangedMeanwhile):
		return apperror.New(apperror.CodeNeedNotOpen)
	case err != nil:
		return fmt.Errorf("save need change: %w", err)
	}
	return nil
}
