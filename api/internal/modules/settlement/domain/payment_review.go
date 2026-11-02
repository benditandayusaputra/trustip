package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
)

const MaxRejectionLength = 200

type Decider struct {
	UserID         string
	OrganizationID string
}

type Review struct {
	PaymentID    string
	SettlementID string
	Status       PaymentStatus
	Reason       string
	By           string
	ByOrgID      string
	At           time.Time
	Event        string
}

func (p Payment) Confirm(consignor Decider, at time.Time) (Review, error) {
	if err := p.checkReviewer(consignor); err != nil {
		return Review{}, err
	}
	return p.reviewOf(consignor, at, PaymentConfirmed), nil
}

func (p Payment) Reject(consignor Decider, reason string, rejectedAt time.Time) (Review, error) {
	if err := p.checkReviewer(consignor); err != nil {
		return Review{}, err
	}
	problems := field.Collect(findReasonProblem(reason))
	if len(problems) > 0 {
		return Review{}, apperror.Validation(problems...)
	}
	review := p.reviewOf(consignor, rejectedAt, PaymentRejected)
	review.Reason = reason
	return review, nil
}

func findReasonProblem(reason string) field.Problem {
	if reason == "" {
		return field.Invalid("reason", apperror.FieldRequired)
	}
	return field.CheckLength("reason", reason, field.Length{Max: MaxRejectionLength})
}

func (p Payment) checkReviewer(consignor Decider) error {
	if consignor.OrganizationID != p.ConsignorID {
		return apperror.New(apperror.CodeForbidden)
	}
	if p.Status != PendingConfirmation {
		return apperror.New(apperror.CodeDocumentNotPending)
	}
	return nil
}

func (p Payment) reviewOf(consignor Decider, at time.Time, status PaymentStatus) Review {
	return Review{
		PaymentID: p.ID, SettlementID: p.SettlementID, Status: status, By: consignor.UserID,
		ByOrgID: consignor.OrganizationID, At: at, Event: "PAYMENT_" + string(status),
	}
}

func (p Payment) ReviewSealRequest(review Review) SealRequest {
	request := p.SealRequest()
	request.Event = review.Event
	request.ActorUserID = review.By
	request.ActorOrgID = review.ByOrgID
	request.At = review.At
	return request
}

func (p Payment) ReviewNotice(review Review) Notice {
	body := fmt.Sprintf("Pembayaran %s untuk %s diterima.", p.Amount.Text(), p.SettlementNumber)
	if review.Status == PaymentRejected {
		body = fmt.Sprintf("Pembayaran %s untuk %s ditolak: %s.", p.Amount.Text(), p.SettlementNumber, review.Reason)
	}
	return Notice{
		OrganizationID: p.StoreID, ActorUserID: review.By, Type: review.Event,
		Title: "Pembayaran " + p.SettlementNumber, Body: body, Link: "/app/setoran/" + p.SettlementID, OwnersOnly: true,
	}
}

var (
	ErrDuplicatePayment = errors.New("a payment with this idempotency key already exists")
	ErrPaymentDecided   = errors.New("the payment is no longer waiting for a response")
)

type PayableKey struct {
	OrganizationID string
	SettlementID   string
	ProofFileID    string
}

type PaymentKey struct {
	OrganizationID string
	PaymentID      string
}

type PaymentKeyRef struct {
	OrganizationID string
	Key            string
}

type StoredPayment struct {
	PaymentID    string
	SettlementID string
	Fingerprint  []byte
}

type Approval struct {
	SettlementID string
	By           string
	At           time.Time
}

func (r Review) Approval() Approval {
	return Approval{SettlementID: r.SettlementID, By: r.By, At: r.At}
}

func (p Payment) SealCheck() SealCheck {
	return SealCheck{DocumentType: PaymentEntityType, DocumentID: p.ID, Content: p.Content()}
}

type PaymentView struct {
	Payment
	RecordedByName string
	ReviewedByName string
	ReviewedAt     time.Time
}

type Payee struct {
	IsShown  bool
	Info     string
	QRFileID string
	Methods  []string
	Bank     BankAccount
}

type BankAccount struct {
	Name   string
	Number string
	Holder string
}
