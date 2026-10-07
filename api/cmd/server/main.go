package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/config"
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	appLog.Info("server starting", "addr", cfg.HTTPAddr, "environment", cfg.Environment)
	if err := newServer(base).Run(ctx, cfg.HTTPAddr); err != nil {
		appLog.Error("server stopped", "error", err)
		stop()
		os.Exit(1)
	}
	appLog.Info("server stopped")
}

func newServer(base *slog.Logger) *httpserver.Server {
	return httpserver.New(httpserver.Options{
		AppLog:    logger.ForStream(base, logger.App),
		AccessLog: logger.ForStream(base, logger.Access),
		IDs:       idgen.UUIDv7{},
		Clock:     clock.System{},
		Build:     httpserver.ReadBuildInfo(),
	})
}
