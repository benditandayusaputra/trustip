package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
)

const (
	hostOnlyPrefix = "__Host-"
	accessCookie   = "trustip_at"
	refreshCookie  = "trustip_rt"
	rootCookiePath = "/"
)

type CookiePolicy struct {
	isProduction bool
}

func ProductionCookies() CookiePolicy {
	return CookiePolicy{isProduction: true}
}

func LocalCookies() CookiePolicy {
	return CookiePolicy{isProduction: false}
}

func (p CookiePolicy) AccessName() string {
	return p.name(accessCookie)
}

func (p CookiePolicy) RefreshName() string {
	return p.name(refreshCookie)
}

func (p CookiePolicy) name(base string) string {
	if p.isProduction {
		return hostOnlyPrefix + base
	}
	return base
}

func (p CookiePolicy) setSession(c fiber.Ctx, tokens domain.SessionTokens) {
	c.Cookie(p.cookie(p.AccessName(), tokens.Access, domain.AccessTokenLifetime))
	c.Cookie(p.cookie(p.RefreshName(), tokens.Refresh, domain.RefreshTokenLifetime))
}

func (p CookiePolicy) clear(c fiber.Ctx) {
	for _, name := range []string{p.AccessName(), p.RefreshName()} {
		expired := p.cookie(name, "", 0)
		expired.Expires = time.Unix(0, 0).UTC()
		c.Cookie(expired)
	}
}

func (p CookiePolicy) cookie(name, value string, lifetime time.Duration) *fiber.Cookie {
	return &fiber.Cookie{
		Name:     name,
		Value:    value,
		Path:     rootCookiePath,
		MaxAge:   int(lifetime.Seconds()),
		Secure:   p.isProduction,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
	}
}
