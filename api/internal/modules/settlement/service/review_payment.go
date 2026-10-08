package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
)

type reviewer struct {
	deps Deps
}

func (r reviewer) lock(ctx context.Context, key domain.PaymentKey) (domain.Payment, error) {
	payment, err := r.deps.Settlements.LockPayment(ctx, key)
	if err != nil {
		return domain.Payment{}, fmt.Errorf("lock payment: %w", err)
	}
	return payment, nil
}

func (r reviewer) save(ctx context.Context, payment domain.Payment, review domain.Review) error {
	err := r.deps.Settlements.SaveReview(ctx, review)
	if errors.Is(err, domain.ErrPaymentDecided) {
		return apperror.New(apperror.CodeDocumentNotPending)
	}
	if err != nil {
		return fmt.Errorf("save payment review: %w", err)
	}
	if review.Status == domain.PaymentConfirmed {
		if err := r.settle(ctx, review.Approval()); err != nil {
			return err
		}
	}
	if err := r.deps.Sealer.Seal(ctx, payment.ReviewSealRequest(review)); err != nil {
		return fmt.Errorf("seal payment review: %w", err)
	}
	if err := r.deps.Announcer.Announce(ctx, payment.ReviewNotice(review)); err != nil {
		return fmt.Errorf("announce payment review: %w", err)
	}
	err = r.deps.Audit.Record(ctx, audit.Entry{
		OrganizationID: review.ByOrgID, ActorID: review.By, Action: audit.Action(review.Event),
		EntityType: audit.PaymentEntity, EntityID: review.PaymentID, At: review.At,
	})
	if err != nil {
		return fmt.Errorf("audit payment review: %w", err)
	}
	return nil
}

func (r reviewer) settle(ctx context.Context, approval domain.Approval) error {
	if err := r.deps.Settlements.MarkPaid(ctx, approval.SettlementID, approval.At); err != nil {
		return fmt.Errorf("mark settlement paid: %w", err)
	}
	if err := r.deps.Approver.ApproveSources(ctx, approval); err != nil {
		return fmt.Errorf("approve source counts: %w", err)
	}
	return nil
}
