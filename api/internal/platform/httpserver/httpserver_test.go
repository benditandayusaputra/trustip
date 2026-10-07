package httpserver_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/logger"
)

type envelope struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Data    json.RawMessage   `json:"data"`
	Errors  []json.RawMessage `json:"errors"`
	Meta    map[string]any    `json:"meta"`
}

type response struct {
	status  int
	header  http.Header
	raw     string
	payload envelope
}

func newTestApp(logs *bytes.Buffer) *fiber.App {
	log := logger.New(logs, slog.LevelInfo)
	app := fiber.New(fiber.Config{ErrorHandler: httpserver.NewErrorHandler(log, log)})
	app.Use(httpserver.RequestID(idgen.NewSequence()), httpserver.Recover())
	app.Get("/ok", func(c fiber.Ctx) error {
		return httpserver.Respond(c, httpserver.Reply{Message: "Tersimpan.", Data: map[string]string{"id": "7"}})
	})
	app.Get("/empty-list", func(c fiber.Ctx) error {
		return httpserver.RespondList(c, httpserver.Page[string]{})
	})
	app.Get("/missing", func(fiber.Ctx) error { return fmt.Errorf("get drop: %w", apperror.NotFound()) })
	app.Get("/elsewhere", func(fiber.Ctx) error { return fmt.Errorf("get drop: %w", apperror.OutsideScope()) })
	app.Get("/invalid", func(fiber.Ctx) error {
		return apperror.Validation(apperror.FieldError{Field: "qty", Code: "MIN", Message: "Jumlah minimal 1."})
	})
	app.Get("/internal", func(fiber.Ctx) error { return errors.New("relation users does not exist") })
	app.Get("/limited", func(fiber.Ctx) error { return apperror.RateLimited(30) })
	app.Get("/boom", func(c fiber.Ctx) error {
		var missing *httpserver.Reply
		return httpserver.Respond(c, *missing)
	})
	return app
}

func newRequest(t *testing.T, path string) *http.Request {
	t.Helper()
	return httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, http.NoBody)
}

func call(t *testing.T, app *fiber.App, path string) response {
	t.Helper()
	return send(t, app, newRequest(t, path))
}

func send(t *testing.T, app *fiber.App, req *http.Request) response {
	t.Helper()
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request %s: %v", req.URL.Path, err)
	}
	body, err := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); err != nil || closeErr != nil {
		t.Fatalf("read body: %v, %v", err, closeErr)
	}
	result := response{status: resp.StatusCode, header: resp.Header, raw: string(body)}
	if err := json.Unmarshal(body, &result.payload); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return result
}

