package httpserver

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/platform/sse"
)

const (
	StreamTicketsPath     = "/stream-tickets"
	StreamPath            = "/stream"
	heartbeatFrame        = ": ping\n\n"
	mimeEventStream       = "text/event-stream"
	headerAccelBuffering  = "X-Accel-Buffering"
	streamTicketParameter = "ticket"
)

var errStreamClosed = errors.New("live stream closed")

type StreamOptions struct {
	Hub           *sse.Hub
	Tickets       *sse.Tickets
	Authenticate  fiber.Handler
	TicketLimiter RateLimiter
	SecurityLog   *slog.Logger
	AppOrigin     string
	Heartbeat     time.Duration
}

func MountStream(api fiber.Router, opts StreamOptions) {
	ticketLimit := RateLimit(opts.TicketLimiter, BySignedInUser, opts.SecurityLog)
	api.Post(StreamTicketsPath, opts.Authenticate, ticketLimit, issueStreamTicket(opts.Tickets))
	api.Get(StreamPath, openStream(opts))
}

func StreamPublicRoutes() []PublicRoute {
	return []PublicRoute{{Method: fiber.MethodGet, Path: APIPrefix + StreamPath}}
}

type streamTicketBody struct {
	Ticket    string `json:"ticket"`
	ExpiresAt string `json:"expires_at"`
}

func issueStreamTicket(tickets *sse.Tickets) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, ok := ActorFrom(c)
		if !ok {
			return apperror.New(apperror.CodeUnauthenticated)
		}
		ticket := tickets.Issue(signedIn.UserID)
		return Respond(c, Reply{Data: streamTicketBody{
			Ticket: ticket.Value, ExpiresAt: ticket.ExpiresAt.UTC().Format(time.RFC3339),
		}})
	}
}

func openStream(opts StreamOptions) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set(fiber.HeaderAccessControlAllowOrigin, opts.AppOrigin)
		c.Set(fiber.HeaderVary, fiber.HeaderOrigin)
		userID, err := opts.Tickets.Redeem(c.Query(streamTicketParameter))
		if err != nil {
			return apperror.New(apperror.CodeUnauthenticated).WithCause(err)
		}
		stream := opts.Hub.Open(userID)
		c.Set(fiber.HeaderContentType, mimeEventStream)
		c.Set(fiber.HeaderCacheControl, "no-store")
		c.Set(headerAccelBuffering, "no")
		return c.SendStreamWriter(func(w *bufio.Writer) {
			relay(w, opts, stream)
		})
	}
}

func relay(w *bufio.Writer, opts StreamOptions, stream *sse.Stream) {
	defer opts.Hub.Close(stream)
	heartbeat := time.NewTicker(opts.Heartbeat)
	defer heartbeat.Stop()
	frame := heartbeatFrame
	for writeFrame(w, frame) == nil {
		next, err := nextFrame(opts.Hub, stream, heartbeat.C)
		if err != nil {
			return
		}
		frame = next
	}
}

func nextFrame(hub *sse.Hub, stream *sse.Stream, heartbeat <-chan time.Time) (string, error) {
	select {
	case event, open := <-stream.Events():
		if !open {
			return "", errStreamClosed
		}
		return "event: " + event.Name + "\ndata: " + event.Data + "\n\n", nil
	case <-heartbeat:
		return heartbeatFrame, nil
	case <-hub.Done():
		return "", errStreamClosed
	}
}

func writeFrame(w *bufio.Writer, frame string) error {
	if _, err := w.WriteString(frame); err != nil {
		return fmt.Errorf("write stream frame: %w", err)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush stream frame: %w", err)
	}
	return nil
}
