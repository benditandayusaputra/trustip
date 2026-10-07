package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type Refresher interface {
	Execute(ctx context.Context, presented string) (domain.SessionTokens, error)
}

type SessionEnder interface {
	Execute(ctx context.Context, presented string) error
}

func refresh(refresher Refresher, cookies CookiePolicy) fiber.Handler {
	return func(c fiber.Ctx) error {
		tokens, err := refresher.Execute(c.Context(), c.Cookies(cookies.RefreshName()))
		if err != nil {
			if apperror.From(err).Kind() == apperror.KindUnauthenticated {
				cookies.clear(c)
			}
			return fmt.Errorf("refresh session: %w", err)
		}
		cookies.setSession(c, tokens)
		return httpserver.Respond(c, httpserver.Reply{})
	}
}

func endSession(ender SessionEnder, cookies CookiePolicy) fiber.Handler {
	return func(c fiber.Ctx) error {
		err := ender.Execute(c.Context(), c.Cookies(cookies.RefreshName()))
		cookies.clear(c)
		if err != nil {
			return fmt.Errorf("end session: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{})
	}
}
