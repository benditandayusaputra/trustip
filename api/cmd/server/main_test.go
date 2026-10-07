package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
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
	return newTestServerAt(ready, clock.NewManual(testNow()))
}

func newTestServerAt(ready httpserver.ReadinessCheck, now *clock.Manual) *httpserver.Server {
	return newTestServerWith(ready, now, testAuthComponents(now))
}

func newTestServerWith(ready httpserver.ReadinessCheck, now *clock.Manual, auth authComponents) *httpserver.Server {
	modules := testModules{
		auth: auth, org: testOrgComponents(now), files: testFileComponents(now), inbox: testNotificationComponents(now),
		products:  testProductComponents(now),
		proposals: testProposalComponents(now),
	}
	return newServerFrom(ready, now, modules)
}

type testModules struct {
	auth      authComponents
	org       orgComponents
	files     fileComponents
	inbox     notificationComponents
	products  productComponents
	proposals proposalComponents
}

func newServerFrom(ready httpserver.ReadinessCheck, now *clock.Manual, modules testModules) *httpserver.Server {
	return newServer(logger.New(&bytes.Buffer{}, slog.LevelInfo), serverDeps{
		ready:     ready,
		edge:      security.NewEdgeVerifier(edgetest.TestKeys(), now),
		limiter:   ratelimit.New(ratelimit.Rule{Limit: testRequestLimit, Window: time.Minute}, now),
		clock:     now,
		auth:      modules.auth,
		org:       modules.org,
		files:     modules.files,
		inbox:     modules.inbox,
		products:  modules.products,
		proposals: modules.proposals,
		appOrigin: testAppOrigin,
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

func TestRoutes_RAPI07_EveryRouteIsPublicOrAuthenticated(t *testing.T) {
	public := publicRoutes()
	server := newTestServer(alwaysReady)

	for index, route := range server.App().GetRoutes(true) {
		method := route.Method
		if method == http.MethodHead {
			method = http.MethodGet
		}
		if slices.Contains(public, httpserver.PublicRoute{Method: method, Path: route.Path}) {
			continue
		}
		req := httptest.NewRequestWithContext(t.Context(), method, withSampleParams(route.Path), http.NoBody)
		fromApp(req)
		edgetest.SignFrom(t, req, edgetest.Origin{At: testNow(), ClientIP: fmt.Sprintf("198.51.100.%d", index)})
		if got := perform(t, server, req); got.status != http.StatusUnauthorized {
			t.Errorf("%s %s answered %d without a session and is not in the public route list",
				method, route.Path, got.status)
		}
	}
}

func withSampleParams(path string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			segments[i] = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"
		}
	}
	return strings.Join(segments, "/")
}

func closeBody(t *testing.T, resp *http.Response) {
	t.Helper()
	if err := resp.Body.Close(); err != nil {
		t.Errorf("close body: %v", err)
	}
}
