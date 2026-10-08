package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
)

type ConfirmPayment struct {
	deps   Deps
	reader *GetSettlement
}

func NewConfirmPayment(deps Deps) *ConfirmPayment {
	return &ConfirmPayment{deps: deps, reader: NewGetSettlement(deps)}
}

func (uc *ConfirmPayment) Execute(
	ctx context.Context, signedIn actor.Actor, paymentID string,
) (domain.SealedView, error) {
	if err := policy.Require(signedIn, policy.ReviewPayments); err != nil {
		return domain.SealedView{}, fmt.Errorf("check payment reviewer: %w", err)
	}
	consignor := domain.Decider{UserID: signedIn.UserID, OrganizationID: signedIn.Membership.OrganizationID}
	review := reviewer{deps: uc.deps}
	var settlementID string
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		payment, err := review.lock(ctx, domain.PaymentKey{
			OrganizationID: consignor.OrganizationID, PaymentID: paymentID,
		})
		if err != nil {
			return err
		}
		settlementID = payment.SettlementID
		decision, err := payment.Confirm(consignor, uc.deps.Clock.Now())
		if err != nil {
			return fmt.Errorf("confirm payment: %w", err)
		}
		return review.save(ctx, payment, decision)
	})
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("confirm payment %s: %w", paymentID, err)
	}
	return uc.reader.read(ctx, signedIn, settlementID)
}
