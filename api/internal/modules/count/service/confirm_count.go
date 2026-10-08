package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
)

type ConfirmCount struct {
	deps   Deps
	reader *GetCount
}

func NewConfirmCount(deps Deps) *ConfirmCount {
	return &ConfirmCount{deps: deps, reader: NewGetCount(deps)}
}

func (uc *ConfirmCount) Execute(ctx context.Context, signedIn actor.Actor, countID string) (domain.SealedView, error) {
	if err := policy.Require(signedIn, policy.ConfirmDocuments); err != nil {
		return domain.SealedView{}, fmt.Errorf("check count confirmer: %w", err)
	}
	decider := domain.Decider{UserID: signedIn.UserID, OrganizationID: signedIn.Membership.OrganizationID}
	key := domain.CountKey{OrganizationID: decider.OrganizationID, CountID: countID}
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.deps.Counts.LockCount(ctx, key); err != nil {
			return fmt.Errorf("lock count: %w", err)
		}
		view, err := uc.deps.Counts.GetCount(ctx, key)
		if err != nil {
			return fmt.Errorf("load count: %w", err)
		}
		decision, err := view.Confirm(decider, uc.deps.Clock.Now())
		if err != nil {
			return fmt.Errorf("confirm count: %w", err)
		}
		return uc.save(ctx, view, decision)
	})
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("confirm count %s: %w", countID, err)
	}
	return uc.reader.read(ctx, signedIn, countID)
}

func (uc *ConfirmCount) save(ctx context.Context, view domain.View, decision domain.Decision) error {
	err := uc.deps.Counts.SaveConfirmation(ctx, decision)
	if errors.Is(err, domain.ErrNotPending) {
		return apperror.New(apperror.CodeDocumentNotPending)
	}
	if err != nil {
		return fmt.Errorf("save confirmation: %w", err)
	}
	if err := uc.deps.Sealer.Seal(ctx, view.SealRequestOf(decision)); err != nil {
		return fmt.Errorf("seal confirmation: %w", err)
	}
	err = uc.deps.Audit.Record(ctx, audit.Entry{
		OrganizationID: decision.ByOrgID, ActorID: decision.By, Action: audit.CountConfirmed,
		EntityType: audit.CountEntity, EntityID: decision.CountID, At: decision.At,
	})
	if err != nil {
		return fmt.Errorf("audit confirmation: %w", err)
	}
	return nil
}
