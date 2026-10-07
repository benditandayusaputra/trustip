package httpserver

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	healthPath          = "/healthz"
	readyPath           = "/readyz"
	APIPrefix           = "/api/v1"
	metaPath            = "/meta"
	productName         = "Trustip API"
	shortRevisionLength = 12
)

type PublicRoute struct {
	Method string
	Path   string
}

func PublicRoutes() []PublicRoute {
	return []PublicRoute{
		{Method: fiber.MethodGet, Path: healthPath},
		{Method: fiber.MethodGet, Path: readyPath},
		{Method: fiber.MethodGet, Path: APIPrefix + metaPath},
	}
}

type BuildInfo struct {
	Version string
}

func ReadBuildInfo() BuildInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return BuildInfo{Version: "dev"}
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" && len(setting.Value) >= shortRevisionLength {
			return BuildInfo{Version: setting.Value[:shortRevisionLength]}
		}
	}
	return BuildInfo{Version: "dev"}
}

type metaBody struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	ServerTime string `json:"server_time"`
}

func health(c fiber.Ctx) error {
	if err := c.SendStatus(fiber.StatusOK); err != nil {
		return fmt.Errorf("write health status: %w", err)
	}
	return nil
}

type readinessBody struct {
	Status string `json:"status"`
}

func readiness(check ReadinessCheck) fiber.Handler {
	return func(c fiber.Ctx) error {
		if err := check(c.Context()); err != nil {
			return fmt.Errorf("readiness check: %w", err)
		}
		return Respond(c, Reply{Data: readinessBody{Status: "ready"}})
	}
}

func metaHandler(build BuildInfo, clock Clock) fiber.Handler {
	return func(c fiber.Ctx) error {
		return Respond(c, Reply{Data: metaBody{
			Name:       productName,
			Version:    build.Version,
			ServerTime: clock.Now().UTC().Format(time.RFC3339),
		}})
	}
}
