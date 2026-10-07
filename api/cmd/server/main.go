package main

import (
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/platform/config"
)

func main() {
	cfg, err := config.Load(os.Environ())
	if err != nil {
		slog.Error("refusing to start", "error", err)
		os.Exit(1)
	}
	if err := newApp().Listen(cfg.HTTPAddr); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func newApp() *fiber.App {
	app := fiber.New()
	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}
