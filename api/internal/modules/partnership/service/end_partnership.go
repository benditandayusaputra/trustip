package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
)

type EndCommand struct {
	PartnershipID string
	Reason        string
}

type EndPartnership struct {
	deps EndDeps
}

func NewEndPartnership(deps EndDeps) *EndPartnership {
	return &EndPartnership{deps: deps}
}

func (uc *EndPartnership) Execute(
	ctx context.Context, signedIn actor.Actor, cmd EndCommand,
) (domain.Partnership, error) {
	if err := policy.Require(signedIn, policy.EndPartnerships); err != nil {
		return domain.Partnership{}, fmt.Errorf("check partnership ender: %w", err)
	}
	key := keyOf(signedIn, cmd.PartnershipID)
	if err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error { return uc.settle(ctx, key) }); err != nil {
		return domain.Partnership{}, fmt.Errorf("settle before ending %s: %w", cmd.PartnershipID, err)
	}
	ender := domain.Ender{UserID: signedIn.UserID, OrganizationID: signedIn.Membership.OrganizationID}
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		now := uc.deps.Clock.Now()
		state, err := uc.deps.Partnerships.LockForEnding(ctx, endKeyOf(key, now))
		if err != nil {
			return fmt.Errorf("lock partnership: %w", err)
		}
		ending, err := state.End(ender, cmd.Reason, now)
		if err != nil {
			return fmt.Errorf("end partnership: %w", err)
		}
		return uc.save(ctx, ending)
	})
	if err != nil {
		return domain.Partnership{}, fmt.Errorf("end partnership %s: %w", cmd.PartnershipID, err)
	}
	partnership, err := uc.deps.Reader.GetPartnership(ctx, key)
	if err != nil {
		return domain.Partnership{}, fmt.Errorf("read ended partnership: %w", err)
	}
	return partnership, nil
}

func (uc *EndPartnership) settle(ctx context.Context, key domain.Key) error {
	now := uc.deps.Clock.Now()
	state, err := uc.deps.Partnerships.LockForEnding(ctx, endKeyOf(key, now))
	if err != nil {
		return fmt.Errorf("lock partnership: %w", err)
	}
	if state.Status != domain.Active {
		return apperror.New(apperror.CodePartnershipNotActive)
	}
	if err := uc.deps.Settler.SettleOutstanding(ctx, key.PartnershipID, now); err != nil {
		return fmt.Errorf("issue outstanding settlement: %w", err)
	}
	return nil
}

func (uc *EndPartnership) save(ctx context.Context, ending domain.Ending) error {
	if err := uc.deps.Partnerships.SaveEnding(ctx, ending); err != nil {
		return fmt.Errorf("save ending: %w", err)
	}
	if err := uc.deps.Sealer.Seal(ctx, ending.SealRequest()); err != nil {
		return fmt.Errorf("seal ending: %w", err)
	}
	if err := uc.deps.Announcer.Announce(ctx, ending.Notice()); err != nil {
		return fmt.Errorf("announce ending: %w", err)
	}
	err := uc.deps.Audit.Record(ctx, audit.Entry{
		OrganizationID: ending.EndedByOrgID, ActorID: ending.EndedBy, Action: audit.PartnershipEnded,
		EntityType: audit.PartnershipEntity, EntityID: ending.PartnershipID, At: ending.At,
	})
	if err != nil {
		return fmt.Errorf("audit ending: %w", err)
	}
	return nil
}
