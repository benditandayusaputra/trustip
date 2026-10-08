package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
)

type GetSettlement struct {
	settlements SettlementReader
	seals       SealChecker
	clock       Clock
}

func NewGetSettlement(deps Deps) *GetSettlement {
	return &GetSettlement{settlements: deps.Settlements, seals: deps.Seals, clock: deps.Clock}
}

func (uc *GetSettlement) Execute(
	ctx context.Context, signedIn actor.Actor, settlementID string,
) (domain.SealedView, error) {
	if err := policy.Require(signedIn, policy.ViewDocumentValues); err != nil {
		return domain.SealedView{}, fmt.Errorf("check settlement reader: %w", err)
	}
	return uc.read(ctx, signedIn, settlementID)
}

func (uc *GetSettlement) read(
	ctx context.Context, signedIn actor.Actor, settlementID string,
) (domain.SealedView, error) {
	view, err := uc.settlements.GetSettlement(ctx, domain.SettlementKey{
		OrganizationID: signedIn.Membership.OrganizationID, SettlementID: settlementID,
	})
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("get settlement %s: %w", settlementID, err)
	}
	verdict, err := uc.seals.Check(ctx, view.SealCheck())
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("check seal of settlement %s: %w", settlementID, err)
	}
	sealed := domain.SealedView{
		View: view, Standing: view.StandingAt(uc.clock.Now()), Seal: verdict,
		PaymentSeals: make(map[string]domain.SealVerdict, len(view.Payments)),
	}
	for _, payment := range view.Payments {
		if sealed.PaymentSeals[payment.ID], err = uc.seals.Check(ctx, payment.SealCheck()); err != nil {
			return domain.SealedView{}, fmt.Errorf("check seal of payment %s: %w", payment.ID, err)
		}
	}
	return sealed, nil
}
