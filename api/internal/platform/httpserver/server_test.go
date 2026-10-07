package httpserver_test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/goleak"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/logger"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

func fixedNow() time.Time {
	return time.Date(2026, 10, 19, 1, 2, 3, 0, time.UTC)
}

type testServer struct {
	server       *httpserver.Server
	accessLogs   *bytes.Buffer
	securityLogs *bytes.Buffer
}

type edgeKeys map[string][]byte

func (k edgeKeys) Lookup(id string) ([]byte, bool) {
	key, ok := k[id]
	return key, ok
}

const testRequestLimit = 10

func testEdgeKey() []byte {
	return bytes.Repeat([]byte{7}, 31)
}

func newTestServer() testServer {
	accessLogs, securityLogs := &bytes.Buffer{}, &bytes.Buffer{}
	now := clock.NewManual(fixedNow())
	server := httpserver.New(httpserver.Options{
		AppLog:      logger.New(&bytes.Buffer{}, 0),
		AccessLog:   logger.ForStream(logger.New(accessLogs, 0), logger.Access),
		SecurityLog: logger.ForStream(logger.New(securityLogs, 0), logger.Security),
		IDs:         idgen.NewSequence(),
		Clock:       now,
		Build:       httpserver.BuildInfo{Version: "abc123def456"},
		Ready:       func(context.Context) error { return nil },
		Edge:        security.NewEdgeVerifier(edgeKeys{"v1": testEdgeKey()}, now),
		Limiter:     ratelimit.New(ratelimit.Rule{Limit: testRequestLimit, Window: time.Minute}, now),
	})
	server.API().Get("/probe/deadline", func(c fiber.Ctx) error {
		deadline, ok := c.Context().Deadline()
		return httpserver.Respond(c, httpserver.Reply{Data: ok && time.Until(deadline) <= 10*time.Second})
	})
	server.API().Get("/probe/panic", func(c fiber.Ctx) error {
		var missing *httpserver.Reply
		return httpserver.Respond(c, *missing)
	})
	return testServer{server: server, accessLogs: accessLogs, securityLogs: securityLogs}
}

func (fixture testServer) do(t *testing.T, req *http.Request) response {
	t.Helper()
	signRequest(t, req, fixedNow())
	return send(t, fixture.server.App(), req)
}

func signRequest(t *testing.T, req *http.Request, signedAt time.Time) {
	t.Helper()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read request body: %v", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("nonce: %v", err)
	}
	signed := security.SignedRequest{
		KeyID: "v1", Method: req.Method, URI: req.URL.RequestURI(), Timestamp: strconv.FormatInt(signedAt.Unix(), 10),
		Nonce: hex.EncodeToString(nonce), ClientIP: "203.0.113.7", Body: body,
	}
	mac := hmac.New(sha256.New, testEdgeKey())
	mac.Write([]byte(signed.CanonicalString()))
	for name, value := range map[string]string{
		httpserver.HeaderEdgeKeyID: signed.KeyID, httpserver.HeaderEdgeTimestamp: signed.Timestamp,
		httpserver.HeaderEdgeNonce: signed.Nonce, httpserver.HeaderEdgeClientIP: signed.ClientIP,
		httpserver.HeaderEdgeSignature: hex.EncodeToString(mac.Sum(nil)),
	} {
		req.Header.Set(name, value)
	}
}

