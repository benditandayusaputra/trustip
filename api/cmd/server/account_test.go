package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
)

func TestMe_FRAUTH06_SignedInBrowserReadsItsOwnAccount(t *testing.T) {
	visitor := signedInBrowser(t)

	got := visitor.getAccount()

	requireCode(t, got, "OK")
	if !strings.Contains(string(got.body), `"full_name":"Sari Wulandari"`) {
		t.Fatalf("body = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestMe_FRAUTH06_WithoutSession_IsUnauthenticated(t *testing.T) {
	server, now := serverWithAccount(t)
	visitor := &browser{t: t, server: server, now: now}

	got := visitor.getAccount()

	requireCode(t, got, "UNAUTHENTICATED")
	loadContract(t).requireConforms(t, got)
}

func TestMe_RSEC16_TamperedOrExpiredAccessTokenIsSessionExpired(t *testing.T) {
	visitor := signedInBrowser(t)
	valid := visitor.access

	visitor.access = valid[:len(valid)-2] + "xx"
	tampered := visitor.getAccount()
	visitor.access = valid
	visitor.now.Advance(15 * time.Minute)
	expired := visitor.getAccount()

	requireCode(t, tampered, "SESSION_EXPIRED")
	requireCode(t, expired, "SESSION_EXPIRED")
}

func TestMe_RSEC19_AccessTokenStopsWorkingAfterLogOutEverywhere(t *testing.T) {
	visitor := signedInBrowser(t)
	stillCached := visitor.access

	visitor.post("/api/v1/sessions/logout-all")
	visitor.access = stillCached
	got := visitor.getAccount()

	requireCode(t, got, "SESSION_EXPIRED")
}

func crossSiteRegistration(t *testing.T, change func(*http.Request)) exchange {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/users",
		strings.NewReader(registration("sari@contoh.id")))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	fromApp(req)
	change(req)
	edgetest.Sign(t, req, testNow())
	return perform(t, newTestServer(alwaysReady), req)
}

func TestRequireSameOrigin_RSEC21_WritesFromAnotherSiteAreRejected(t *testing.T) {
	cases := map[string]func(*http.Request){
		"other origin":   func(req *http.Request) { req.Header.Set(fiber.HeaderOrigin, "https://evil.example") },
		"no origin":      func(req *http.Request) { req.Header.Del(fiber.HeaderOrigin) },
		"no client mark": func(req *http.Request) { req.Header.Del(httpserver.HeaderClient) },
		"wrong client":   func(req *http.Request) { req.Header.Set(httpserver.HeaderClient, "form") },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			got := crossSiteRegistration(t, change)

			requireCode(t, got, "CROSS_SITE_REJECTED")
			loadContract(t).requireConforms(t, got)
		})
	}
}

func TestRequireSameOrigin_RSEC22_WriteBodyMustBeJSON(t *testing.T) {
	got := crossSiteRegistration(t, func(req *http.Request) {
		req.Header.Set(fiber.HeaderContentType, "text/plain")
	})

	requireCode(t, got, "UNSUPPORTED_MEDIA_TYPE")
}
