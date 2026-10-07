package main

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"

	notificationhandler "github.com/benditandayusaputra/trustip/api/internal/modules/notification/handler"
	notificationrepository "github.com/benditandayusaputra/trustip/api/internal/modules/notification/repository"
	notificationservice "github.com/benditandayusaputra/trustip/api/internal/modules/notification/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type notificationInbox interface {
	notificationservice.NotificationReader
	notificationservice.NotificationMarker
}

type notificationComponents struct {
	inbox notificationInbox
	clock notificationservice.Clock
}

func buildNotificationComponents(pool *database.Pool, clock notificationservice.Clock) notificationComponents {
	return notificationComponents{inbox: notificationrepository.NewInbox(pool), clock: clock}
}

type notificationRoutes struct {
	authenticate fiber.Handler
	writeLimiter httpserver.RateLimiter
	securityLog  *slog.Logger
}

func notificationDeps(components notificationComponents, routes notificationRoutes) notificationhandler.Deps {
	return notificationhandler.Deps{
		Authenticate: routes.authenticate,
		List:         notificationservice.NewListNotifications(components.inbox),
		Count:        notificationservice.NewCountUnread(components.inbox),
		MarkRead:     notificationservice.NewMarkRead(components.inbox, components.clock),
		MarkAllRead:  notificationservice.NewMarkAllRead(components.inbox, components.clock),
		WriteLimiter: routes.writeLimiter,
		SecurityLog:  routes.securityLog,
	}
}
