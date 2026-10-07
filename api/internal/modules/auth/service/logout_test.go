package service_test

import (
	"slices"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
)

func logoutDeps(fixture refreshFixture) service.LogoutDeps {
	return service.LogoutDeps{
		Sessions: fixture.sessions,
		Users:    fixture.sessions,
		Tokens:   fixture.tokens,
		Access:   fixture.access,
		Tx:       fixture.tx,
		Clock:    fixedClock{now: testNow()},
	}
}

func TestLogOut_FRAUTH03_RevokesTheFamilyAndItsAccessTokens(t *testing.T) {
	fixture := newRefreshFixture(storedSession(time.Time{}))

	err := service.NewLogOut(logoutDeps(fixture)).Execute(t.Context(), presentedCookie)

	if err != nil || !slices.Equal(fixture.sessions.revoked, []string{"family-1"}) {
		t.Fatalf("err = %v, revoked = %v", err, fixture.sessions.revoked)
	}
	if !fixture.access.until["family-1"].Equal(testNow().Add(domain.AccessTokenLifetime)) {
		t.Fatalf("access until = %v", fixture.access.until)
	}
}

func TestLogOut_WithoutAKnownToken_StillSucceeds(t *testing.T) {
	fixture := newRefreshFixture()

	for _, presented := range []string{"", "family-1.unknown"} {
		if err := service.NewLogOut(logoutDeps(fixture)).Execute(t.Context(), presented); err != nil {
			t.Fatalf("%q: err = %v", presented, err)
		}
	}
	if len(fixture.sessions.revoked) != 0 {
		t.Fatalf("revoked = %v", fixture.sessions.revoked)
	}
}

func TestLogOutEverywhere_RSEC19_RevokesEveryFamilyAndBansTheirAccessTokens(t *testing.T) {
	fixture := newRefreshFixture(storedSession(time.Time{}))

	err := service.NewLogOutEverywhere(logoutDeps(fixture)).Execute(t.Context(), presentedCookie)

	if err != nil || !slices.Equal(fixture.sessions.userWide, []string{knownUserID}) {
		t.Fatalf("err = %v, user-wide = %v", err, fixture.sessions.userWide)
	}
	for _, familyID := range []string{"family-1", "family-9"} {
		if !fixture.access.until[familyID].Equal(testNow().Add(domain.AccessTokenLifetime)) {
			t.Fatalf("%s access until = %v", familyID, fixture.access.until[familyID])
		}
	}
}

func TestLogOutEverywhere_NeedsAUsableSession(t *testing.T) {
	reused := storedSession(testNow().Add(-time.Minute))
	fixture := newRefreshFixture(reused)

	err := service.NewLogOutEverywhere(logoutDeps(fixture)).Execute(t.Context(), presentedCookie)

	if errorCode(err) != apperror.CodeSessionExpired || len(fixture.sessions.userWide) != 0 {
		t.Fatalf("err = %v, user-wide = %v", err, fixture.sessions.userWide)
	}
}
