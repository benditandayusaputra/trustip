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
	usersPath    = "/users"
	sessionsPath = "/sessions"
	accountPath  = httpserver.APIPrefix + "/me"
)

type Registrar interface {
	Execute(ctx context.Context, cmd service.RegisterCommand) (domain.User, error)
}

type Authenticator interface {
	Execute(ctx context.Context, cmd service.LoginCommand) (service.LoginResult, error)
}

type LoginLimiters struct {
	PerIPMinute    httpserver.RateLimiter
	PerIPHour      httpserver.RateLimiter
	PerEmailMinute httpserver.RateLimiter
	PerEmailHour   httpserver.RateLimiter
}

type Deps struct {
	Register      Registrar
	Login         Authenticator
	SignupLimiter httpserver.RateLimiter
	LoginLimiters LoginLimiters
	Cookies       CookiePolicy
	SecurityLog   *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	limit := func(limiter httpserver.RateLimiter, keyOf httpserver.RateLimitKey) fiber.Handler {
		return httpserver.RateLimit(limiter, keyOf, deps.SecurityLog)
	}
	api.Post(usersPath, limit(deps.SignupLimiter, httpserver.ByClientIP), register(deps.Register))
	api.Post(sessionsPath,
		limit(deps.LoginLimiters.PerIPMinute, httpserver.ByClientIP),
		limit(deps.LoginLimiters.PerIPHour, httpserver.ByClientIP),
		limit(deps.LoginLimiters.PerEmailMinute, byLoginEmail),
		limit(deps.LoginLimiters.PerEmailHour, byLoginEmail),
		login(deps.Login, deps.Cookies),
	)
}

func PublicRoutes() []httpserver.PublicRoute {
	return []httpserver.PublicRoute{
		{Method: fiber.MethodPost, Path: httpserver.APIPrefix + usersPath},
		{Method: fiber.MethodPost, Path: httpserver.APIPrefix + sessionsPath},
	}
}
