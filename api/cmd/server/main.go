package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	authhandler "github.com/benditandayusaputra/trustip/api/internal/modules/auth/handler"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/config"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/logger"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

const requestsPerMinutePerIP = 300

func main() {
	cfg, err := config.Load(os.Environ())
	if err != nil {
		logger.New(os.Stderr, slog.LevelInfo).Error("refusing to start", "error", err)
		os.Exit(1)
	}
	base := logger.New(os.Stdout, cfg.LogLevel)
	appLog := logger.ForStream(base, logger.App)
	if err := serve(cfg, base); err != nil {
		appLog.Error("server stopped", "error", err)
		os.Exit(1)
	}
	appLog.Info("server stopped")
}

func serve(cfg config.Config, base *slog.Logger) error {
	pool, err := database.Open(context.Background(), database.Options{
		URL:               cfg.DatabaseURL,
		Logger:            logger.ForStream(base, logger.App),
		Clock:             clock.System{},
		RefuseSchemaOwner: cfg.Environment == config.Production,
	})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer pool.Close()
	systemClock := clock.System{}
	components, err := buildAuthComponents(cfg, pool, systemClock)
	if err != nil {
		return err
	}
	server := newServer(base, serverDeps{
		ready:   pool.Ping,
		edge:    security.NewEdgeVerifier(cfg.EdgeKeys, systemClock),
		limiter: ratelimit.New(ratelimit.Rule{Limit: requestsPerMinutePerIP, Window: time.Minute}, systemClock),
		clock:   systemClock,
		auth:    components,
	})
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	logger.ForStream(base, logger.App).Info("server starting", "addr", cfg.HTTPAddr, "environment", cfg.Environment)
	if err := server.Run(ctx, cfg.HTTPAddr); err != nil {
		return fmt.Errorf("run http server: %w", err)
	}
	return nil
}

type serverDeps struct {
	ready   httpserver.ReadinessCheck
	edge    httpserver.EdgeVerifier
	limiter httpserver.RateLimiter
	clock   serverClock
	auth    authComponents
}

type serverClock interface {
	Now() time.Time
}

func newServer(base *slog.Logger, deps serverDeps) *httpserver.Server {
	securityLog := logger.ForStream(base, logger.Security)
	server := httpserver.New(httpserver.Options{
		AppLog:      logger.ForStream(base, logger.App),
		AccessLog:   logger.ForStream(base, logger.Access),
		SecurityLog: securityLog,
		IDs:         idgen.UUIDv7{},
		Clock:       deps.clock,
		Build:       httpserver.ReadBuildInfo(),
		Ready:       deps.ready,
		Edge:        deps.edge,
		Limiter:     deps.limiter,
	})
	authhandler.Mount(server.API(), authDeps(deps.auth, securityLog, deps.clock))
	return server
}

func publicRoutes() []httpserver.PublicRoute {
	return append(httpserver.PublicRoutes(), authhandler.PublicRoutes()...)
}
