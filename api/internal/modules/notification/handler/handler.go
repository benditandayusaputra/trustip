package handler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	notificationsPath = "/notifications"
	unreadCountPath   = notificationsPath + "/unread-count"
	readPath          = notificationsPath + "/:id/read"
	readAllPath       = notificationsPath + "/read-all"
)

type Lister interface {
	Execute(ctx context.Context, signedIn actor.Actor, request service.PageRequest) (service.Page, error)
}

type Counter interface {
	Execute(ctx context.Context, signedIn actor.Actor) (int, error)
}

type Marker interface {
	Execute(ctx context.Context, signedIn actor.Actor, notificationID string) error
}

type AllMarker interface {
	Execute(ctx context.Context, signedIn actor.Actor) error
}

type Deps struct {
	Authenticate fiber.Handler
	List         Lister
	Count        Counter
	MarkRead     Marker
	MarkAllRead  AllMarker
	WriteLimiter httpserver.RateLimiter
	SecurityLog  *slog.Logger
}

type notificationResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Link      string `json:"link"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

type unreadCountResponse struct {
	Count int `json:"count"`
}

func Mount(api fiber.Router, deps Deps) {
	writeLimit := httpserver.RateLimit(deps.WriteLimiter, httpserver.BySignedInUser, deps.SecurityLog)
	api.Get(notificationsPath, deps.Authenticate, list(deps.List))
	api.Get(unreadCountPath, deps.Authenticate, count(deps.Count))
	api.Post(readAllPath, deps.Authenticate, writeLimit, markAll(deps.MarkAllRead))
	api.Post(readPath, deps.Authenticate, writeLimit, mark(deps.MarkRead))
}

func signedInActor(c fiber.Ctx) (actor.Actor, error) {
	signedIn, ok := httpserver.ActorFrom(c)
	if !ok {
		return actor.Actor{}, apperror.New(apperror.CodeUnauthenticated)
	}
	return signedIn, nil
}

func list(lister Lister) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		page, err := lister.Execute(c.Context(), signedIn,
			service.PageRequest{Cursor: c.Query("cursor"), Limit: c.Query("limit")})
		if err != nil {
			return fmt.Errorf("list notifications: %w", err)
		}
		items := make([]notificationResponse, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, toResponse(item))
		}
		return httpserver.RespondList(c, httpserver.Page[notificationResponse]{
			Items: items, NextCursor: page.NextCursor, HasMore: page.HasMore,
		})
	}
}

func toResponse(item domain.Notification) notificationResponse {
	return notificationResponse{
		ID: item.ID, Type: item.Type, Title: item.Title, Body: item.Body, Link: item.Link, IsRead: item.IsRead(),
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func count(counter Counter) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		unread, err := counter.Execute(c.Context(), signedIn)
		if err != nil {
			return fmt.Errorf("count unread: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: unreadCountResponse{Count: unread}})
	}
}

func mark(marker Marker) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		notificationID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read notification id: %w", err)
		}
		if err := marker.Execute(c.Context(), signedIn, notificationID); err != nil {
			return fmt.Errorf("mark notification read: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{})
	}
}

func markAll(marker AllMarker) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		if err := marker.Execute(c.Context(), signedIn); err != nil {
			return fmt.Errorf("mark all notifications read: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{})
	}
}
