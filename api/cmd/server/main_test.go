package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/logger"
)

func TestHealthz_ProcessAlive_ReturnsOK(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", http.NoBody)

	resp, err := newServer(logger.New(&bytes.Buffer{}, slog.LevelInfo)).App().Test(req)
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
	app := newServer(logger.New(&bytes.Buffer{}, slog.LevelInfo)).App()

	for _, route := range app.GetRoutes(true) {
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
