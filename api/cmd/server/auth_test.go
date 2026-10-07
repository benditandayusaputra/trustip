package main

import (
	"bytes"
	"context"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	authhandler "github.com/benditandayusaputra/trustip/api/internal/modules/auth/handler"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/revocation"
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

func (m *memoryUsers) GetAccount(_ context.Context, userID string) (domain.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	credentials := m.byID(userID)
	if credentials.UserID == "" {
		return domain.Account{}, apperror.NotFound()
	}
	return domain.Account{ID: credentials.UserID, Email: credentials.Email, FullName: credentials.FullName}, nil
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
	mu     sync.Mutex
	byHash map[string]domain.StoredSession
}

func (m *memorySessions) CreateSession(_ context.Context, session domain.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byHash[string(session.TokenHash)] = domain.StoredSession{
		SessionIdentity: session.SessionIdentity, ExpiresAt: session.ExpiresAt,
	}
	return nil
}

func (m *memorySessions) LockSession(_ context.Context, tokenHash []byte) (domain.StoredSession, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.byHash[string(tokenHash)]
	return session, ok, nil
}

func (m *memorySessions) MarkRotated(_ context.Context, sessionID string, at time.Time) error {
	m.update(func(session *domain.StoredSession) {
		if session.ID == sessionID {
			session.RotatedAt = at
		}
	})
	return nil
}

func (m *memorySessions) RevokeFamily(_ context.Context, familyID string, at time.Time) error {
	m.update(func(session *domain.StoredSession) {
		if session.FamilyID == familyID && session.RevokedAt.IsZero() {
			session.RevokedAt = at
		}
	})
	return nil
}

func (m *memorySessions) RevokeUser(_ context.Context, userID string, at time.Time) ([]string, error) {
	families := map[string]bool{}
	m.update(func(session *domain.StoredSession) {
		if session.UserID == userID && session.RevokedAt.IsZero() {
			session.RevokedAt = at
			families[session.FamilyID] = true
		}
	})
	return slices.Sorted(maps.Keys(families)), nil
}

func (m *memorySessions) update(change func(*domain.StoredSession)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for hash, session := range m.byHash {
		change(&session)
		m.byHash[hash] = session
	}
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
	sessions := &memorySessions{byHash: map[string]domain.StoredSession{}}
	return authComponents{
		captcha:      acceptSolvedCaptcha{},
		passwords:    passwordHasher{hasher: hasher},
		users:        users,
		accounts:     users,
		credentials:  users,
		sessions:     sessions,
		userSessions: sessions,
		tokens:       newTokenIssuer(testPepper{}),
		accessKeys:   testPepper{},
		revoked:      revocation.New(now),
		tx:           inlineTx{},
		clock:        now,
		ids:          idgen.NewSequence(),
		cookies:      authhandler.LocalCookies(),
	}
}

func postUsers(t *testing.T, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/users", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	fromApp(req)
	edgetest.Sign(t, req, testNow())
	return req
}

func registration(email string) string {
	return `{"full_name":"Sari Wulandari","email":"` + email + `","password":"Kp7#mQz2vX",` +
		`"whatsapp":"0812-3456-7890","privacy_accepted":true,"captcha_token":"` + solvedCaptcha + `"}`
}

const testAppOrigin = "http://localhost:5173"

func fromApp(req *http.Request) {
	req.Header.Set(fiber.HeaderOrigin, testAppOrigin)
	req.Header.Set(httpserver.HeaderClient, "web")
}
