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
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	countsPath           = "/counts"
	countPath            = countsPath + "/:id"
	sheetPath            = "/partnerships/:id/count-sheet"
	confirmPath          = countPath + "/confirm"
	idempotencyKeyHeader = "Idempotency-Key"
	replayHeader         = "Idempotent-Replay"
)

var uuidFormat = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type Recorder interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.RecordCountCommand) (service.CountReceipt, error)
}

type Reader interface {
	Execute(ctx context.Context, signedIn actor.Actor, countID string) (domain.SealedView, error)
}

type SheetReader interface {
	Execute(ctx context.Context, signedIn actor.Actor, partnershipID string) (domain.Sheet, error)
}

type Lister interface {
	Execute(ctx context.Context, signedIn actor.Actor, request service.ListRequest) (service.Page, error)
}

type Clock interface {
	Now() time.Time
}

type Deps struct {
	Authenticate fiber.Handler
	Record       Recorder
	Get          Reader
	Sheet        SheetReader
	Confirm      Reader
	List         Lister
	Clock        Clock
	WriteLimiter httpserver.RateLimiter
	SecurityLog  *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	writeLimit := httpserver.RateLimit(deps.WriteLimiter, httpserver.BySignedInUser, deps.SecurityLog)
	api.Post(countsPath, deps.Authenticate, writeLimit, record(deps))
	api.Get(countsPath, deps.Authenticate, list(deps.List))
	api.Get(countPath, deps.Authenticate, byID(deps.Get, deps.SecurityLog))
	api.Post(confirmPath, deps.Authenticate, writeLimit, byID(deps.Confirm, deps.SecurityLog))
	api.Get(sheetPath, deps.Authenticate, sheet(deps.Sheet))
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
		if !uuidFormat.MatchString(key) {
			return apperror.New(apperror.CodeMalformedRequest)
		}
		var request countRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read count: %w", err)
		}
		input, err := request.toInput(deps.Clock.Now())
		if err != nil {
			return err
		}
		receipt, err := deps.Record.Execute(c.Context(), signedIn, service.RecordCountCommand{
			IdempotencyKey: key, Input: input,
		})
		if err != nil {
			return fmt.Errorf("record count: %w", err)
		}
		if receipt.IsReplay {
			c.Set(replayHeader, "true")
		}
		reportInvalidSeal(c, deps.SecurityLog, receipt.View)
		c.Location(httpserver.APIPrefix + countsPath + "/" + receipt.View.ID)
		return httpserver.Respond(c, httpserver.Reply{
			Status: fiber.StatusCreated, Message: "Hitung Sisa tersimpan.", Data: toCountResponse(receipt.View),
		})
	}
}

func byID(reader Reader, securityLog *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		countID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read count id: %w", err)
		}
		sealed, err := reader.Execute(c.Context(), signedIn, countID)
		if err != nil {
			return fmt.Errorf("act on count: %w", err)
		}
		reportInvalidSeal(c, securityLog, sealed)
		return httpserver.Respond(c, httpserver.Reply{Data: toCountResponse(sealed)})
	}
}

func sheet(reader SheetReader) fiber.Handler {
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
			return fmt.Errorf("read count sheet: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toSheetResponse(found)})
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
				PartnershipID: c.Query("partnership_id"), From: c.Query("from"), To: c.Query("to"),
			},
			Cursor: c.Query("cursor"), Limit: c.Query("limit"),
		})
		if err != nil {
			return fmt.Errorf("list counts: %w", err)
		}
		items := make([]summaryResponse, 0, len(page.Items))
		for _, summary := range page.Items {
			items = append(items, toSummaryResponse(summary))
		}
		return httpserver.RespondList(c, httpserver.Page[summaryResponse]{
			Items: items, NextCursor: page.NextCursor, HasMore: page.HasMore,
		})
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

func timestamp(at time.Time) string {
	return at.UTC().Format(time.RFC3339)
}
