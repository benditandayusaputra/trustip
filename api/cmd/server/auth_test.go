package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

const solvedCaptcha = "solved-captcha"

type testPepper struct{}

func (testPepper) Active() (string, []byte) { return "v1", bytes.Repeat([]byte{3}, 32) }

func (testPepper) Lookup(string) ([]byte, bool) { return bytes.Repeat([]byte{3}, 32), true }

type acceptSolvedCaptcha struct{}

func (acceptSolvedCaptcha) Verify(_ context.Context, challenge service.CaptchaChallenge) error {
	if challenge.Token != solvedCaptcha {
		return apperror.New(apperror.CodeCaptchaFailed)
	}
	return nil
}

type memoryUsers struct {
	mu     sync.Mutex
	emails map[string]bool
}

func (m *memoryUsers) CreateUser(_ context.Context, user domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.emails[user.Email] {
		return domain.ErrEmailTaken
	}
	m.emails[user.Email] = true
	return nil
}

func testAuthComponents(now *clock.Manual) authComponents {
	hasher, err := security.NewPasswordHasher(testPepper{}, security.MinimumArgon2Params())
	if err != nil {
		panic(err)
	}
	return authComponents{
		captcha: acceptSolvedCaptcha{},
		hasher:  passwordHasher{hasher: hasher},
		users:   &memoryUsers{emails: map[string]bool{}},
		clock:   now,
		ids:     idgen.NewSequence(),
	}
}

func postUsers(t *testing.T, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/users", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	edgetest.Sign(t, req, testNow())
	return req
}

func registration(email string) string {
	return `{"full_name":"Sari Wulandari","email":"` + email + `","password":"Kp7#mQz2vX",` +
		`"whatsapp":"0812-3456-7890","privacy_accepted":true,"captcha_token":"` + solvedCaptcha + `"}`
}
