package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Settlements) FindPaymentByKey(
	ctx context.Context, key domain.PaymentKeyRef,
) (domain.StoredPayment, bool, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).FindPaymentByKey(readCtx, sqlc.FindPaymentByKeyParams{
		OrganizationID: key.OrganizationID, IdempotencyKey: key.Key,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.StoredPayment{}, false, nil
	}
	if err != nil {
		return domain.StoredPayment{}, false, fmt.Errorf("find payment by idempotency key: %w", err)
	}
	return domain.StoredPayment{
		PaymentID: row.ID, SettlementID: row.SettlementID, Fingerprint: row.RequestFingerprint,
	}, true, nil
}

func (r *Settlements) LockPayable(ctx context.Context, key domain.PayableKey) (domain.Payable, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.LockPayable(writeCtx, sqlc.LockPayableParams{
		ID: key.SettlementID, OrganizationID: key.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Payable{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Payable{}, fmt.Errorf("lock settlement %s: %w", key.SettlementID, err)
	}
	payable := domain.Payable{
		SettlementID: row.ID, Number: row.Number, PartnershipID: row.PartnershipID,
		ConsignorID: row.ConsignorOrgID, StoreID: row.StoreOrgID, StoreName: row.StoreName,
		Status: domain.Status(row.Status), Remaining: money.Rupiah(row.ConsignorShareAmount),
		HasOpenPayment: row.HasOpenPayment, Methods: row.PaymentMethods,
	}
	if key.ProofFileID == "" {
		return payable, nil
	}
	sha, err := queries.GetProofFile(writeCtx, sqlc.GetProofFileParams{
		ID: key.ProofFileID, OrganizationID: key.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return payable, nil
	}
	if err != nil {
		return domain.Payable{}, fmt.Errorf("find proof %s: %w", key.ProofFileID, err)
	}
	payable.Proof = domain.Proof{FileID: key.ProofFileID, SHA256: sha}
	return payable, nil
}

func (r *Settlements) CreatePayment(ctx context.Context, payment domain.Payment) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	inserted, err := sqlc.New(r.pool.Querier(ctx)).CreatePayment(writeCtx, sqlc.CreatePaymentParams{
		ID: payment.ID, SettlementID: payment.SettlementID, PartnershipID: payment.PartnershipID,
		Amount: payment.Amount.Int64(), Method: string(payment.Method), PaidOn: payment.PaidOn.Format(dateLayout),
		ProofFileID: payment.Proof.FileID, ProofSha256: payment.Proof.SHA256, Status: string(payment.Status),
		RecordedBy: payment.RecordedBy, RecordedByOrgID: payment.RecordedByOrg, RecordedAt: payment.RecordedAt,
		IdempotencyKey: payment.IdempotencyKey, RequestFingerprint: payment.Fingerprint,
	})
	if err != nil {
		return fmt.Errorf("create payment %s: %w", payment.ID, err)
	}
	if inserted == 0 {
		return domain.ErrDuplicatePayment
	}
	return nil
}

func (r *Settlements) LockPayment(ctx context.Context, key domain.PaymentKey) (domain.Payment, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).LockPayment(writeCtx, sqlc.LockPaymentParams{
		ID: key.PaymentID, OrganizationID: key.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Payment{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Payment{}, fmt.Errorf("lock payment %s: %w", key.PaymentID, err)
	}
	paidOn, err := domain.ParseDueDate(row.PaidOn)
	if err != nil {
		return domain.Payment{}, fmt.Errorf("payment %s: %w", row.ID, err)
	}
	return domain.Payment{
		ID: row.ID, SettlementID: row.SettlementID, SettlementNumber: row.SettlementNumber,
		PartnershipID: row.PartnershipID, ConsignorID: row.ConsignorOrgID, StoreID: row.StoreOrgID,
		StoreName: row.StoreName, Amount: money.Rupiah(row.Amount), Method: domain.Method(row.Method),
		PaidOn: paidOn, Proof: proofOf(row.ProofFileID, row.ProofSha256), Status: domain.PaymentStatus(row.Status),
		RecordedBy: row.RecordedBy, RecordedByOrg: row.RecordedByOrgID, RecordedAt: row.CreatedAt,
	}, nil
}

func proofOf(fileID string, sha []byte) domain.Proof {
	if fileID == "" {
		return domain.Proof{}
	}
	return domain.Proof{FileID: fileID, SHA256: sha}
}

func (r *Settlements) SaveReview(ctx context.Context, review domain.Review) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	changed, err := sqlc.New(r.pool.Querier(ctx)).ReviewPayment(writeCtx, sqlc.ReviewPaymentParams{
		Status: string(review.Status), RejectionReason: review.Reason, ReviewedBy: review.By,
		ReviewedAt: review.At, ID: review.PaymentID,
	})
	if err != nil {
		return fmt.Errorf("review payment %s: %w", review.PaymentID, err)
	}
	if changed == 0 {
		return domain.ErrPaymentDecided
	}
	return nil
}

func (r *Settlements) MarkPaid(ctx context.Context, settlementID string, paidAt time.Time) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	changed, err := sqlc.New(r.pool.Querier(ctx)).MarkSettlementPaid(writeCtx, sqlc.MarkSettlementPaidParams{
		ID: settlementID, PaidAt: paidAt,
	})
	if err != nil {
		return fmt.Errorf("mark settlement %s paid: %w", settlementID, err)
	}
	if changed == 0 {
		return apperror.New(apperror.CodeSettlementPaid)
	}
	return nil
}

func withPayments(ctx context.Context, queries *sqlc.Queries, view domain.View) (domain.View, error) {
	rows, err := queries.ListSettlementPayments(ctx, view.ID)
	if err != nil {
		return domain.View{}, fmt.Errorf("list payments of settlement %s: %w", view.ID, err)
	}
	for _, row := range rows {
		paidOn, err := domain.ParseDueDate(row.PaidOn)
		if err != nil {
			return domain.View{}, fmt.Errorf("payment %s: %w", row.ID, err)
		}
		view.Payments = append(view.Payments, domain.PaymentView{
			Payment: domain.Payment{
				ID: row.ID, SettlementID: view.ID, SettlementNumber: view.Number, PartnershipID: view.PartnershipID,
				ConsignorID: view.ConsignorID, StoreID: view.StoreID, Amount: money.Rupiah(row.Amount),
				Method: domain.Method(row.Method), PaidOn: paidOn, Proof: proofOf(row.ProofFileID, row.ProofSha256),
				Status: domain.PaymentStatus(row.Status), RejectionReason: row.RejectionReason,
				RecordedBy: row.RecordedBy, RecordedByOrg: row.RecordedByOrgID, RecordedAt: row.CreatedAt,
			},
			RecordedByName: row.RecordedByName, ReviewedByName: row.ReviewedByName, ReviewedAt: row.ReviewedAt,
		})
	}
	return view, nil
}
