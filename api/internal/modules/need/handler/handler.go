package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	needsPath   = "/needs"
	needPath    = needsPath + "/:id"
	extendPath  = needPath + "/extend"
	fulfillPath = needPath + "/fulfill"
	closePath   = needPath + "/close"
)

type Poster interface {
	Execute(ctx context.Context, signedIn actor.Actor, form domain.Form) (domain.Need, error)
}

type Changer interface {
	Execute(ctx context.Context, signedIn actor.Actor, needID string) (domain.Need, error)
}

type Reader interface {
	Execute(ctx context.Context, signedIn actor.Actor, needID string) (domain.Detail, error)
}

type Lister interface {
	Execute(ctx context.Context, signedIn actor.Actor, request service.ListRequest) (service.Page, error)
}

type Deps struct {
	Authenticate fiber.Handler
	Post         Poster
	Get          Reader
	List         Lister
	Extend       Changer
	Fulfill      Changer
	Close        Changer
	WriteLimiter httpserver.RateLimiter
	SecurityLog  *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	writeLimit := httpserver.RateLimit(deps.WriteLimiter, httpserver.BySignedInUser, deps.SecurityLog)
	api.Post(needsPath, deps.Authenticate, writeLimit, post(deps.Post))
	api.Get(needsPath, deps.Authenticate, list(deps.List))
	api.Get(needPath, deps.Authenticate, get(deps.Get))
	api.Post(extendPath, deps.Authenticate, writeLimit, change(deps.Extend))
	api.Post(fulfillPath, deps.Authenticate, writeLimit, change(deps.Fulfill))
	api.Post(closePath, deps.Authenticate, writeLimit, change(deps.Close))
}

func signedInActor(c fiber.Ctx) (actor.Actor, error) {
	signedIn, ok := httpserver.ActorFrom(c)
	if !ok {
		return actor.Actor{}, apperror.New(apperror.CodeUnauthenticated)
	}
	return signedIn, nil
}

func post(poster Poster) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		var request needRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read need: %w", err)
		}
		form, err := request.toForm()
		if err != nil {
			return err
		}
		need, err := poster.Execute(c.Context(), signedIn, form)
		if err != nil {
			return fmt.Errorf("post need: %w", err)
		}
		c.Location(httpserver.APIPrefix + needsPath + "/" + need.ID)
		return httpserver.Respond(c, httpserver.Reply{Status: fiber.StatusCreated, Data: toNeedResponse(need)})
	}
}

func get(reader Reader) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		needID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read need id: %w", err)
		}
		detail, err := reader.Execute(c.Context(), signedIn, needID)
		if err != nil {
			return fmt.Errorf("get need: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toDetailResponse(detail)})
	}
}

func change(changer Changer) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		needID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read need id: %w", err)
		}
		need, err := changer.Execute(c.Context(), signedIn, needID)
		if err != nil {
			return fmt.Errorf("change need: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toNeedResponse(need)})
	}
}

func list(lister Lister) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		page, err := lister.Execute(c.Context(), signedIn, service.ListRequest{
			Status: c.Query("status"), Cursor: c.Query("cursor"), Limit: c.Query("limit"),
		})
		if err != nil {
			return fmt.Errorf("list needs: %w", err)
		}
		items := make([]needResponse, 0, len(page.Items))
		for _, need := range page.Items {
			items = append(items, toNeedResponse(need))
		}
		return httpserver.RespondList(c, httpserver.Page[needResponse]{
			Items: items, NextCursor: page.NextCursor, HasMore: page.HasMore,
		})
	}
}
