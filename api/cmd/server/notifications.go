package main

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	notificationhandler "github.com/benditandayusaputra/trustip/api/internal/modules/notification/handler"
	notificationrepository "github.com/benditandayusaputra/trustip/api/internal/modules/notification/repository"
	notificationservice "github.com/benditandayusaputra/trustip/api/internal/modules/notification/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
	"github.com/benditandayusaputra/trustip/api/internal/platform/sse"
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

const (
	streamsPerUser         = 3
	streamBacklog          = 16
	streamTicketLifetime   = 30 * time.Second
	streamTicketsPerMinute = 10
	streamHeartbeat        = 25 * time.Second
)

type liveComponents struct {
	hub     *sse.Hub
	tickets *sse.Tickets
}

func newLiveComponents(shutdown <-chan struct{}, clock sse.Clock) liveComponents {
	return liveComponents{
		hub:     sse.NewHub(shutdown, sse.Limits{StreamsPerUser: streamsPerUser, Backlog: streamBacklog}),
		tickets: sse.NewTickets(clock, streamTicketLifetime),
	}
}

type streamRoutes struct {
	authenticate fiber.Handler
	clock        serverClock
	securityLog  *slog.Logger
	appOrigin    string
}

func streamOptions(live liveComponents, routes streamRoutes) httpserver.StreamOptions {
	return httpserver.StreamOptions{
		Hub: live.hub, Tickets: live.tickets, Authenticate: routes.authenticate,
		TicketLimiter: ratelimit.New(ratelimit.Rule{Limit: streamTicketsPerMinute, Window: time.Minute}, routes.clock),
		SecurityLog:   routes.securityLog, AppOrigin: routes.appOrigin, Heartbeat: streamHeartbeat,
	}
}
