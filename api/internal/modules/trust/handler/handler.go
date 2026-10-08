package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/trust/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const reportsPath = "/listing-reports"

type Reporter interface {
	Execute(ctx context.Context, signedIn actor.Actor, form domain.ReportForm) (domain.Filed, error)
}

type Deps struct {
	Authenticate  fiber.Handler
	Report        Reporter
	ReportLimiter httpserver.RateLimiter
	SecurityLog   *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	reportLimit := httpserver.RateLimit(deps.ReportLimiter, httpserver.BySignedInUser, deps.SecurityLog)
	api.Post(reportsPath, deps.Authenticate, reportLimit, report(deps.Report))
}

type reportRequest struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Reason     string `json:"reason"`
	Note       string `json:"note"`
}

type reportResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func report(reporter Reporter) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, ok := httpserver.ActorFrom(c)
		if !ok {
			return apperror.New(apperror.CodeUnauthenticated)
		}
		var request reportRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read report: %w", err)
		}
		note := httpserver.MultilineInput("note", request.Note)
		if err := httpserver.RejectUncleanInputs(note); err != nil {
			return fmt.Errorf("check report note: %w", err)
		}
		filed, err := reporter.Execute(c.Context(), signedIn, domain.ReportForm{
			TargetType: request.TargetType, TargetID: request.TargetID, Reason: request.Reason, Note: note.Value,
		})
		if err != nil {
			return fmt.Errorf("file report: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{
			Status: fiber.StatusCreated, Data: reportResponse{ID: filed.ID, Status: filed.Status},
		})
	}
}
