package captcha_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/platform/captcha"
)

const (
	testSecret   = "0x4AAAAAAAtest-secret"
	testHostname = "trustip.id"
	widgetAnswer = "answer-from-widget"
)

type siteVerify struct {
	received url.Values
	calls    int
}

func fakeSiteVerify(t *testing.T, status int, body string) (*httptest.Server, *siteVerify) {
	t.Helper()
	recorded := &siteVerify{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		recorded.calls++
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		recorded.received = r.PostForm
		writer.WriteHeader(status)
		if _, err := io.WriteString(writer, body); err != nil {
			t.Errorf("write body: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return server, recorded
}

func newVerifier(endpoint string, timeout time.Duration) *captcha.Turnstile {
	return captcha.NewTurnstile(captcha.Options{
		Secret:           testSecret,
		ExpectedHostname: testHostname,
		Endpoint:         endpoint,
		Timeout:          timeout,
	})
}

func loginChallenge() captcha.Challenge {
	return captcha.Challenge{Token: widgetAnswer, Action: "login", RemoteIP: "203.0.113.7"}
}

func requireCaptchaFailed(t *testing.T, err error) {
	t.Helper()
	if apperror.From(err).Code() != apperror.CodeCaptchaFailed {
		t.Fatalf("err = %v, want CAPTCHA_FAILED", err)
	}
}

func TestVerify_RSEC10_MatchingHostnameAndActionPasses(t *testing.T) {
	server, recorded := fakeSiteVerify(t, http.StatusOK,
		`{"success":true,"hostname":"trustip.id","action":"login","error-codes":[]}`)

	err := newVerifier(server.URL, time.Second).Verify(t.Context(), loginChallenge())
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	want := url.Values{"secret": {testSecret}, "response": {widgetAnswer}, "remoteip": {"203.0.113.7"}}
	if recorded.received.Encode() != want.Encode() {
		t.Fatalf("form = %v", recorded.received)
	}
}

func TestVerify_BRAUTH05_Rejections(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"BR-AUTH-05 token rejected", http.StatusOK, `{"success":false,"error-codes":["invalid-input-response"]}`},
		{"BR-AUTH-05 hostname differs", http.StatusOK, `{"success":true,"hostname":"evil.example","action":"login"}`},
		{"BR-AUTH-05 action differs", http.StatusOK, `{"success":true,"hostname":"trustip.id","action":"signup"}`},
		{"BR-AUTH-05 service error", http.StatusInternalServerError, `oops`},
		{"BR-AUTH-05 unreadable answer", http.StatusOK, `{"success":tru`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server, _ := fakeSiteVerify(t, tc.status, tc.body)

			requireCaptchaFailed(t, newVerifier(server.URL, time.Second).Verify(t.Context(), loginChallenge()))
		})
	}
}

func TestVerify_BRAUTH05_EmptyTokenFailsWithoutCallingTheService(t *testing.T) {
	server, recorded := fakeSiteVerify(t, http.StatusOK, `{"success":true}`)
	challenge := loginChallenge()
	challenge.Token = ""

	requireCaptchaFailed(t, newVerifier(server.URL, time.Second).Verify(t.Context(), challenge))
	if recorded.calls != 0 {
		t.Fatalf("service called %d times", recorded.calls)
	}
}

func TestVerify_BRAUTH05_OversizedTokenFailsWithoutCallingTheService(t *testing.T) {
	server, recorded := fakeSiteVerify(t, http.StatusOK, `{"success":true}`)
	challenge := loginChallenge()
	challenge.Token = strings.Repeat("a", 2049)

	requireCaptchaFailed(t, newVerifier(server.URL, time.Second).Verify(t.Context(), challenge))
	if recorded.calls != 0 {
		t.Fatalf("service called %d times", recorded.calls)
	}
}

func TestVerify_RSEC11_SilentServiceFailsAfterTimeout(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	t.Cleanup(server.Close)
	t.Cleanup(func() { close(release) })

	requireCaptchaFailed(t, newVerifier(server.URL, 20*time.Millisecond).Verify(t.Context(), loginChallenge()))
}

func TestVerify_RSEC11_UnreachableServiceFails(t *testing.T) {
	server, _ := fakeSiteVerify(t, http.StatusOK, `{"success":true}`)
	server.Close()

	requireCaptchaFailed(t, newVerifier(server.URL, time.Second).Verify(t.Context(), loginChallenge()))
}

func TestVerify_RSEC12_OfficialTestingKeyResultSkipsHostnameAndAction(t *testing.T) {
	server, _ := fakeSiteVerify(t, http.StatusOK,
		`{"challenge_ts":"2026-10-07T07:40:19.628Z","error-codes":[],"hostname":"example.com",`+
			`"metadata":{"result_with_testing_key":true},"success":true}`)

	if err := newVerifier(server.URL, time.Second).Verify(t.Context(), loginChallenge()); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

func TestVerify_RSEC12_OfficialTestingKeyFailureStillFails(t *testing.T) {
	server, _ := fakeSiteVerify(t, http.StatusOK,
		`{"error-codes":["invalid-input-response"],"success":false,"metadata":{"result_with_testing_key":true}}`)

	requireCaptchaFailed(t, newVerifier(server.URL, time.Second).Verify(t.Context(), loginChallenge()))
}
