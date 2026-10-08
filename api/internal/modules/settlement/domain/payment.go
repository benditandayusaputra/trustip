package domain

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/canonical"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
)

const PaymentEntityType = "PAYMENT"

type Method string

const (
	Transfer Method = "TRANSFER"
	QRIS     Method = "QRIS"
	Cash     Method = "CASH"
)

func Methods() []string {
	return []string{string(Transfer), string(QRIS), string(Cash)}
}

func (m Method) needsProof() bool {
	return m == Transfer || m == QRIS
}

type PaymentStatus string

const (
	PendingConfirmation PaymentStatus = "PENDING_CONFIRMATION"
	PaymentConfirmed    PaymentStatus = "CONFIRMED"
	PaymentRejected     PaymentStatus = "REJECTED"
)

type Proof struct {
	FileID string
	SHA256 []byte
}

func (p Proof) isGiven() bool {
	return p.FileID != ""
}

type PaymentInput struct {
	Amount      int64
	Method      string
	PaidOn      string
	ProofFileID string
}

func (i PaymentInput) Fingerprint() ([]byte, error) {
	digest, err := canonical.Digest(map[string]any{
		"amount": i.Amount, "method": i.Method, "paid_on": i.PaidOn, "proof_file_id": i.ProofFileID,
	})
	if err != nil {
		return nil, fmt.Errorf("fingerprint payment request: %w", err)
	}
	return digest, nil
}

type Payable struct {
	SettlementID   string
	Number         string
	PartnershipID  string
	ConsignorID    string
	StoreID        string
	StoreName      string
	Status         Status
	Remaining      money.Rupiah
	HasOpenPayment bool
	Proof          Proof
}

type PaymentDraft struct {
	payable Payable
	Amount  money.Rupiah
	Method  Method
	PaidOn  time.Time
	Proof   Proof
}

func (p Payable) CheckPayment(input PaymentInput, now time.Time) (PaymentDraft, error) {
	if err := p.checkOpen(); err != nil {
		return PaymentDraft{}, err
	}
	paidOn, dateProblem := parsePaidOn(input.PaidOn, now)
	problems := field.Collect(
		p.findAmountProblem(input.Amount), findMethodProblem(input.Method), dateProblem,
		p.findProofProblem(input),
	)
	if len(problems) > 0 {
		return PaymentDraft{}, apperror.Validation(problems...)
	}
	draft := PaymentDraft{payable: p, Amount: p.Remaining, Method: Method(input.Method), PaidOn: paidOn}
	if input.ProofFileID != "" {
		draft.Proof = p.Proof
	}
	return draft, nil
}

func (p Payable) checkOpen() error {
	if p.Status == Paid {
		return apperror.New(apperror.CodeSettlementPaid)
	}
	if p.HasOpenPayment {
		return apperror.New(apperror.CodePaymentPending)
	}
	return nil
}

func (p Payable) findAmountProblem(amount int64) field.Problem {
	if amount != p.Remaining.Int64() {
		return field.Invalid("amount", apperror.FieldAmountMismatch,
			apperror.Param{Name: "Rp", Value: p.Remaining.Text()})
	}
	return field.Problem{}
}

func findMethodProblem(method string) field.Problem {
	if method == "" {
		return field.Invalid("method", apperror.FieldRequired)
	}
	return field.CheckChoice("method", method, Methods())
}

func parsePaidOn(raw string, now time.Time) (time.Time, field.Problem) {
	paidOn, err := ParseDueDate(raw)
	if err != nil {
		return time.Time{}, field.Invalid("paid_on", apperror.FieldInvalidDate)
	}
	if paidOn.After(startOfDay(now)) {
		return time.Time{}, field.Invalid("paid_on", apperror.FieldOutOfRange)
	}
	return paidOn, field.Problem{}
}

