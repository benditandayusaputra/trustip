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
	authhandler "github.com/benditandayusaputra/trustip/api/internal/modules/auth/handler"
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
	mu      sync.Mutex
	byEmail map[string]domain.Credentials
}

func (m *memoryUsers) CreateUser(_ context.Context, user domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, taken := m.byEmail[user.Email]; taken {
		return domain.ErrEmailTaken
	}
	m.byEmail[user.Email] = domain.Credentials{
		UserID: user.ID, Email: user.Email, FullName: user.FullName, Password: user.Password,
	}
	return nil
}

func (m *memoryUsers) FindCredentials(_ context.Context, email string) (domain.Credentials, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	credentials, ok := m.byEmail[email]
	return credentials, ok, nil
}

func (m *memoryUsers) LockLoginAttempts(_ context.Context, userID string) (domain.LoginAttempts, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.byID(userID).Attempts, nil
}

func (m *memoryUsers) SaveLoginAttempts(_ context.Context, userID string, attempts domain.LoginAttempts) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	credentials := m.byID(userID)
	credentials.Attempts = attempts
	m.byEmail[credentials.Email] = credentials
	return nil
}

func (m *memoryUsers) SavePassword(_ context.Context, userID string, password domain.StoredPassword) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	credentials := m.byID(userID)
	credentials.Password = password
	m.byEmail[credentials.Email] = credentials
	return nil
}

func (m *memoryUsers) byID(userID string) domain.Credentials {
	for _, credentials := range m.byEmail {
		if credentials.UserID == userID {
			return credentials
		}
	}
	return domain.Credentials{}
}

type memorySessions struct {
	mu      sync.Mutex
	created []domain.Session
}

func (m *memorySessions) CreateSession(_ context.Context, session domain.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.created = append(m.created, session)
	return nil
}

type inlineTx struct{}

func (inlineTx) WithinTx(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

func testAuthComponents(now *clock.Manual) authComponents {
	hasher, err := security.NewPasswordHasher(testPepper{}, security.MinimumArgon2Params())
	if err != nil {
		panic(err)
	}
	users := &memoryUsers{byEmail: map[string]domain.Credentials{}}
	return authComponents{
		captcha:     acceptSolvedCaptcha{},
		passwords:   passwordHasher{hasher: hasher},
		users:       users,
		credentials: users,
		sessions:    &memorySessions{},
		tokens:      tokenIssuer{signer: security.NewTokenSigner(testPepper{})},
		tx:          inlineTx{},
		clock:       now,
		ids:         idgen.NewSequence(),
		cookies:     authhandler.LocalCookies(),
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
