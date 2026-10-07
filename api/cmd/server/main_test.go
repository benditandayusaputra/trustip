package main

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
	"github.com/benditandayusaputra/trustip/api/internal/platform/logger"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

const testRequestLimit = 5

func testNow() time.Time {
	return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
}

func newTestServer(ready httpserver.ReadinessCheck) *httpserver.Server {
	now := clock.NewManual(testNow())
	return newServer(logger.New(&bytes.Buffer{}, slog.LevelInfo), serverDeps{
		ready:   ready,
		edge:    security.NewEdgeVerifier(edgetest.TestKeys(), now),
		limiter: ratelimit.New(ratelimit.Rule{Limit: testRequestLimit, Window: time.Minute}, now),
		clock:   now,
	})
}

func alwaysReady(context.Context) error {
	return nil
}

func TestHealthz_ProcessAlive_ReturnsOK(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", http.NoBody)

	resp, err := newTestServer(alwaysReady).App().Test(req)
	if err != nil {
		t.Fatalf("request healthz: %v", err)
	}
	defer closeBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRoutes_EveryRouteWithoutAuthenticationIsListedAsPublic(t *testing.T) {
	public := httpserver.PublicRoutes()

	for _, route := range newTestServer(alwaysReady).App().GetRoutes(true) {
		method := route.Method
		if method == http.MethodHead {
			method = http.MethodGet
		}
		if !slices.Contains(public, httpserver.PublicRoute{Method: method, Path: route.Path}) {
			t.Errorf("%s %s has no authentication and is not in the public route list", route.Method, route.Path)
		}
	}
}

func closeBody(t *testing.T, resp *http.Response) {
	t.Helper()
	if err := resp.Body.Close(); err != nil {
		t.Errorf("close body: %v", err)
	}
}
