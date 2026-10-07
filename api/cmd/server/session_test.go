package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
)

type browser struct {
	t       *testing.T
	server  *httpserver.Server
	now     *clock.Manual
	refresh string
	visits  int
}

func signedInBrowser(t *testing.T) *browser {
	t.Helper()
	server, now := serverWithAccount(t)
	visitor := &browser{t: t, server: server, now: now}
	visitor.signIn()
	return visitor
}

func (b *browser) signIn() {
	b.t.Helper()
	got := b.send(postSessions(b.t, attemptAt(credentials("sari@contoh.id", acceptedInput), b.now.Now())))
	if got.status != http.StatusOK {
		b.t.Fatalf("login = %d %s", got.status, got.body)
	}
}

func (b *browser) post(path string) exchange {
	b.t.Helper()
	req := httptest.NewRequestWithContext(b.t.Context(), http.MethodPost, path, http.NoBody)
	if b.refresh != "" {
		req.Header.Set("Cookie", "trustip_rt="+b.refresh)
	}
	b.visits++
	edgetest.SignFrom(b.t, req, edgetest.Origin{At: b.now.Now(), ClientIP: fmt.Sprintf("198.51.100.%d", b.visits)})
	return b.send(req)
}

func (b *browser) send(req *http.Request) exchange {
	b.t.Helper()
	got := perform(b.t, b.server, req)
	if cookie := cookieNamed(got.header, "trustip_rt"); cookie != nil {
		b.refresh = cookie.Value
	}
	return got
}

func requireCode(t *testing.T, got exchange, code string) {
	t.Helper()
	status := http.StatusOK
	if code != "OK" {
		status = apperror.New(apperror.Code(code)).Status()
	}
	if got.status != status || !strings.Contains(string(got.body), `"code":"`+code+`"`) {
		t.Fatalf("response = %d %s, want %d %s", got.status, got.body, status, code)
	}
}

func TestRefresh_RSEC17_RotatesBothCookies(t *testing.T) {
	visitor := signedInBrowser(t)
	before := visitor.refresh

	got := visitor.post("/api/v1/sessions/refresh")

	requireCode(t, got, "OK")
	if visitor.refresh == before || cookieNamed(got.header, "trustip_at") == nil {
		t.Fatalf("refresh cookie did not rotate: %v", got.header.Values("Set-Cookie"))
	}
	loadContract(t).requireConforms(t, got)
}

func TestRefresh_RSEC18_TwoTabsWithinTenSecondsGetTheSamePair(t *testing.T) {
	visitor := signedInBrowser(t)
	shared := visitor.refresh
	first := visitor.post("/api/v1/sessions/refresh")
	visitor.now.Advance(4 * time.Second)
	visitor.refresh = shared

	second := visitor.post("/api/v1/sessions/refresh")

	firstPair := cookieNamed(first.header, "trustip_rt").Value + cookieNamed(first.header, "trustip_at").Value
	secondPair := cookieNamed(second.header, "trustip_rt").Value + cookieNamed(second.header, "trustip_at").Value
	if second.status != http.StatusOK || firstPair != secondPair {
		t.Fatalf("second = %d %s, pairs equal = %v", second.status, second.body, firstPair == secondPair)
	}
}

func TestRefresh_RSEC17_ReusingARotatedTokenRevokesTheWholeFamily(t *testing.T) {
	visitor := signedInBrowser(t)
	stolen := visitor.refresh
	visitor.post("/api/v1/sessions/refresh")
	current := visitor.refresh
	visitor.now.Advance(11 * time.Second)

	visitor.refresh = stolen
	reuse := visitor.post("/api/v1/sessions/refresh")
	visitor.refresh = current
	afterwards := visitor.post("/api/v1/sessions/refresh")

	requireCode(t, reuse, "SESSION_REVOKED")
	requireCode(t, afterwards, "SESSION_EXPIRED")
	if cleared := cookieNamed(reuse.header, "trustip_rt"); cleared == nil || cleared.Value != "" {
		t.Fatalf("refresh cookie was not cleared: %v", reuse.header.Values("Set-Cookie"))
	}
}

func TestRefresh_WithoutCookie_IsSessionExpired(t *testing.T) {
	server, now := serverWithAccount(t)
	visitor := &browser{t: t, server: server, now: now}

	got := visitor.post("/api/v1/sessions/refresh")

	requireCode(t, got, "SESSION_EXPIRED")
	loadContract(t).requireConforms(t, got)
}

func TestLogOut_FRAUTH03_ClearsCookiesAndEndsTheSession(t *testing.T) {
	visitor := signedInBrowser(t)
	token := visitor.refresh

	got := visitor.post("/api/v1/sessions/logout")
	visitor.refresh = token
	refreshAfter := visitor.post("/api/v1/sessions/refresh")

	requireCode(t, got, "OK")
	for _, name := range []string{"trustip_at", "trustip_rt"} {
		if cleared := cookieNamed(got.header, name); cleared == nil || cleared.Value != "" {
			t.Fatalf("%s was not cleared: %v", name, got.header.Values("Set-Cookie"))
		}
	}
	requireCode(t, refreshAfter, "SESSION_EXPIRED")
	loadContract(t).requireConforms(t, got)
}

func TestLogOutEverywhere_FRAUTH03_EndsSessionsOnOtherDevices(t *testing.T) {
	visitor := signedInBrowser(t)
	otherDevice := visitor.refresh
	visitor.now.Advance(time.Minute)
	visitor.signIn()

	got := visitor.post("/api/v1/sessions/logout-all")
	visitor.refresh = otherDevice
	otherRefresh := visitor.post("/api/v1/sessions/refresh")

	requireCode(t, got, "OK")
	requireCode(t, otherRefresh, "SESSION_EXPIRED")
}

func TestRefresh_ThirtyFirstRefreshOfOneFamilyInAMinute_IsRateLimited(t *testing.T) {
	visitor := signedInBrowser(t)
	for i := range 30 {
		if got := visitor.post("/api/v1/sessions/refresh"); got.status != http.StatusOK {
			t.Fatalf("refresh %d = %d %s", i+1, got.status, got.body)
		}
	}

	got := visitor.post("/api/v1/sessions/refresh")

	requireCode(t, got, "RATE_LIMITED")
}
