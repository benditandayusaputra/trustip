package main

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	authhandler "github.com/benditandayusaputra/trustip/api/internal/modules/auth/handler"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/revocation"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

var errSessionFamilyRevoked = errors.New("session family was revoked")

func newTokenIssuer(jwtKeys security.SecretRing) tokenIssuer {
	return tokenIssuer{signer: security.NewTokenSigner(jwtKeys), refresh: security.NewRefreshTokens(jwtKeys)}
}

func newRefreshSession(components authComponents) *service.RefreshSession {
	return service.NewRefreshSession(service.RefreshDeps{
		Sessions:    components.sessions,
		Tokens:      components.tokens,
		Memberships: components.memberships,
		Access:      components.revoked,
		Tx:          components.tx,
		Clock:       components.clock,
		IDs:         components.ids,
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

type accessVerifier struct {
	signer  *security.TokenSigner
	revoked *revocation.List
	clock   service.Clock
}

func newAccessVerifier(components authComponents) accessVerifier {
	return accessVerifier{
		signer:  security.NewTokenSigner(components.accessKeys),
		revoked: components.revoked,
		clock:   components.clock,
	}
}

func (v accessVerifier) VerifyAccess(token string) (actor.Actor, error) {
	claims, err := v.signer.Verify(token, v.clock.Now())
	if err != nil {
		return actor.Actor{}, fmt.Errorf("verify access token: %w", err)
	}
	if v.revoked.IsRevoked(claims.SessionID) {
		return actor.Actor{}, errSessionFamilyRevoked
	}
	return actor.Actor{UserID: claims.Subject, SessionFamily: claims.SessionID, Membership: membershipOf(claims)}, nil
}

func membershipOf(claims security.AccessClaims) actor.Membership {
	return actor.Membership{
		OrganizationID:   claims.Organization,
		OrganizationType: actor.OrganizationType(claims.OrganizationType),
		Role:             actor.Role(claims.Role),
	}
}
