package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
)

type RecordPaymentCommand struct {
	SettlementID   string
	IdempotencyKey string
	Input          domain.PaymentInput
}

type PaymentReceipt struct {
	View     domain.SealedView
	IsReplay bool
}

type RecordPayment struct {
	deps   Deps
	reader *GetSettlement
}

func NewRecordPayment(deps Deps) *RecordPayment {
	return &RecordPayment{deps: deps, reader: NewGetSettlement(deps)}
}

func (uc *RecordPayment) Execute(
	ctx context.Context, signedIn actor.Actor, cmd RecordPaymentCommand,
) (PaymentReceipt, error) {
	if err := policy.Require(signedIn, policy.RecordPayments); err != nil {
		return PaymentReceipt{}, fmt.Errorf("check payment recorder: %w", err)
	}
	fingerprint, err := cmd.Input.Fingerprint()
	if err != nil {
		return PaymentReceipt{}, fmt.Errorf("fingerprint payment request: %w", err)
	}
	key := domain.PaymentKeyRef{OrganizationID: signedIn.Membership.OrganizationID, Key: cmd.IdempotencyKey}
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		return uc.record(ctx, signedIn, cmd, fingerprint)
	})
	if errors.Is(err, domain.ErrDuplicatePayment) {
		return uc.replay(ctx, signedIn, key, fingerprint)
	}
	if err != nil {
		return PaymentReceipt{}, fmt.Errorf("record payment for %s: %w", cmd.SettlementID, err)
	}
	view, err := uc.reader.read(ctx, signedIn, cmd.SettlementID)
	return PaymentReceipt{View: view}, err
}

func (uc *RecordPayment) record(
	ctx context.Context, signedIn actor.Actor, cmd RecordPaymentCommand, fingerprint []byte,
) error {
	storeID := signedIn.Membership.OrganizationID
	_, isStored, err := uc.deps.Settlements.FindPaymentByKey(ctx, domain.PaymentKeyRef{
		OrganizationID: storeID, Key: cmd.IdempotencyKey,
	})
	if err != nil {
		return fmt.Errorf("find payment by key: %w", err)
	}
	if isStored {
		return domain.ErrDuplicatePayment
	}
	payable, err := uc.deps.Settlements.LockPayable(ctx, domain.PayableKey{
		OrganizationID: storeID, SettlementID: cmd.SettlementID, ProofFileID: cmd.Input.ProofFileID,
	})
	if err != nil {
		return fmt.Errorf("lock settlement: %w", err)
	}
	if payable.StoreID != storeID {
		return apperror.New(apperror.CodeForbidden)
	}
	now := uc.deps.Clock.Now()
	draft, err := payable.CheckPayment(cmd.Input, now)
	if err != nil {
		return fmt.Errorf("check payment: %w", err)
	}
	id, err := uc.deps.IDs.NewID()
	if err != nil {
		return fmt.Errorf("generate payment id: %w", err)
	}
	payment := draft.Record(domain.PaymentRecording{
		ID: id, RecordedBy: signedIn.UserID, IdempotencyKey: cmd.IdempotencyKey, Fingerprint: fingerprint, At: now,
	})
	return uc.save(ctx, payment)
}

func (uc *RecordPayment) save(ctx context.Context, payment domain.Payment) error {
	if err := uc.deps.Settlements.CreatePayment(ctx, payment); err != nil {
		return fmt.Errorf("save payment: %w", err)
	}
	if err := uc.deps.Sealer.Seal(ctx, payment.SealRequest()); err != nil {
		return fmt.Errorf("seal payment: %w", err)
	}
	if err := uc.deps.Announcer.Announce(ctx, payment.RecordedNotice()); err != nil {
		return fmt.Errorf("announce payment: %w", err)
	}
	return recordPaymentAudit(ctx, uc.deps.Audit, payment, audit.PaymentRecorded)
}

func recordPaymentAudit(ctx context.Context, trail AuditTrail, payment domain.Payment, action audit.Action) error {
	err := trail.Record(ctx, audit.Entry{
		OrganizationID: payment.RecordedByOrg, ActorID: payment.RecordedBy, Action: action,
		EntityType: audit.PaymentEntity, EntityID: payment.ID, At: payment.RecordedAt,
	})
	if err != nil {
		return fmt.Errorf("audit payment %s: %w", payment.ID, err)
	}
	return nil
}

func (uc *RecordPayment) replay(
	ctx context.Context, signedIn actor.Actor, key domain.PaymentKeyRef, fingerprint []byte,
) (PaymentReceipt, error) {
	stored, isStored, err := uc.deps.Settlements.FindPaymentByKey(ctx, key)
	if err != nil {
		return PaymentReceipt{}, fmt.Errorf("find replayed payment: %w", err)
	}
	if !isStored {
		return PaymentReceipt{}, fmt.Errorf("find replayed payment: %w", domain.ErrDuplicatePayment)
	}
	if !bytes.Equal(stored.Fingerprint, fingerprint) {
		return PaymentReceipt{}, apperror.New(apperror.CodeIdempotencyKeyReused)
	}
	view, err := uc.reader.read(ctx, signedIn, stored.SettlementID)
	return PaymentReceipt{View: view, IsReplay: true}, err
}
