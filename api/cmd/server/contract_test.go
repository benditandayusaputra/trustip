package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
)

type contract struct {
	doc    *openapi3.T
	router routers.Router
}

func loadContract(t *testing.T) contract {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(filepath.Join("..", "..", "openapi", "openapi.yaml"))
	if err != nil {
		t.Fatalf("load openapi: %v", err)
	}
	if err := doc.Validate(loader.Context); err != nil {
		t.Fatalf("openapi document is invalid: %v", err)
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	return contract{doc: doc, router: router}
}

type exchange struct {
	request *http.Request
	status  int
	header  http.Header
	body    []byte
}

func perform(t *testing.T, server *httpserver.Server, req *http.Request) exchange {
	t.Helper()
	resp, err := server.App().Test(req)
	if err != nil {
		t.Fatalf("request %s: %v", req.URL.Path, err)
	}
	defer closeBody(t, resp)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return exchange{request: req, status: resp.StatusCode, header: resp.Header, body: body}
}

func (c contract) requireConforms(t *testing.T, got exchange) {
	t.Helper()
	route, pathParams, err := c.router.FindRoute(got.request)
	if _, undocumented := errors.AsType[*routers.RouteError](err); undocumented {
		c.requireErrorEnvelope(t, got)
		return
	}
	if err != nil {
		t.Fatalf("find route: %v", err)
	}
	input := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: got.request, PathParams: pathParams, Route: route,
		},
		Status:  got.status,
		Header:  got.header,
		Body:    io.NopCloser(bytes.NewReader(got.body)),
		Options: &openapi3filter.Options{IncludeResponseStatus: true},
	}
	if err := openapi3filter.ValidateResponse(context.Background(), input); err != nil {
		t.Fatalf("%s %d does not match the contract: %v\nbody: %s", got.request.URL.Path, got.status, err, got.body)
	}
}

func (c contract) requireErrorEnvelope(t *testing.T, got exchange) {
	t.Helper()
	if got.status < http.StatusBadRequest {
		t.Fatalf("undocumented route %s answered %d", got.request.URL.Path, got.status)
	}
	var payload any
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if err := c.doc.Components.Schemas["ErrorResponse"].Value.VisitJSON(payload); err != nil {
		t.Fatalf("%s %d is not an ErrorResponse: %v\nbody: %s", got.request.URL.Path, got.status, err, got.body)
	}
}

func signedGet(t *testing.T, path string) *http.Request {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, http.NoBody)
	edgetest.Sign(t, req, testNow())
	return req
}

func TestContract_Responses(t *testing.T) {
	spec := loadContract(t)
	failing := func(context.Context) error { return apperror.Unavailable(5) }
	oversized := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/meta",
		strings.NewReader(`{"note":"`+strings.Repeat("a", 70*1024)+`"}`))
	oversized.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	unsigned := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/meta", http.NoBody)
	cases := []struct {
		name    string
		server  *httpserver.Server
		request *http.Request
		status  int
	}{
		{"meta", newTestServer(alwaysReady), signedGet(t, "/api/v1/meta"), http.StatusOK},
		{"unsigned meta", newTestServer(alwaysReady), unsigned, http.StatusUnauthorized},
		{"ready", newTestServer(alwaysReady), signedGet(t, "/readyz"), http.StatusOK},
		{"database asleep", newTestServer(failing), signedGet(t, "/readyz"), http.StatusServiceUnavailable},
		{"unknown route", newTestServer(alwaysReady), signedGet(t, "/api/v1/nowhere"), http.StatusNotFound},
		{"oversized json", newTestServer(alwaysReady), oversized, http.StatusRequestEntityTooLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := perform(t, tc.server, tc.request)
			if got.status != tc.status {
				t.Fatalf("status = %d, want %d: %s", got.status, tc.status, got.body)
			}
			spec.requireConforms(t, got)
		})
	}
}

func TestContract_RateLimitedResponse(t *testing.T) {
	spec := loadContract(t)
	server := newTestServer(alwaysReady)
	for i := range testRequestLimit {
		if warmup := perform(t, server, signedGet(t, "/api/v1/meta")); warmup.status != http.StatusOK {
			t.Fatalf("warm-up request %d: status %d %s", i+1, warmup.status, warmup.body)
		}
	}

	got := perform(t, server, signedGet(t, "/api/v1/meta"))

	if got.status != http.StatusTooManyRequests {
		t.Fatalf("status = %d", got.status)
	}
	spec.requireConforms(t, got)
}

func TestContract_InternalErrorResponse(t *testing.T) {
	spec := loadContract(t)
	server := newTestServer(alwaysReady)
	server.API().Get("/probe/internal", func(fiber.Ctx) error { return errors.New("relation users does not exist") })

	got := perform(t, server, signedGet(t, "/api/v1/probe/internal"))

	if got.status != http.StatusInternalServerError {
		t.Fatalf("status = %d", got.status)
	}
	spec.requireConforms(t, got)
}

func TestContract_ErrorCodesMatchCatalog(t *testing.T) {
	spec := loadContract(t)
	var documented []string
	for _, value := range spec.doc.Components.Schemas["ErrorCode"].Value.Enum {
		documented = append(documented, value.(string))
	}
	var catalog []string
	for _, code := range apperror.Codes() {
		catalog = append(catalog, string(code))
	}
	slices.Sort(documented)
	slices.Sort(catalog)

	if !slices.Equal(documented, catalog) {
		t.Fatalf("openapi ErrorCode = %v\ncatalog = %v", documented, catalog)
	}
}

func TestContract_RegisterUserResponses(t *testing.T) {
	spec := loadContract(t)
	unknownField := strings.Replace(registration("baru@contoh.id"), `{`, `{"role":"OWNER",`, 1)
	invalidFields := `{"full_name":"S","email":"x","password":"a","whatsapp":"1",` +
		`"privacy_accepted":false,"captcha_token":"` + solvedCaptcha + `"}`
	forgedCaptcha := strings.Replace(registration("cap@contoh.id"), solvedCaptcha, "forged", 1)
	wrongType := postUsers(t, registration("tipe@contoh.id"))
	wrongType.Header.Set(fiber.HeaderContentType, "text/plain")
	cases := []struct {
		name    string
		request *http.Request
		status  int
	}{
		{"created", postUsers(t, registration("sari@contoh.id")), http.StatusCreated},
		{"invalid fields", postUsers(t, invalidFields), http.StatusUnprocessableEntity},
		{"captcha failed", postUsers(t, forgedCaptcha), http.StatusUnprocessableEntity},
		{"unknown field", postUsers(t, unknownField), http.StatusBadRequest},
		{"not json", wrongType, http.StatusUnsupportedMediaType},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := perform(t, newTestServer(alwaysReady), tc.request)
			if got.status != tc.status {
				t.Fatalf("status = %d, want %d: %s", got.status, tc.status, got.body)
			}
			spec.requireConforms(t, got)
		})
	}
}