func TestRespond_Success_UsesStandardEnvelope(t *testing.T) {
	got := call(t, newTestApp(&bytes.Buffer{}), "/ok")

	if got.status != http.StatusOK || got.payload.Code != "OK" || got.payload.Message != "Tersimpan." {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
	if got.payload.Meta["request_id"] != got.header.Get(httpserver.HeaderRequestID) {
		t.Fatalf("meta request id %v differs from header", got.payload.Meta["request_id"])
	}
	if got.header.Get("Cache-Control") != "no-store" {
		t.Fatalf("cache-control = %q", got.header.Get("Cache-Control"))
	}
}

func TestRespondList_NoItems_ReturnsEmptyArrayAndNullCursor(t *testing.T) {
	got := call(t, newTestApp(&bytes.Buffer{}), "/empty-list")

	if string(got.payload.Data) != "[]" || !strings.Contains(got.raw, `"next_cursor":null`) {
		t.Fatalf("response = %s", got.raw)
	}
}

func TestErrorHandler_WrappedNotFound_Returns404Envelope(t *testing.T) {
	got := call(t, newTestApp(&bytes.Buffer{}), "/missing")

	if got.status != http.StatusNotFound || got.payload.Code != "NOT_FOUND" {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
	if got.payload.Errors != nil || got.payload.Data != nil {
		t.Fatalf("error response carries data or errors: %s", got.raw)
	}
}

func TestErrorHandler_RLOG06_CrossOrganizationAccessIsANotFoundAndASecurityEvent(t *testing.T) {
	logs := &bytes.Buffer{}

	got := call(t, newTestApp(logs), "/elsewhere")
	plain := call(t, newTestApp(&bytes.Buffer{}), "/missing")

	if got.status != http.StatusNotFound || got.payload.Code != plain.payload.Code ||
		got.payload.Message != plain.payload.Message {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
	if !strings.Contains(logs.String(), `"event":"cross_organization_access"`) {
		t.Fatalf("logs = %s", logs.String())
	}
}

func TestErrorHandler_PlainNotFoundIsNotASecurityEvent(t *testing.T) {
	logs := &bytes.Buffer{}

	call(t, newTestApp(logs), "/missing")

	if strings.Contains(logs.String(), "cross_organization_access") {
		t.Fatalf("logs = %s", logs.String())
	}
}

func TestErrorHandler_Validation_ListsFieldErrors(t *testing.T) {
	got := call(t, newTestApp(&bytes.Buffer{}), "/invalid")

	if got.status != http.StatusUnprocessableEntity || len(got.payload.Errors) != 1 {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
}

func TestErrorHandler_UnexpectedError_HidesInternalsAndShowsRequestID(t *testing.T) {
	var logs bytes.Buffer

	got := call(t, newTestApp(&logs), "/internal")

	if got.status != http.StatusInternalServerError || strings.Contains(got.raw, "relation") {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
	requestID := got.header.Get(httpserver.HeaderRequestID)
	if !strings.HasSuffix(got.payload.Message, "Kode: "+requestID+".") {
		t.Fatalf("message = %q, want request id %s", got.payload.Message, requestID)
	}
	if !strings.Contains(logs.String(), "relation users does not exist") {
		t.Fatalf("internal error was not logged: %s", logs.String())
	}
}

func TestRecover_Panic_Returns500AndLogsStackOnce(t *testing.T) {
	var logs bytes.Buffer

	got := call(t, newTestApp(&logs), "/boom")

	if got.status != http.StatusInternalServerError || strings.Contains(got.raw, "nil pointer") {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
	if strings.Count(logs.String(), "\n") != 1 || !strings.Contains(logs.String(), `"stack"`) {
		t.Fatalf("logs = %s", logs.String())
	}
}

func TestErrorHandler_RateLimited_SetsRetryAfter(t *testing.T) {
	got := call(t, newTestApp(&bytes.Buffer{}), "/limited")

	if got.status != http.StatusTooManyRequests || got.header.Get("Retry-After") != "30" {
		t.Fatalf("response = %d retry-after %q", got.status, got.header.Get("Retry-After"))
	}
}

func TestErrorHandler_UnknownRoute_Returns404Envelope(t *testing.T) {
	got := call(t, newTestApp(&bytes.Buffer{}), "/nowhere")

	if got.status != http.StatusNotFound || got.payload.Code != "NOT_FOUND" {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
}

func TestRequestID_ValidIncomingID_IsReused(t *testing.T) {
	incoming := "0192f0c4-8a11-7d55-b0aa-3c1f2e4d5a6b"

	req := newRequest(t, "/ok")
	req.Header.Set(httpserver.HeaderRequestID, incoming)

	got := send(t, newTestApp(&bytes.Buffer{}), req)

	if got.header.Get(httpserver.HeaderRequestID) != incoming {
		t.Fatalf("request id = %q", got.header.Get(httpserver.HeaderRequestID))
	}
}

func TestRequestID_MalformedIncomingID_IsReplaced(t *testing.T) {
	req := newRequest(t, "/ok")
	req.Header.Set(httpserver.HeaderRequestID, "x\" injected")

	got := send(t, newTestApp(&bytes.Buffer{}), req)

	if got.header.Get(httpserver.HeaderRequestID) != "00000000-0000-7000-8000-000000000001" {
		t.Fatalf("request id = %q", got.header.Get(httpserver.HeaderRequestID))
	}
}
