package main

import (
	"github.com/gofiber/fiber/v3"

	authhandler "github.com/benditandayusaputra/trustip/api/internal/modules/auth/handler"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

func newTokenIssuer(jwtKeys security.SecretRing) tokenIssuer {
	return tokenIssuer{signer: security.NewTokenSigner(jwtKeys), refresh: security.NewRefreshTokens(jwtKeys)}
}

func newRefreshSession(components authComponents) *service.RefreshSession {
	return service.NewRefreshSession(service.RefreshDeps{
		Sessions: components.sessions,
		Tokens:   components.tokens,
		Access:   components.revoked,
		Tx:       components.tx,
		Clock:    components.clock,
		IDs:      components.ids,
	})
}

func logoutDeps(components authComponents) service.LogoutDeps {
	return service.LogoutDeps{
		Sessions: components.sessions,
		Users:    components.userSessions,
		Tokens:   components.tokens,
		Access:   components.revoked,
		Tx:       components.tx,
		Clock:    components.clock,
	}
}

func byRefreshFamily(cookies authhandler.CookiePolicy) httpserver.RateLimitKey {
	return func(c fiber.Ctx) string {
		if familyID, ok := security.FamilyOf(c.Cookies(cookies.RefreshName())); ok {
			return "family:" + familyID
		}
		return httpserver.ByClientIP(c)
	}
}
