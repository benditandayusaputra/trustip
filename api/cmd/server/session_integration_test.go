package main

import (
	"net/http"
	"sync"
	"testing"

	authhandler "github.com/benditandayusaputra/trustip/api/internal/modules/auth/handler"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/revocation"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

func databaseBrowser(t *testing.T) *browser {
	t.Helper()
	pool := dbtest.MigratedPool(t)
	hasher, err := security.NewPasswordHasher(testPepper{}, security.MinimumArgon2Params())
	if err != nil {
		t.Fatalf("hasher: %v", err)
	}
	now := clock.NewManual(testNow())
	users, sessions := repository.NewUsers(pool), repository.NewSessions(pool)
	components := authComponents{
		captcha: acceptSolvedCaptcha{}, passwords: passwordHasher{hasher: hasher},
		users: users, credentials: users, sessions: sessions, userSessions: sessions,
		tokens: newTokenIssuer(testPepper{}), revoked: revocation.New(now), tx: pool,
		clock: now, ids: idgen.UUIDv7{}, cookies: authhandler.LocalCookies(),
	}
	server := newTestServerWith(alwaysReady, now, components)
	if got := perform(t, server, postUsers(t, registration("sari@contoh.id"))); got.status != http.StatusCreated {
		t.Fatalf("register = %d %s", got.status, got.body)
	}
	visitor := &browser{t: t, server: server, now: now}
	visitor.signIn()
	return visitor
}

func TestRefresh_RSEC18_ConcurrentRefreshesOfOneTokenAgreeOnOnePair(t *testing.T) {
	visitor := databaseBrowser(t)
	shared := visitor.refresh
	pairs := make([]string, 2)
	var workers sync.WaitGroup
	for i := range pairs {
		tab := &browser{t: t, server: visitor.server, now: visitor.now, refresh: shared, visits: 10 * (i + 1)}
		workers.Go(func() {
			got := tab.post("/api/v1/sessions/refresh")
			if got.status == http.StatusOK {
				pairs[i] = tab.refresh
			}
		})
	}
	workers.Wait()

	if pairs[0] == "" || pairs[0] != pairs[1] || pairs[0] == shared {
		t.Fatalf("pairs = %q", pairs)
	}
}

func TestLogOutEverywhere_RSEC19_RevokesRealSessionsAcrossFamilies(t *testing.T) {
	visitor := databaseBrowser(t)
	otherDevice := visitor.refresh
	visitor.signIn()

	visitor.post("/api/v1/sessions/logout-all")
	visitor.refresh = otherDevice
	got := visitor.post("/api/v1/sessions/refresh")

	requireCode(t, got, "SESSION_EXPIRED")
}
