package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz_ProcessAlive_ReturnsOK(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", http.NoBody)

	resp, err := newApp().Test(req)
	if err != nil {
		t.Fatalf("request healthz: %v", err)
	}
	defer closeBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func closeBody(t *testing.T, resp *http.Response) {
	t.Helper()
	if err := resp.Body.Close(); err != nil {
		t.Errorf("close body: %v", err)
	}
}
