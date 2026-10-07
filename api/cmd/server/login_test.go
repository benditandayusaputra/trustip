package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
)

type loginAttempt struct {
	body     string
	at       time.Time
	clientIP string
}

func attemptAt(body string, at time.Time) loginAttempt {
	return loginAttempt{body: body, at: at, clientIP: edgetest.ClientIP}
}

func postSessions(t *testing.T, attempt loginAttempt) *http.Request {
	t.Helper()
	body := strings.NewReader(attempt.body)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/sessions", body)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	edgetest.SignFrom(t, req, edgetest.Origin{At: attempt.at, ClientIP: attempt.clientIP})
	return req
}

const (
	acceptedInput = "Kp7#mQz2vX"
	mistypedInput = "Salah#2026x"
)

func signIn(t *testing.T, server *httpserver.Server, attempt loginAttempt) exchange {
	t.Helper()
	return perform(t, server, postSessions(t, attempt))
}

func credentials(email, password string) string {
	return `{"email":"` + email + `","password":"` + password + `","captcha_token":"` + solvedCaptcha + `"}`
}

func serverWithAccount(t *testing.T) (*httpserver.Server, *clock.Manual) {
	t.Helper()
	now := clock.NewManual(testNow())
	server := newTestServerAt(alwaysReady, now)
	if got := perform(t, server, postUsers(t, registration("sari@contoh.id"))); got.status != http.StatusCreated {
		t.Fatalf("register = %d %s", got.status, got.body)
	}
	return server, now
}

func cookieNamed(header http.Header, name string) *http.Cookie {
	for _, line := range header.Values("Set-Cookie") {
		cookie, err := http.ParseSetCookie(line)
		if err == nil && cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func TestLogin_RSEC14_TokensLiveOnlyInHttpOnlyCookies(t *testing.T) {
	server, now := serverWithAccount(t)

	got := signIn(t, server, attemptAt(credentials("sari@contoh.id", acceptedInput), now.Now()))

	if got.status != http.StatusOK {
		t.Fatalf("login = %d %s", got.status, got.body)
	}
	for name, maxAge := range map[string]int{"trustip_at": 900, "trustip_rt": 2592000} {
		cookie := cookieNamed(got.header, name)
		if cookie == nil || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" ||
			cookie.MaxAge != maxAge || cookie.Domain != "" {
			t.Fatalf("%s cookie = %+v", name, cookie)
		}
		if strings.Contains(string(got.body), cookie.Value) {
			t.Fatalf("%s value leaked into the response body", name)
		}
	}
}

func TestLogin_RSEC08_WrongPasswordAndUnknownEmailLookTheSame(t *testing.T) {
	server, now := serverWithAccount(t)

	knownEmail := signIn(t, server, attemptAt(credentials("sari@contoh.id", mistypedInput), now.Now()))
	unknownEmail := signIn(t, server, attemptAt(credentials("asing@contoh.id", acceptedInput), now.Now()))

	isUniform := stripRequestID(knownEmail.body) == stripRequestID(unknownEmail.body)
	if knownEmail.status != http.StatusUnauthorized || !isUniform {
		t.Fatalf("known email = %d %s\nunknown email = %d %s",
			knownEmail.status, knownEmail.body, unknownEmail.status, unknownEmail.body)
	}
}

func TestLogin_BRAUTH01_FiveFailuresLockTheAccountEvenForTheRightPassword(t *testing.T) {
	server, now := serverWithAccount(t)
	for i := range 5 {
		now.Advance(time.Minute)
		got := signIn(t, server, attemptAt(credentials("sari@contoh.id", mistypedInput), now.Now()))
		if i < 4 && got.status != http.StatusUnauthorized {
			t.Fatalf("failure %d = %d %s", i+1, got.status, got.body)
		}
	}
	now.Advance(time.Minute)

	got := signIn(t, server, attemptAt(credentials("sari@contoh.id", acceptedInput), now.Now()))

	if got.status != http.StatusTooManyRequests || !strings.Contains(string(got.body), "ACCOUNT_LOCKED") ||
		got.header.Get("Retry-After") != "840" || cookieNamed(got.header, "trustip_at") != nil {
		t.Fatalf("response = %d %v %s", got.status, got.header, got.body)
	}
}

func TestLogin_SixthAttemptOnOneEmailWithinAMinute_IsRateLimitedAcrossIPs(t *testing.T) {
	server, now := serverWithAccount(t)
	for i := range 5 {
		ip := "198.51.100." + string(rune('1'+i))
		attempt := attemptAt(credentials("sari@contoh.id", mistypedInput), now.Now())
		attempt.clientIP = ip
		perform(t, server, postSessions(t, attempt))
	}

	attempt := attemptAt(credentials("sari@contoh.id", acceptedInput), now.Now())
	attempt.clientIP = "198.51.100.9"

	got := perform(t, server, postSessions(t, attempt))

	if got.status != http.StatusTooManyRequests || !strings.Contains(string(got.body), "RATE_LIMITED") {
		t.Fatalf("response = %d %s", got.status, got.body)
	}
}

func TestContract_LoginResponses(t *testing.T) {
	spec := loadContract(t)
	cases := []struct {
		name   string
		body   string
		status int
	}{
		{"signed in", credentials("sari@contoh.id", acceptedInput), http.StatusOK},
		{"wrong password", credentials("sari@contoh.id", mistypedInput), http.StatusUnauthorized},
		{"captcha failed", strings.Replace(credentials("sari@contoh.id", "x"), solvedCaptcha, "forged", 1), 422},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server, now := serverWithAccount(t)

			got := signIn(t, server, attemptAt(tc.body, now.Now()))

			if got.status != tc.status {
				t.Fatalf("status = %d, want %d: %s", got.status, tc.status, got.body)
			}
			spec.requireConforms(t, got)
		})
	}
}

func stripRequestID(body []byte) string {
	text := string(body)
	if start := strings.Index(text, `"meta"`); start >= 0 {
		return text[:start]
	}
	return text
}
