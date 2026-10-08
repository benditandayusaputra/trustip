package handler

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	dropsPath            = "/drops"
	dropPath             = dropsPath + "/:id"
	confirmPath          = dropPath + "/confirm"
	cancelPath           = dropPath + "/cancel"
	idempotencyKeyHeader = "Idempotency-Key"
	replayHeader         = "Idempotent-Replay"
)

var uuidFormat = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func isUUID(value string) bool {
	return uuidFormat.MatchString(value)
}

type Recorder interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.RecordDropCommand) (service.DropReceipt, error)
}

type Reader interface {
	Execute(ctx context.Context, signedIn actor.Actor, dropID string) (domain.SealedView, error)
}

type Clock interface {
	Now() time.Time
}

type Deps struct {
	Authenticate fiber.Handler
	Record       Recorder
	Get          Reader
	Confirm      Reader
	Cancel       Reader
	Clock        Clock
	WriteLimiter httpserver.RateLimiter
	SecurityLog  *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	writeLimit := httpserver.RateLimit(deps.WriteLimiter, httpserver.BySignedInUser, deps.SecurityLog)
	api.Post(dropsPath, deps.Authenticate, writeLimit, record(deps))
	api.Get(dropPath, deps.Authenticate, byID(deps.Get, deps.SecurityLog))
	api.Post(confirmPath, deps.Authenticate, writeLimit, byID(deps.Confirm, deps.SecurityLog))
	api.Post(cancelPath, deps.Authenticate, writeLimit, byID(deps.Cancel, deps.SecurityLog))
}

func signedInActor(c fiber.Ctx) (actor.Actor, error) {
	signedIn, ok := httpserver.ActorFrom(c)
	if !ok {
		return actor.Actor{}, apperror.New(apperror.CodeUnauthenticated)
	}
	return signedIn, nil
}

func record(deps Deps) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		key := c.Get(idempotencyKeyHeader)
		if !isUUID(key) {
			return apperror.New(apperror.CodeMalformedRequest)
		}
		var request dropRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read drop: %w", err)
		}
		input, err := request.toInput(deps.Clock.Now())
		if err != nil {
			return err
		}
		receipt, err := deps.Record.Execute(c.Context(), signedIn, service.RecordDropCommand{
			IdempotencyKey: key, Input: input,
		})
		if err != nil {
			return fmt.Errorf("record drop: %w", err)
		}
		if receipt.IsReplay {
			c.Set(replayHeader, "true")
		}
		reportInvalidSeal(c, deps.SecurityLog, receipt.View)
		c.Location(httpserver.APIPrefix + dropsPath + "/" + receipt.View.ID)
		return httpserver.Respond(c, httpserver.Reply{
			Status: fiber.StatusCreated, Message: "Titipan masuk tersimpan.", Data: toDropResponse(receipt.View),
		})
	}
}

func byID(reader Reader, securityLog *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		dropID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read drop id: %w", err)
		}
		sealed, err := reader.Execute(c.Context(), signedIn, dropID)
		if err != nil {
			return fmt.Errorf("act on drop: %w", err)
		}
		reportInvalidSeal(c, securityLog, sealed)
		return httpserver.Respond(c, httpserver.Reply{Data: toDropResponse(sealed)})
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
