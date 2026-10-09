package httpserver_test

import (
	"bufio"
	"bytes"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
	"github.com/benditandayusaputra/trustip/api/internal/platform/sse"
)

const streamAppOrigin = "https://app.trustip.test"

type streamFixture struct {
	live     liveServer
	hub      *sse.Hub
	tickets  *sse.Tickets
	shutdown chan struct{}
}

func startStream(t *testing.T) streamFixture {
	t.Helper()
	fixture := newTestServer()
	now := clock.NewManual(fixedNow())
	shutdown := make(chan struct{})
	hub := sse.NewHub(shutdown, sse.Limits{StreamsPerUser: 1, Backlog: 4})
	tickets := sse.NewTickets(now, 30*time.Second)
	httpserver.MountStream(fixture.server.API(), httpserver.StreamOptions{
		Hub: hub, Tickets: tickets, AppOrigin: streamAppOrigin, Heartbeat: time.Hour,
		Authenticate:  func(c fiber.Ctx) error { return c.Next() },
		TicketLimiter: ratelimit.New(ratelimit.Rule{Limit: 1, Window: time.Minute}, now),
		SecurityLog:   slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
	})
	live := startLive(t, fixture)
	t.Cleanup(func() {
		closeOnce(shutdown)
		live.stop()
		<-live.done
	})
	return streamFixture{live: live, hub: hub, tickets: tickets, shutdown: shutdown}
}

func closeOnce(signal chan struct{}) {
	select {
	case <-signal:
	default:
		close(signal)
	}
}

func (f streamFixture) open(t *testing.T, ticket string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet,
		f.live.baseURL+"/api/v1/stream?ticket="+ticket, http.NoBody)
	if err != nil {
		t.Fatalf("build stream request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	return resp
}

func readFrame(t *testing.T, reader *bufio.Reader) string {
	t.Helper()
	var frame strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read frame: %v", err)
		}
		if line == "\n" {
			return frame.String()
		}
		frame.WriteString(line)
	}
}

func TestStream_FRNTF02_ValidTicket_RelaysEventsOfItsUser(t *testing.T) {
	stream := startStream(t)
	ticket := stream.tickets.Issue("user-a")

	resp := stream.open(t, ticket.Value)
	defer closeBody(t, resp)
	reader := bufio.NewReader(resp.Body)
	readFrame(t, reader)
	stream.hub.Publish("user-a", sse.Event{Name: "notification", Data: `{"id":"n-1"}`})

	if got := readFrame(t, reader); got != "event: notification\ndata: {\"id\":\"n-1\"}\n" {
		t.Fatalf("frame = %q", got)
	}
	if resp.Header.Get("Content-Type") != "text/event-stream" ||
		resp.Header.Get("Access-Control-Allow-Origin") != streamAppOrigin {
		t.Fatalf("headers = %v", resp.Header)
	}
}

func TestStream_RSEC26_UsedTicket_IsUnauthenticated(t *testing.T) {
	stream := startStream(t)
	ticket := stream.tickets.Issue("user-a")
	if _, err := stream.tickets.Redeem(ticket.Value); err != nil {
		t.Fatalf("redeem: %v", err)
	}

	resp := stream.open(t, ticket.Value)
	defer closeBody(t, resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestStream_StreamLimitReached_ClosesTheOldestStream(t *testing.T) {
	stream := startStream(t)
	first := stream.open(t, stream.tickets.Issue("user-a").Value)
	defer closeBody(t, first)
	oldest := bufio.NewReader(first.Body)
	readFrame(t, oldest)

	second := stream.open(t, stream.tickets.Issue("user-a").Value)
	defer closeBody(t, second)
	readFrame(t, bufio.NewReader(second.Body))

	if _, err := oldest.ReadString('\n'); err == nil {
		t.Fatal("oldest stream still open after a newer one took its place")
	}
}

func TestStream_ServerShutdown_EndsTheStream(t *testing.T) {
	stream := startStream(t)
	resp := stream.open(t, stream.tickets.Issue("user-a").Value)
	defer closeBody(t, resp)
	reader := bufio.NewReader(resp.Body)
	readFrame(t, reader)

	closeOnce(stream.shutdown)

	if _, err := reader.ReadString('\n'); err == nil {
		t.Fatal("stream still open after shutdown")
	}
}
