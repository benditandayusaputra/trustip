package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	settlementsPath = "/settlements"
	settlementPath  = settlementsPath + "/:id"
	runningPath     = "/partnerships/:id/running-sales"
)

type Reader interface {
	Execute(ctx context.Context, signedIn actor.Actor, settlementID string) (domain.SealedView, error)
}

type Lister interface {
	Execute(ctx context.Context, signedIn actor.Actor, request service.ListRequest) (service.Page, error)
}

type RunningReader interface {
	Execute(ctx context.Context, signedIn actor.Actor, partnershipID string) (domain.Running, error)
}

type Deps struct {
	Authenticate fiber.Handler
	Get          Reader
	List         Lister
	Running      RunningReader
	Payments     PaymentDeps
	SecurityLog  *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	api.Get(settlementsPath, deps.Authenticate, list(deps.List))
	api.Get(settlementPath, deps.Authenticate, byID(deps.Get, deps.SecurityLog))
	api.Get(runningPath, deps.Authenticate, running(deps.Running))
	mountPayments(api, deps)
}

func signedInActor(c fiber.Ctx) (actor.Actor, error) {
	signedIn, ok := httpserver.ActorFrom(c)
	if !ok {
		return actor.Actor{}, apperror.New(apperror.CodeUnauthenticated)
	}
	return signedIn, nil
}

func byID(reader Reader, securityLog *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		settlementID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read settlement id: %w", err)
		}
		sealed, err := reader.Execute(c.Context(), signedIn, settlementID)
		if err != nil {
			return fmt.Errorf("read settlement: %w", err)
		}
		reportInvalidSeal(c, securityLog, sealed)
		return httpserver.Respond(c, httpserver.Reply{Data: toSettlementResponse(sealed)})
	}
}

func list(lister Lister) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		page, err := lister.Execute(c.Context(), signedIn, service.ListRequest{
			Filter: domain.ListFilter{
				PartnershipID: c.Query("partnership_id"), Status: c.Query("status"), Age: c.Query("age"),
			},
			Cursor: c.Query("cursor"), Limit: c.Query("limit"),
		})
		if err != nil {
			return fmt.Errorf("list settlements: %w", err)
		}
		items := make([]summaryResponse, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, toSummaryResponse(item))
		}
		return httpserver.RespondList(c, httpserver.Page[summaryResponse]{
			Items: items, NextCursor: page.NextCursor, HasMore: page.HasMore,
		})
	}
}

func running(reader RunningReader) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		partnershipID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read partnership id: %w", err)
		}
		found, err := reader.Execute(c.Context(), signedIn, partnershipID)
		if err != nil {
			return fmt.Errorf("read running sales: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toRunningResponse(found)})
	}
}

func reportInvalidSeal(c fiber.Ctx, securityLog *slog.Logger, sealed domain.SealedView) {
	if sealed.Seal.IsValid {
		return
	}
	signedIn, _ := httpserver.ActorFrom(c)
	securityLog.Warn("seal invalid", "event", "seal_invalid", "request_id", httpserver.RequestIDFrom(c),
		"entity_type", domain.EntityType, "entity_id", sealed.ID, "user_id", signedIn.UserID,
		"ip", httpserver.ClientIP(c))
}
