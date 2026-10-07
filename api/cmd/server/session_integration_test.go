package main

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	auditrepository "github.com/benditandayusaputra/trustip/api/internal/modules/audit/repository"
	authhandler "github.com/benditandayusaputra/trustip/api/internal/modules/auth/handler"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/repository"
	orgrepository "github.com/benditandayusaputra/trustip/api/internal/modules/org/repository"
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
		users: users, accounts: users, credentials: users, sessions: sessions, userSessions: sessions,
		tokens: newTokenIssuer(testPepper{}), memberships: orgrepository.NewMemberships(pool),
		audit:      newAuditTrail(auditrepository.NewRecords(pool)),
		accessKeys: testPepper{}, revoked: revocation.New(now), tx: pool,
		clock: now, ids: idgen.UUIDv7{}, cookies: authhandler.LocalCookies(),
	}
	server := newServerFrom(alwaysReady, now, testModules{
		auth: components, org: buildOrgComponents(pool, now), files: databaseFileComponents(t, pool, now),
	})
	if got := perform(t, server, postUsers(t, registration("sari@contoh.id"))); got.status != http.StatusCreated {
		t.Fatalf("register = %d %s", got.status, got.body)
	}
	visitor := &browser{t: t, server: server, now: now, pool: pool}
	visitor.signIn()
	return visitor
}

func (b *browser) anotherUserOnTheSameServer(email string) *browser {
	b.t.Helper()
	b.visits++
	signup := postUsersFrom(b.t, registration(email), fmt.Sprintf("198.51.100.%d", b.visits))
	if got := perform(b.t, b.server, signup); got.status != http.StatusCreated {
		b.t.Fatalf("register %s = %d %s", email, got.status, got.body)
	}
	other := &browser{t: b.t, server: b.server, now: b.now, visits: b.visits}
	login := postSessions(b.t, attemptAt(credentials(email, acceptedInput), b.now.Now()))
	if got := other.send(login); got.status != http.StatusOK {
		b.t.Fatalf("login %s = %d %s", email, got.status, got.body)
	}
	return other
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
