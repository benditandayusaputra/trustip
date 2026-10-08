package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
)

type RejectPaymentCommand struct {
	PaymentID string
	Reason    string
}

type RejectPayment struct {
	deps   Deps
	reader *GetSettlement
}

func NewRejectPayment(deps Deps) *RejectPayment {
	return &RejectPayment{deps: deps, reader: NewGetSettlement(deps)}
}

func (uc *RejectPayment) Execute(
	ctx context.Context, signedIn actor.Actor, cmd RejectPaymentCommand,
) (domain.SealedView, error) {
	if err := policy.Require(signedIn, policy.ReviewPayments); err != nil {
		return domain.SealedView{}, fmt.Errorf("check payment reviewer: %w", err)
	}
	consignor := domain.Decider{UserID: signedIn.UserID, OrganizationID: signedIn.Membership.OrganizationID}
	review := reviewer{deps: uc.deps}
	var settlementID string
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		payment, err := review.lock(ctx, domain.PaymentKey{
			OrganizationID: consignor.OrganizationID, PaymentID: cmd.PaymentID,
		})
		if err != nil {
			return err
		}
		settlementID = payment.SettlementID
		decision, err := payment.Reject(consignor, cmd.Reason, uc.deps.Clock.Now())
		if err != nil {
			return fmt.Errorf("reject payment: %w", err)
		}
		return review.save(ctx, payment, decision)
	})
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("reject payment %s: %w", cmd.PaymentID, err)
	}
	return uc.reader.read(ctx, signedIn, settlementID)
}
