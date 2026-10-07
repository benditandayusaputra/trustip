package main

import (
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v3"
)

const defaultHTTPAddr = ":8080"

func main() {
	app := newApp()
	if err := app.Listen(listenAddr()); err != nil {
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

func listenAddr() string {
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		return addr
	}
	return defaultHTTPAddr
}
