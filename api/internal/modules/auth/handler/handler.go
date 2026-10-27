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
	usersPath     = "/users"
	sessionsPath  = "/sessions"
	refreshPath   = sessionsPath + "/refresh"
	logoutPath    = sessionsPath + "/logout"
	logoutAllPath = sessionsPath + "/logout-all"
	mePath        = "/me"
	passwordPath  = mePath + "/password"
	accountPath   = httpserver.APIPrefix + mePath
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

type PasswordLimiters struct {
	PerUserMinute httpserver.RateLimiter
	PerUserHour   httpserver.RateLimiter
}

type SessionEnders struct {
	ThisDevice SessionEnder
	AllDevices SessionEnder
}

type Deps struct {
	Authenticate    fiber.Handler
	Account         AccountGetter
	UpdateAccount   AccountUpdater
	ChangePassword  PasswordChanger
	Register        Registrar
	Login           Authenticator
	Refresh         Refresher
	LogOut          SessionEnders
	SignupLimiter   httpserver.RateLimiter
	LoginLimiters   LoginLimiters
	RefreshLimiter  httpserver.RateLimiter
	RefreshLimitKey httpserver.RateLimitKey
	WriteLimiter    httpserver.RateLimiter
	PasswordLimits  PasswordLimiters
	Cookies         CookiePolicy
	SecurityLog     *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	limit := func(limiter httpserver.RateLimiter, keyOf httpserver.RateLimitKey) fiber.Handler {
		return httpserver.RateLimit(limiter, keyOf, deps.SecurityLog)
	}
	api.Get(mePath, deps.Authenticate, currentAccount(deps.Account))
	api.Patch(mePath, deps.Authenticate, limit(deps.WriteLimiter, httpserver.BySignedInUser),
		updateAccount(deps.UpdateAccount))
	api.Post(passwordPath, deps.Authenticate,
		limit(deps.PasswordLimits.PerUserMinute, httpserver.BySignedInUser),
		limit(deps.PasswordLimits.PerUserHour, httpserver.BySignedInUser),
		changePassword(deps.ChangePassword),
	)
	api.Post(usersPath, limit(deps.SignupLimiter, httpserver.ByClientIP), register(deps.Register))
	api.Post(sessionsPath,
		limit(deps.LoginLimiters.PerIPMinute, httpserver.ByClientIP),
		limit(deps.LoginLimiters.PerIPHour, httpserver.ByClientIP),
		limit(deps.LoginLimiters.PerEmailMinute, byLoginEmail),
		limit(deps.LoginLimiters.PerEmailHour, byLoginEmail),
		login(deps.Login, deps.Cookies),
	)
	api.Post(refreshPath, limit(deps.RefreshLimiter, deps.RefreshLimitKey), refresh(deps.Refresh, deps.Cookies))
	api.Post(logoutPath, endSession(deps.LogOut.ThisDevice, deps.Cookies))
	api.Post(logoutAllPath, endSession(deps.LogOut.AllDevices, deps.Cookies))
}

func PublicRoutes() []httpserver.PublicRoute {
	return []httpserver.PublicRoute{
		{Method: fiber.MethodPost, Path: httpserver.APIPrefix + usersPath},
		{Method: fiber.MethodPost, Path: httpserver.APIPrefix + sessionsPath},
		{Method: fiber.MethodPost, Path: httpserver.APIPrefix + refreshPath},
		{Method: fiber.MethodPost, Path: httpserver.APIPrefix + logoutPath},
		{Method: fiber.MethodPost, Path: httpserver.APIPrefix + logoutAllPath},
	}
}
