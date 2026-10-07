package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/config"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/logger"
)

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
	server := newServer(base, pool.Ping)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	logger.ForStream(base, logger.App).Info("server starting", "addr", cfg.HTTPAddr, "environment", cfg.Environment)
	if err := server.Run(ctx, cfg.HTTPAddr); err != nil {
		return fmt.Errorf("run http server: %w", err)
	}
	return nil
}

func newServer(base *slog.Logger, ready httpserver.ReadinessCheck) *httpserver.Server {
	return httpserver.New(httpserver.Options{
		AppLog:    logger.ForStream(base, logger.App),
		AccessLog: logger.ForStream(base, logger.Access),
		IDs:       idgen.UUIDv7{},
		Clock:     clock.System{},
		Build:     httpserver.ReadBuildInfo(),
		Ready:     ready,
	})
}
