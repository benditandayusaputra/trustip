package handler

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
)

type reviewResponse struct {
	ByName string `json:"by_name"`
	At     string `json:"at"`
}

type paymentResponse struct {
	ID              string          `json:"id"`
	Amount          int64           `json:"amount"`
	Method          string          `json:"method"`
	PaidOn          string          `json:"paid_on"`
	ProofFileID     *string         `json:"proof_file_id"`
	Status          string          `json:"status"`
	RejectionReason *string         `json:"rejection_reason"`
	RecordedByName  string          `json:"recorded_by_name"`
	RecordedAt      string          `json:"recorded_at"`
	Review          *reviewResponse `json:"review"`
	Seal            sealResponse    `json:"seal"`
}

type payeeResponse struct {
	PaymentInfo       string   `json:"payment_info"`
	PaymentQRFileID   *string  `json:"payment_qr_file_id"`
	Methods           []string `json:"methods"`
	BankName          string   `json:"bank_name"`
	BankAccountNumber string   `json:"bank_account_number"`
	BankAccountHolder string   `json:"bank_account_holder"`
}

func toPayments(sealed domain.SealedView) []paymentResponse {
	payments := make([]paymentResponse, 0, len(sealed.Payments))
	for _, payment := range sealed.Payments {
		response := paymentResponse{
			ID: payment.ID, Amount: payment.Amount.Int64(), Method: string(payment.Method),
			PaidOn: payment.PaidOn.Format(time.DateOnly), ProofFileID: optional(payment.Proof.FileID),
			Status: string(payment.Status), RejectionReason: optional(payment.RejectionReason),
			RecordedByName: payment.RecordedByName, RecordedAt: timestamp(payment.RecordedAt),
			Seal: toSealResponse(sealed.PaymentSeals[payment.ID]),
		}
		if payment.Status != domain.PendingConfirmation {
			response.Review = &reviewResponse{ByName: payment.ReviewedByName, At: timestamp(payment.ReviewedAt)}
		}
		payments = append(payments, response)
	}
	return payments
}

func toPayee(payee domain.Payee) *payeeResponse {
	if !payee.IsShown {
		return nil
	}
	return &payeeResponse{
		PaymentInfo: payee.Info, PaymentQRFileID: optional(payee.QRFileID), Methods: payee.Methods,
		BankName: payee.Bank.Name, BankAccountNumber: payee.Bank.Number, BankAccountHolder: payee.Bank.Holder,
	}
}

func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
