package service_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
)

const presentedCookie = "family-1.presented"

type fakeRotator struct {
	byHash   map[string]domain.StoredSession
	rotated  []string
	created  []domain.Session
	revoked  []string
	userWide []string
}

func (f *fakeRotator) LockSession(_ context.Context, tokenHash []byte) (domain.StoredSession, bool, error) {
	session, ok := f.byHash[string(tokenHash)]
	return session, ok, nil
}

func (f *fakeRotator) MarkRotated(_ context.Context, sessionID string, _ time.Time) error {
	f.rotated = append(f.rotated, sessionID)
	return nil
}

func (f *fakeRotator) CreateSession(_ context.Context, session domain.Session) error {
	f.created = append(f.created, session)
	return nil
}

func (f *fakeRotator) RevokeFamily(_ context.Context, familyID string, _ time.Time) error {
	f.revoked = append(f.revoked, familyID)
	return nil
}

func (f *fakeRotator) RevokeUser(_ context.Context, userID string, _ time.Time) ([]string, error) {
	f.userWide = append(f.userWide, userID)
	return []string{"family-1", "family-9"}, nil
}

type fakeAccessRevoker struct{ until map[string]time.Time }

func (f *fakeAccessRevoker) Revoke(familyID string, until time.Time) {
	f.until[familyID] = until
}

type refreshFixture struct {
	sessions *fakeRotator
	tokens   *fakeTokens
	access   *fakeAccessRevoker
	tx       *inlineTx
	useCase  *service.RefreshSession
}

func storedSession(rotatedAt time.Time) domain.StoredSession {
	return domain.StoredSession{
		SessionIdentity: domain.SessionIdentity{ID: "session-1", UserID: knownUserID, FamilyID: "family-1"},
		ExpiresAt:       testNow().Add(domain.RefreshTokenLifetime),
		RotatedAt:       rotatedAt,
	}
}

func newRefreshFixture(stored ...domain.StoredSession) refreshFixture {
	fixture := refreshFixture{
		sessions: &fakeRotator{byHash: map[string]domain.StoredSession{}},
		tokens:   &fakeTokens{},
		access:   &fakeAccessRevoker{until: map[string]time.Time{}},
		tx:       &inlineTx{},
	}
	for _, session := range stored {
		fixture.sessions.byHash[string(fixture.tokens.HashRefreshToken(presentedCookie))] = session
	}
	fixture.useCase = service.NewRefreshSession(service.RefreshDeps{
		Sessions: fixture.sessions,
		Tokens:   fixture.tokens,
		Access:   fixture.access,
		Tx:       fixture.tx,
		Clock:    fixedClock{now: testNow()},
		IDs:      &sequenceIDs{},
	})
	return fixture
}

func TestRefresh_UnknownOrMissingToken_IsSessionExpired(t *testing.T) {
	for _, presented := range []string{"", "family-1.unknown"} {
		_, err := newRefreshFixture().useCase.Execute(t.Context(), presented)

		if errorCode(err) != apperror.CodeSessionExpired {
			t.Errorf("%q: err = %v", presented, err)
		}
	}
}

func TestRefresh_RSEC17_RotatesIntoANewSessionOfTheSameFamily(t *testing.T) {
	fixture := newRefreshFixture(storedSession(time.Time{}))

	tokens, err := fixture.useCase.Execute(t.Context(), presentedCookie)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	successor := fixture.sessions.created[0]
	if !slices.Equal(fixture.sessions.rotated, []string{"session-1"}) || successor.FamilyID != "family-1" ||
		string(successor.TokenHash) != "hash-of-"+presentedCookie+"+next" {
		t.Fatalf("rotated = %v, successor = %+v", fixture.sessions.rotated, successor)
	}
	if tokens.Refresh != presentedCookie+"+next" || tokens.Access != "signed-access" ||
		!tokens.RefreshExpires.Equal(testNow().Add(domain.RefreshTokenLifetime)) {
		t.Fatalf("tokens = %+v", tokens)
	}
}

func TestRefresh_RSEC18_ReplayWithinGraceReturnsTheSamePairWithoutWriting(t *testing.T) {
	rotatedAt := testNow().Add(-4 * time.Second)
	fixture := newRefreshFixture(storedSession(rotatedAt))

	tokens, err := fixture.useCase.Execute(t.Context(), presentedCookie)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	signed := fixture.tokens.signed[0]
	if tokens.Refresh != presentedCookie+"+next" || !signed.IssuedAt.Equal(rotatedAt) ||
		len(fixture.sessions.created) != 0 || len(fixture.sessions.rotated) != 0 {
		t.Fatalf("tokens = %+v, signed = %+v, created = %d", tokens, signed, len(fixture.sessions.created))
	}
}

func TestRefresh_RSEC17_ReuseAfterGraceRevokesTheWholeFamily(t *testing.T) {
	fixture := newRefreshFixture(storedSession(testNow().Add(-time.Minute)))

	_, err := fixture.useCase.Execute(t.Context(), presentedCookie)

	if errorCode(err) != apperror.CodeSessionRevoked || !slices.Equal(fixture.sessions.revoked, []string{"family-1"}) {
		t.Fatalf("err = %v, revoked = %v", err, fixture.sessions.revoked)
	}
	if until := fixture.access.until["family-1"]; !until.Equal(testNow().Add(domain.AccessTokenLifetime)) {
		t.Fatalf("access revoked until %v", until)
	}
}

func TestRefresh_ExpiredSession_IsRejectedWithoutRotation(t *testing.T) {
	expired := storedSession(time.Time{})
	expired.ExpiresAt = testNow()
	fixture := newRefreshFixture(expired)

	_, err := fixture.useCase.Execute(t.Context(), presentedCookie)

	if errorCode(err) != apperror.CodeSessionExpired || len(fixture.sessions.created) != 0 {
		t.Fatalf("err = %v, created = %d", err, len(fixture.sessions.created))
	}
}
