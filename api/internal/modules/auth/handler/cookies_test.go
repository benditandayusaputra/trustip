package handler

import (
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestCookiePolicy_RSEC15_ProductionCookiesAreBoundToOneHost(t *testing.T) {
	production := ProductionCookies()

	access := production.cookie(production.AccessName(), "token", time.Minute)
	refresh := production.cookie(production.RefreshName(), "token", time.Minute)

	for _, cookie := range []*fiber.Cookie{access, refresh} {
		if cookie.Name[:len(hostOnlyPrefix)] != hostOnlyPrefix || cookie.Domain != "" || cookie.Path != "/" ||
			!cookie.Secure || !cookie.HTTPOnly || cookie.SameSite != fiber.CookieSameSiteLaxMode {
			t.Fatalf("cookie = %+v", cookie)
		}
	}
}

func TestCookiePolicy_RSEC15_LocalCookiesDropThePrefixAndSecure(t *testing.T) {
	local := LocalCookies()

	cookie := local.cookie(local.AccessName(), "token", time.Minute)

	if cookie.Name != accessCookie || cookie.Secure || !cookie.HTTPOnly {
		t.Fatalf("cookie = %+v", cookie)
	}
}
