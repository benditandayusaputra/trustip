package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
)

type decideFunc func(ctx context.Context, view domain.View, decider domain.Decider) (domain.Decision, error)

type decisionFlow struct {
	deps   Deps
	reader *GetDrop
	action audit.Action
}

func (f decisionFlow) run(
	ctx context.Context, signedIn actor.Actor, dropID string, decide decideFunc,
) (domain.SealedView, error) {
	decider := domain.Decider{UserID: signedIn.UserID, OrganizationID: signedIn.Membership.OrganizationID}
	err := f.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		key := domain.DropKey{OrganizationID: decider.OrganizationID, DropID: dropID}
		if err := f.deps.Drops.LockDrop(ctx, key); err != nil {
			return fmt.Errorf("lock drop: %w", err)
		}
		view, err := f.deps.Drops.GetDrop(ctx, key)
		if err != nil {
			return fmt.Errorf("load drop: %w", err)
		}
		decision, err := decide(ctx, view, decider)
		if err != nil {
			return err
		}
		return f.record(ctx, view, decision)
	})
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("decide drop %s: %w", dropID, err)
	}
	return f.reader.read(ctx, signedIn, dropID)
}

func (f decisionFlow) record(ctx context.Context, view domain.View, decision domain.Decision) error {
	reversal, err := f.reversalOf(decision)
	if err != nil {
		return err
	}
	err = f.deps.Drops.SaveDecision(ctx, decision, reversal)
	if errors.Is(err, domain.ErrNotPending) {
		return apperror.New(apperror.CodeDocumentNotPending)
	}
	if err != nil {
		return fmt.Errorf("save decision: %w", err)
	}
	if err := f.deps.Sealer.Seal(ctx, view.SealRequestOf(decision)); err != nil {
		return fmt.Errorf("seal %s: %w", decision.Event, err)
	}
	if err := f.announce(ctx, decision); err != nil {
		return err
	}
	return f.audit(ctx, decision)
}

func (f decisionFlow) reversalOf(decision domain.Decision) ([]domain.Movement, error) {
	ids := make([]string, 0, decision.ReversedLines())
	for range decision.ReversedLines() {
		id, err := f.deps.IDs.NewID()
		if err != nil {
			return nil, fmt.Errorf("generate movement id: %w", err)
		}
		ids = append(ids, id)
	}
	return decision.Reversal(ids), nil
}

func (f decisionFlow) announce(ctx context.Context, decision domain.Decision) error {
	if decision.Notice.Type == "" {
		return nil
	}
	if err := f.deps.Announcer.Announce(ctx, decision.Notice); err != nil {
		return fmt.Errorf("announce %s: %w", decision.Notice.Type, err)
	}
	return nil
}

func (f decisionFlow) audit(ctx context.Context, decision domain.Decision) error {
	err := f.deps.Audit.Record(ctx, audit.Entry{
		OrganizationID: decision.ByOrgID, ActorID: decision.By, Action: f.action,
		EntityType: audit.DropEntity, EntityID: decision.DropID, At: decision.At,
	})
	if err != nil {
		return fmt.Errorf("audit %s: %w", f.action, err)
	}
	return nil
}

type ConfirmDrop struct {
	flow decisionFlow
}

func NewConfirmDrop(deps Deps) *ConfirmDrop {
	return &ConfirmDrop{flow: decisionFlow{deps: deps, reader: NewGetDrop(deps), action: audit.DropConfirmed}}
}

func (uc *ConfirmDrop) Execute(ctx context.Context, signedIn actor.Actor, dropID string) (domain.SealedView, error) {
	if err := policy.Require(signedIn, policy.ConfirmDocuments); err != nil {
		return domain.SealedView{}, fmt.Errorf("check drop confirmer: %w", err)
	}
	confirm := func(_ context.Context, view domain.View, decider domain.Decider) (domain.Decision, error) {
		return view.Confirm(decider, uc.flow.deps.Clock.Now())
	}
	return uc.flow.run(ctx, signedIn, dropID, confirm)
}

type CancelDrop struct {
	flow decisionFlow
}

func NewCancelDrop(deps Deps) *CancelDrop {
	return &CancelDrop{flow: decisionFlow{deps: deps, reader: NewGetDrop(deps), action: audit.DropCancelled}}
}

func (uc *CancelDrop) Execute(ctx context.Context, signedIn actor.Actor, dropID string) (domain.SealedView, error) {
	if err := policy.Require(signedIn, policy.RecordDrops); err != nil {
		return domain.SealedView{}, fmt.Errorf("check drop canceller: %w", err)
	}
	return uc.flow.run(ctx, signedIn, dropID, uc.cancel)
}

func (uc *CancelDrop) cancel(ctx context.Context, view domain.View, decider domain.Decider) (domain.Decision, error) {
	isStockUsed, err := uc.flow.deps.Drops.IsStockUsed(ctx, view.ID)
	if err != nil {
		return domain.Decision{}, fmt.Errorf("check stock of drop: %w", err)
	}
	decision, err := view.Cancel(domain.Cancellation{
		Decider: decider, At: uc.flow.deps.Clock.Now(), IsStockUsed: isStockUsed,
	})
	if err != nil {
		return domain.Decision{}, fmt.Errorf("cancel drop: %w", err)
	}
	return decision, nil
}
