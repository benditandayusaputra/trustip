package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
)

type MarkPaid struct {
	deps   Deps
	reader *GetSettlement
}

func NewMarkPaid(deps Deps) *MarkPaid {
	return &MarkPaid{deps: deps, reader: NewGetSettlement(deps)}
}

func (uc *MarkPaid) Execute(
	ctx context.Context, signedIn actor.Actor, settlementID string,
) (domain.SealedView, error) {
	if err := policy.Require(signedIn, policy.ReviewPayments); err != nil {
		return domain.SealedView{}, fmt.Errorf("check payment reviewer: %w", err)
	}
	consignor := domain.Decider{UserID: signedIn.UserID, OrganizationID: signedIn.Membership.OrganizationID}
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		payable, err := uc.deps.Settlements.LockPayable(ctx, domain.PayableKey{
			OrganizationID: consignor.OrganizationID, SettlementID: settlementID,
		})
		if err != nil {
			return fmt.Errorf("lock settlement: %w", err)
		}
		recording, err := uc.recording()
		if err != nil {
			return err
		}
		payment, err := payable.SettleDirectly(consignor, recording)
		if err != nil {
			return fmt.Errorf("settle directly: %w", err)
		}
		return uc.save(ctx, payment)
	})
	if err != nil {
		return domain.SealedView{}, fmt.Errorf("mark settlement %s paid: %w", settlementID, err)
	}
	return uc.reader.read(ctx, signedIn, settlementID)
}

func (uc *MarkPaid) recording() (domain.PaymentRecording, error) {
	id, err := uc.deps.IDs.NewID()
	if err != nil {
		return domain.PaymentRecording{}, fmt.Errorf("generate payment id: %w", err)
	}
	key, err := uc.deps.IDs.NewID()
	if err != nil {
		return domain.PaymentRecording{}, fmt.Errorf("generate payment key: %w", err)
	}
	return domain.PaymentRecording{
		ID: id, IdempotencyKey: key, Fingerprint: make([]byte, fingerprintLength),
		At: uc.deps.Clock.Now(),
	}, nil
}

const fingerprintLength = 32

func (uc *MarkPaid) save(ctx context.Context, payment domain.Payment) error {
	if err := uc.deps.Settlements.CreatePayment(ctx, payment); err != nil {
		return fmt.Errorf("save payment: %w", err)
	}
	approval := domain.Approval{SettlementID: payment.SettlementID, By: payment.RecordedBy, At: payment.RecordedAt}
	if err := (reviewer{deps: uc.deps}).settle(ctx, approval); err != nil {
		return err
	}
	if err := uc.deps.Sealer.Seal(ctx, payment.SealRequest()); err != nil {
		return fmt.Errorf("seal payment: %w", err)
	}
	notice := payment.ReviewNotice(domain.Review{
		Status: domain.PaymentConfirmed, By: payment.RecordedBy, Event: "PAYMENT_CONFIRMED",
	})
	if err := uc.deps.Announcer.Announce(ctx, notice); err != nil {
		return fmt.Errorf("announce payment: %w", err)
	}
	return recordPaymentAudit(ctx, uc.deps.Audit, payment, audit.PaymentConfirmed)
}
