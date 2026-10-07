package httpserver_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
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
)

func fixedNow() time.Time {
	return time.Date(2026, 10, 19, 1, 2, 3, 0, time.UTC)
}

type testServer struct {
	server     *httpserver.Server
	accessLogs *bytes.Buffer
}

func newTestServer() testServer {
	accessLogs := &bytes.Buffer{}
	server := httpserver.New(httpserver.Options{
		AppLog:    logger.New(&bytes.Buffer{}, 0),
		AccessLog: logger.ForStream(logger.New(accessLogs, 0), logger.Access),
		IDs:       idgen.NewSequence(),
		Clock:     clock.NewManual(fixedNow()),
		Build:     httpserver.BuildInfo{Version: "abc123def456"},
		Ready:     func(context.Context) error { return nil },
	})
	server.API().Get("/probe/deadline", func(c fiber.Ctx) error {
		deadline, ok := c.Context().Deadline()
		return httpserver.Respond(c, httpserver.Reply{Data: ok && time.Until(deadline) <= 10*time.Second})
	})
	server.API().Get("/probe/panic", func(c fiber.Ctx) error {
		var missing *httpserver.Reply
		return httpserver.Respond(c, *missing)
	})
	return testServer{server: server, accessLogs: accessLogs}
}

func (fixture testServer) do(t *testing.T, req *http.Request) response {
	t.Helper()
	return send(t, fixture.server.App(), req)
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
	server := httpserver.New(httpserver.Options{
		AppLog:    logger.New(&bytes.Buffer{}, 0),
		AccessLog: logger.New(&bytes.Buffer{}, 0),
		IDs:       idgen.NewSequence(),
		Clock:     clock.NewManual(fixedNow()),
		Ready:     func(context.Context) error { return apperror.Unavailable(5) },
	})

	got := send(t, server.App(), newRequest(t, "/readyz"))

	if got.status != http.StatusServiceUnavailable || got.header.Get("Retry-After") != "5" {
		t.Fatalf("response = %d %s", got.status, got.raw)
	}
}
