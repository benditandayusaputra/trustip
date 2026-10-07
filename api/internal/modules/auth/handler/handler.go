package handler

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	usersPath   = "/users"
	accountPath = httpserver.APIPrefix + "/me"
)

type Registrar interface {
	Execute(ctx context.Context, cmd service.RegisterCommand) (domain.User, error)
}

type Deps struct {
	Register      Registrar
	SignupLimiter httpserver.RateLimiter
	SecurityLog   *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	signupLimit := httpserver.RateLimit(deps.SignupLimiter, httpserver.ByClientIP, deps.SecurityLog)
	api.Post(usersPath, signupLimit, register(deps.Register))
}

func PublicRoutes() []httpserver.PublicRoute {
	return []httpserver.PublicRoute{
		{Method: fiber.MethodPost, Path: httpserver.APIPrefix + usersPath},
	}
}