func (p Payable) findProofProblem(input PaymentInput) field.Problem {
	if input.ProofFileID == "" {
		if Method(input.Method).needsProof() {
			return field.Invalid("proof_file_id", apperror.FieldProofRequired)
		}
		return field.Problem{}
	}
	if !p.Proof.isGiven() || p.Proof.FileID != input.ProofFileID {
		return field.Invalid("proof_file_id", apperror.FieldUnknownChoice)
	}
	return field.Problem{}
}

type PaymentRecording struct {
	ID             string
	RecordedBy     string
	IdempotencyKey string
	Fingerprint    []byte
	At             time.Time
}

type Payment struct {
	ID               string
	SettlementID     string
	SettlementNumber string
	PartnershipID    string
	ConsignorID      string
	StoreID          string
	StoreName        string
	Amount           money.Rupiah
	Method           Method
	PaidOn           time.Time
	Proof            Proof
	Status           PaymentStatus
	RejectionReason  string
	RecordedBy       string
	RecordedByOrg    string
	IdempotencyKey   string
	Fingerprint      []byte
	RecordedAt       time.Time
}

func (d PaymentDraft) Record(recording PaymentRecording) Payment {
	payable := d.payable
	return Payment{
		ID: recording.ID, SettlementID: payable.SettlementID, SettlementNumber: payable.Number,
		PartnershipID: payable.PartnershipID, ConsignorID: payable.ConsignorID, StoreID: payable.StoreID,
		StoreName: payable.StoreName, Amount: d.Amount, Method: d.Method, PaidOn: d.PaidOn, Proof: d.Proof,
		Status: PendingConfirmation, RecordedBy: recording.RecordedBy, RecordedByOrg: payable.StoreID,
		IdempotencyKey: recording.IdempotencyKey, Fingerprint: recording.Fingerprint, RecordedAt: recording.At,
	}
}

func (p Payable) SettleDirectly(consignor Decider, recording PaymentRecording) (Payment, error) {
	if consignor.OrganizationID != p.ConsignorID {
		return Payment{}, apperror.New(apperror.CodeForbidden)
	}
	if err := p.checkOpen(); err != nil {
		return Payment{}, err
	}
	draft := PaymentDraft{payable: p, Amount: p.Remaining, Method: Cash, PaidOn: startOfDay(recording.At)}
	recording.RecordedBy = consignor.UserID
	payment := draft.Record(recording)
	payment.Status = PaymentConfirmed
	payment.RecordedByOrg = consignor.OrganizationID
	return payment, nil
}

func (p Payment) Content() map[string]any {
	return map[string]any{
		"payment_id": p.ID, "settlement_id": p.SettlementID, "settlement_number": p.SettlementNumber,
		"amount": p.Amount.Int64(), "method": string(p.Method), "paid_on": p.PaidOn.Format(dateLayout),
		"proof_sha256": hex.EncodeToString(p.Proof.SHA256), "recorded_by": p.RecordedBy,
		"recorded_by_org_id": p.RecordedByOrg,
	}
}

func (p Payment) SealRequest() SealRequest {
	event := "PAYMENT_RECORDED"
	if p.Status == PaymentConfirmed {
		event = "PAYMENT_CONFIRMED"
	}
	return SealRequest{
		PartnershipID: p.PartnershipID, Event: event, DocumentType: PaymentEntityType, DocumentID: p.ID,
		DocumentNumber: p.SettlementNumber, Content: p.Content(), ActorUserID: p.RecordedBy,
		ActorOrgID: p.RecordedByOrg, At: p.RecordedAt,
	}
}

func (p Payment) RecordedNotice() Notice {
	return Notice{
		OrganizationID: p.ConsignorID, ActorUserID: p.RecordedBy, Type: "PAYMENT_RECORDED",
		Title: "Pembayaran " + p.SettlementNumber, Link: "/app/setoran/" + p.SettlementID, OwnersOnly: true,
		Body: fmt.Sprintf("%s mencatat pembayaran %s untuk %s.", p.StoreName, p.Amount.Text(), p.SettlementNumber),
	}
}
