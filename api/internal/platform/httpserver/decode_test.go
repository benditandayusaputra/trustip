package httpserver_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type probeBody struct {
	Name string `json:"name"`
}

func decodeApp() *fiber.App {
	app := newTestApp(&bytes.Buffer{})
	app.Post("/decode", func(c fiber.Ctx) error {
		var body probeBody
		if err := httpserver.DecodeJSON(c, &body); err != nil {
			return fmt.Errorf("decode probe: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: body})
	})
	return app
}

func postDecode(t *testing.T, contentType, body string) response {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/decode", strings.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	return send(t, decodeApp(), req)
}

func TestDecodeJSON_RSEC22_Layer1Rejections(t *testing.T) {
	const unsupported = http.StatusUnsupportedMediaType
	cases := []struct {
		name        string
		contentType string
		body        string
		status      int
		code        string
	}{
		{"form post", "application/x-www-form-urlencoded", "name=x", unsupported, "UNSUPPORTED_MEDIA_TYPE"},
		{"text plain", "text/plain", `{"name":"x"}`, unsupported, "UNSUPPORTED_MEDIA_TYPE"},
		{"unknown field", "application/json", `{"name":"x","role":"OWNER"}`, http.StatusBadRequest, "UNKNOWN_FIELD"},
		{"broken json", "application/json", `{"name":`, http.StatusBadRequest, "MALFORMED_REQUEST"},
		{"wrong type", "application/json", `{"name":7}`, http.StatusBadRequest, "MALFORMED_REQUEST"},
		{"two documents", "application/json", `{"name":"x"}{"name":"y"}`, http.StatusBadRequest, "MALFORMED_REQUEST"},
		{"invalid utf8", "application/json", "{\"name\":\"\xff\"}", http.StatusBadRequest, "MALFORMED_REQUEST"},
		{"empty body", "application/json", "", http.StatusBadRequest, "MALFORMED_REQUEST"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := postDecode(t, tc.contentType, tc.body)

			if got.status != tc.status || got.payload.Code != tc.code {
				t.Fatalf("response = %d %s", got.status, got.raw)
			}
		})
	}
}

func TestDecodeJSON_JSONWithCharset_IsAccepted(t *testing.T) {
	got := postDecode(t, "application/json; charset=utf-8", `{"name":"Roti <3 Keju"}`)

	if got.status != http.StatusOK || !strings.Contains(got.raw, "Roti \\u003c3 Keju") {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
}

func TestNormalizeLine_RSEC50_TrimsCollapsesAndComposes(t *testing.T) {
	got, ok := httpserver.NormalizeLine("  Toko   Bu\u0308di  Jaya  ")

	if !ok || got != "Toko Büdi Jaya" {
		t.Fatalf("got %q, ok = %v", got, ok)
	}
}

func TestNormalizeLine_RSEC51_RejectsControlZeroWidthAndDirectionMarks(t *testing.T) {
	for _, raw := range []string{"Sari\x00", "Sa\u200Bri", "Sari\u202E", "Sari\nWulan", "Sari\tWulan", "\uFEFFSari"} {
		if _, ok := httpserver.NormalizeLine(raw); ok {
			t.Errorf("%q was accepted", raw)
		}
	}
}

func TestNormalizePassword_RSEC01_AppliesNFKC(t *testing.T) {
	if got := httpserver.NormalizePassword("\uFF2B\uFF50\uFF17#mQz2vX"); got != "Kp7#mQz2vX" {
		t.Fatalf("got %q", got)
	}
}
