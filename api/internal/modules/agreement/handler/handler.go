package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const agreementPath = "/partnerships/:id/agreement"

type Reader interface {
	Execute(ctx context.Context, signedIn actor.Actor, partnershipID string) (domain.SealedView, error)
}

type Deps struct {
	Authenticate fiber.Handler
	Get          Reader
	SecurityLog  *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	api.Get(agreementPath, deps.Authenticate, read(deps.Get, deps.SecurityLog))
}

func read(reader Reader, securityLog *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, ok := httpserver.ActorFrom(c)
		if !ok {
			return apperror.New(apperror.CodeUnauthenticated)
		}
		partnershipID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read partnership id: %w", err)
		}
		sealed, err := reader.Execute(c.Context(), signedIn, partnershipID)
		if err != nil {
			return fmt.Errorf("read agreement: %w", err)
		}
		if !sealed.Seal.IsValid {
			securityLog.Warn("seal invalid", "event", "seal_invalid", "request_id", httpserver.RequestIDFrom(c),
				"entity_type", domain.EntityType, "entity_id", sealed.View.Agreement.ID, "user_id", signedIn.UserID,
				"ip", httpserver.ClientIP(c))
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toResponse(sealed)})
	}
}
