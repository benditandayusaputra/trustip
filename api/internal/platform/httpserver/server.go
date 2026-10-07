package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	uploadBodyLimit = 1536 * 1024
	shutdownTimeout = 10 * time.Second
	readTimeout     = 15 * time.Second
	idleTimeout     = 60 * time.Second
)

type Clock interface {
	Now() time.Time
}

type ReadinessCheck func(ctx context.Context) error

type Options struct {
	AppLog      *slog.Logger
	AccessLog   *slog.Logger
	SecurityLog *slog.Logger
	IDs         IDGenerator
	Clock       Clock
	Build       BuildInfo
	Ready       ReadinessCheck
	Edge        EdgeVerifier
	Limiter     RateLimiter
}

type Server struct {
	app *fiber.App
	api fiber.Router
}

func New(opts Options) *Server {
	errorHandler := NewErrorHandler(opts.AppLog)
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
		Immutable:    true,
		BodyLimit:    uploadBodyLimit,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
	})
	app.Use(
		RequestID(opts.IDs),
		AccessLog(opts.AccessLog, opts.Clock, errorHandler),
		Recover(),
		SecurityHeaders(),
		LimitJSONBody(),
		EdgeSignature(opts.Edge, opts.SecurityLog),
		RateLimit(opts.Limiter, ByClientIP, opts.SecurityLog),
		Deadline(requestTimeout),
	)
	app.Get(healthPath, health)
	app.Get(readyPath, readiness(opts.Ready))
	api := app.Group(apiPrefix)
	api.Get(metaPath, metaHandler(opts.Build, opts.Clock))
	return &Server{app: app, api: api}
}

func (s *Server) App() *fiber.App {
	return s.app
}

func (s *Server) API() fiber.Router {
	return s.api
}

func (s *Server) Run(ctx context.Context, addr string) error {
	err := s.app.Listen(addr, fiber.ListenConfig{
		GracefulContext:       ctx,
		ShutdownTimeout:       shutdownTimeout,
		DisableStartupMessage: true,
	})
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	return nil
}