func TestMeta_ReturnsNameVersionAndServerTime(t *testing.T) {
	got := newTestServer().do(t, newRequest(t, "/api/v1/meta"))

	var data map[string]string
	if err := json.Unmarshal(got.payload.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["version"] != "abc123def456" || data["server_time"] != "2026-10-19T01:02:03Z" || data["name"] == "" {
		t.Fatalf("meta = %v", data)
	}
}

func TestSecurityHeaders_AreOnEveryResponseIncluding404(t *testing.T) {
	got := newTestServer().do(t, newRequest(t, "/api/v1/nowhere"))

	want := map[string]string{
		"Strict-Transport-Security":    "max-age=63072000; includeSubDomains",
		"X-Content-Type-Options":       "nosniff",
		"Referrer-Policy":              "strict-origin-when-cross-origin",
		"Permissions-Policy":           "geolocation=(self), camera=(), microphone=(), payment=()",
		"Cross-Origin-Opener-Policy":   "same-origin",
		"X-Frame-Options":              "DENY",
		"Cross-Origin-Resource-Policy": "same-origin",
	}
	for name, value := range want {
		if got.header.Get(name) != value {
			t.Errorf("%s = %q, want %q", name, got.header.Get(name), value)
		}
	}
}

func TestLimitJSONBody_JSONOver64KB_Returns413(t *testing.T) {
	body := strings.NewReader(`{"note":"` + strings.Repeat("a", 64*1024) + `"}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/meta", body)
	req.Header.Set("Content-Type", "application/json")

	got := newTestServer().do(t, req)

	if got.status != http.StatusRequestEntityTooLarge || got.payload.Code != "PAYLOAD_TOO_LARGE" {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
}

func TestLimitJSONBody_UploadOverLimit_Returns413(t *testing.T) {
	live := startLive(t, newTestServer())
	defer live.stop()
	conn, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", strings.TrimPrefix(live.baseURL, "http://"))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer closeConn(t, conn)
	request := "POST /api/v1/meta HTTP/1.1\r\nHost: api\r\n" +
		"Content-Type: multipart/form-data; boundary=x\r\nContent-Length: 1572865\r\n\r\n"

	if _, err := conn.Write([]byte(request)); err != nil {
		t.Fatalf("write: %v", err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	defer closeBody(t, resp)

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func closeConn(t *testing.T, conn net.Conn) {
	t.Helper()
	if err := conn.Close(); err != nil {
		t.Errorf("close conn: %v", err)
	}
}

func closeBody(t *testing.T, resp *http.Response) {
	t.Helper()
	if err := resp.Body.Close(); err != nil {
		t.Errorf("close body: %v", err)
	}
}

func TestDeadline_HandlerContextHasTenSecondDeadline(t *testing.T) {
	got := newTestServer().do(t, newRequest(t, "/api/v1/probe/deadline"))

	if string(got.payload.Data) != "true" {
		t.Fatalf("handler context deadline missing: %s", got.raw)
	}
}

func TestAccessLog_WritesOneLineWithRoutePatternAndStatus(t *testing.T) {
	fixture := newTestServer()

	fixture.do(t, newRequest(t, "/api/v1/meta?email=a@b.c"))

	var entry map[string]any
	if err := json.Unmarshal(fixture.accessLogs.Bytes(), &entry); err != nil {
		t.Fatalf("decode access log %q: %v", fixture.accessLogs.String(), err)
	}
	if entry["route"] != "/api/v1/meta" || entry["status"] != float64(200) || entry["log_type"] != "access" {
		t.Fatalf("access log = %v", entry)
	}
	if strings.Contains(fixture.accessLogs.String(), "email") {
		t.Fatalf("query string leaked into access log: %s", fixture.accessLogs.String())
	}
}

func TestAccessLog_PanickedRequest_IsLoggedWith500(t *testing.T) {
	fixture := newTestServer()

	got := fixture.do(t, newRequest(t, "/api/v1/probe/panic"))

	if got.status != http.StatusInternalServerError || !strings.Contains(fixture.accessLogs.String(), `"status":500`) {
		t.Fatalf("status = %d, access log = %s", got.status, fixture.accessLogs.String())
	}
}

func TestRun_ContextCancelled_ShutsDownWithoutLeaks(t *testing.T) {
	defer goleak.VerifyNone(t,
		goleak.IgnoreCurrent(),
		goleak.IgnoreAnyFunction("github.com/valyala/fasthttp.(*workerPool).Start.func2"),
	)
	live := startLive(t, newTestServer())

	live.stop()

	if err := <-live.done; err != nil {
		t.Fatalf("Run returned %v", err)
	}
}

type liveServer struct {
	baseURL string
	stop    context.CancelFunc
	done    chan error
}

func startLive(t *testing.T, fixture testServer) liveServer {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	addresses := make(chan string, 1)
	fixture.server.App().Hooks().OnListen(func(data fiber.ListenData) error {
		addresses <- net.JoinHostPort(data.Host, data.Port)
		return nil
	})
	done := make(chan error, 1)
	go func() { done <- fixture.server.Run(ctx, "127.0.0.1:0") }()
	live := liveServer{baseURL: "http://" + <-addresses, stop: cancel, done: done}
	waitUntilServing(t, live.baseURL)
	return live
}

func waitUntilServing(t *testing.T, baseURL string) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, baseURL+"/healthz", http.NoBody)
	if err != nil {
		t.Fatalf("build health request: %v", err)
	}
	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("server is not serving: %v", err)
	}
	closeBody(t, resp)
}

func TestRun_AddressInUse_ReturnsError(t *testing.T) {
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			t.Errorf("close listener: %v", err)
		}
	}()

	if err := newTestServer().server.Run(t.Context(), listener.Addr().String()); err == nil {
		t.Fatal("expected listen error")
	}
}

func TestReadyz_CheckPasses_Returns200(t *testing.T) {
	got := newTestServer().do(t, newRequest(t, "/readyz"))

	if got.status != http.StatusOK || !strings.Contains(got.raw, `"status":"ready"`) {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
}

func TestReadyz_DatabaseAsleep_Returns503WithRetryAfter(t *testing.T) {
	now := clock.NewManual(fixedNow())
	server := httpserver.New(httpserver.Options{
		AppLog:      logger.New(&bytes.Buffer{}, 0),
		AccessLog:   logger.New(&bytes.Buffer{}, 0),
		SecurityLog: logger.New(&bytes.Buffer{}, 0),
		IDs:         idgen.NewSequence(),
		Clock:       now,
		Ready:       func(context.Context) error { return apperror.Unavailable(5) },
		Edge:        security.NewEdgeVerifier(edgeKeys{"v1": testEdgeKey()}, now),
		Limiter:     ratelimit.New(ratelimit.Rule{Limit: testRequestLimit, Window: time.Minute}, now),
	})
	req := newRequest(t, "/readyz")
	signRequest(t, req, fixedNow())

	got := send(t, server.App(), req)

	if got.status != http.StatusServiceUnavailable || got.header.Get("Retry-After") != "5" {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
}

func TestEdgeSignature_UnsignedRequest_Returns401AndLogsSecurityEvent(t *testing.T) {
	fixture := newTestServer()

	got := send(t, fixture.server.App(), newRequest(t, "/api/v1/meta"))

	if got.status != http.StatusUnauthorized || got.payload.Code != "EDGE_SIGNATURE_INVALID" {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
	if !strings.Contains(fixture.securityLogs.String(), `"event":"edge_signature_invalid"`) {
		t.Fatalf("security log = %s", fixture.securityLogs.String())
	}
}

func TestEdgeSignature_ExpiredSignature_Returns401(t *testing.T) {
	fixture := newTestServer()
	req := newRequest(t, "/api/v1/meta")
	signRequest(t, req, fixedNow().Add(-61*time.Second))

	got := send(t, fixture.server.App(), req)

	if got.status != http.StatusUnauthorized {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
}

func TestEdgeSignature_ReplayedNonce_Returns401(t *testing.T) {
	fixture := newTestServer()
	first := newRequest(t, "/api/v1/meta")
	signRequest(t, first, fixedNow())
	replay := newRequest(t, "/api/v1/meta")
	replay.Header = first.Header.Clone()

	send(t, fixture.server.App(), first)
	got := send(t, fixture.server.App(), replay)

	if got.status != http.StatusUnauthorized {
		t.Fatalf("replay response = %d %s", got.status, got.raw)
	}
}

func TestEdgeSignature_Healthz_NeedsNoSignature(t *testing.T) {
	req := newRequest(t, "/healthz")

	resp, err := newTestServer().server.App().Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer closeBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestAccessLog_UsesSignedClientIP(t *testing.T) {
	fixture := newTestServer()
	req := newRequest(t, "/api/v1/meta")
	req.Header.Set("X-Forwarded-For", "10.9.9.9")

	fixture.do(t, req)

	if !strings.Contains(fixture.accessLogs.String(), `"ip":"203.0.113.7"`) {
		t.Fatalf("access log = %s", fixture.accessLogs.String())
	}
}

func TestRateLimit_EveryResponseCarriesRateLimitHeaders(t *testing.T) {
	got := newTestServer().do(t, newRequest(t, "/api/v1/meta"))

	if got.header.Get("RateLimit-Limit") != "10" || got.header.Get("RateLimit-Remaining") != "9" {
		t.Fatalf("headers = %v", got.header)
	}
	if got.header.Get("RateLimit-Reset") != "6" {
		t.Fatalf("RateLimit-Reset = %q", got.header.Get("RateLimit-Reset"))
	}
}

func TestRateLimit_OverLimit_Returns429WithRetryAfterAndLogsEvent(t *testing.T) {
	fixture := newTestServer()
	for range testRequestLimit {
		fixture.do(t, newRequest(t, "/api/v1/meta"))
	}

	got := fixture.do(t, newRequest(t, "/api/v1/meta"))

	if got.status != http.StatusTooManyRequests || got.payload.Code != "RATE_LIMITED" {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
	wantMessage := "Terlalu banyak percobaan. Coba lagi dalam 1 menit."
	if got.header.Get("Retry-After") != "6" || got.payload.Message != wantMessage {
		t.Fatalf("retry-after = %q, message = %q", got.header.Get("Retry-After"), got.payload.Message)
	}
	if !strings.Contains(fixture.securityLogs.String(), `"event":"rate_limited"`) {
		t.Fatalf("security log = %s", fixture.securityLogs.String())
	}
}
