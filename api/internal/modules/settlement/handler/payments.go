package handler

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	recordPaymentPath    = settlementPath + "/payments"
	markPaidPath         = settlementPath + "/mark-paid"
	confirmPaymentPath   = "/payments/:id/confirm"
	rejectPaymentPath    = "/payments/:id/reject"
	idempotencyKeyHeader = "Idempotency-Key"
	replayHeader         = "Idempotent-Replay"
)

var uuidFormat = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type PaymentRecorder interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.RecordPaymentCommand) (service.PaymentReceipt, error)
}

type PaymentRejecter interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.RejectPaymentCommand) (domain.SealedView, error)
}

type PaymentDeps struct {
	Record       PaymentRecorder
	Confirm      Reader
	Reject       PaymentRejecter
	MarkPaid     Reader
	WriteLimiter httpserver.RateLimiter
}

func mountPayments(api fiber.Router, deps Deps) {
	writeLimit := httpserver.RateLimit(deps.Payments.WriteLimiter, httpserver.BySignedInUser, deps.SecurityLog)
	api.Post(recordPaymentPath, deps.Authenticate, writeLimit, recordPayment(deps.Payments.Record, deps.SecurityLog))
	api.Post(markPaidPath, deps.Authenticate, writeLimit, byID(deps.Payments.MarkPaid, deps.SecurityLog))
	api.Post(confirmPaymentPath, deps.Authenticate, writeLimit, byID(deps.Payments.Confirm, deps.SecurityLog))
	api.Post(rejectPaymentPath, deps.Authenticate, writeLimit, rejectPayment(deps.Payments.Reject, deps.SecurityLog))
}

type paymentRequest struct {
	Amount      int64   `json:"amount"`
	Method      string  `json:"method"`
	PaidOn      string  `json:"paid_on"`
	ProofFileID *string `json:"proof_file_id"`
}

func (r paymentRequest) toInput() (domain.PaymentInput, error) {
	input := domain.PaymentInput{Amount: r.Amount, Method: r.Method, PaidOn: r.PaidOn}
	if r.ProofFileID == nil {
		return input, nil
	}
	if !uuidFormat.MatchString(*r.ProofFileID) {
		return domain.PaymentInput{}, apperror.Validation(
			apperror.InvalidField("proof_file_id", apperror.FieldUnknownChoice))
	}
	input.ProofFileID = *r.ProofFileID
	return input, nil
}

func recordPayment(recorder PaymentRecorder, securityLog *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		settlementID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read settlement id: %w", err)
		}
		key := c.Get(idempotencyKeyHeader)
		if !uuidFormat.MatchString(key) {
			return apperror.New(apperror.CodeMalformedRequest)
		}
		var request paymentRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read payment: %w", err)
		}
		input, err := request.toInput()
		if err != nil {
			return err
		}
		receipt, err := recorder.Execute(c.Context(), signedIn, service.RecordPaymentCommand{
			SettlementID: settlementID, IdempotencyKey: key, Input: input,
		})
		if err != nil {
			return fmt.Errorf("record payment: %w", err)
		}
		if receipt.IsReplay {
			c.Set(replayHeader, "true")
		}
		reportInvalidSeal(c, securityLog, receipt.View)
		c.Location(httpserver.APIPrefix + settlementsPath + "/" + settlementID)
		return httpserver.Respond(c, httpserver.Reply{
			Status: fiber.StatusCreated, Message: "Pembayaran tercatat.", Data: toSettlementResponse(receipt.View),
		})
	}
}

type rejectRequest struct {
	Reason string `json:"reason"`
}

func rejectPayment(rejecter PaymentRejecter, securityLog *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		paymentID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read payment id: %w", err)
		}
		var request rejectRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read rejection: %w", err)
		}
		reason := httpserver.MultilineInput("reason", request.Reason)
		if err := httpserver.RejectUncleanInputs(reason); err != nil {
			return fmt.Errorf("check rejection reason: %w", err)
		}
		sealed, err := rejecter.Execute(c.Context(), signedIn, service.RejectPaymentCommand{
			PaymentID: paymentID, Reason: reason.Value,
		})
		if err != nil {
			return fmt.Errorf("reject payment: %w", err)
		}
		reportInvalidSeal(c, securityLog, sealed)
		return httpserver.Respond(c, httpserver.Reply{Data: toSettlementResponse(sealed)})
	}
}
