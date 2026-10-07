package domain

import (
	"testing"
	"time"
)

func activeSession() StoredSession {
	return StoredSession{
		SessionIdentity: SessionIdentity{ID: "s1", UserID: "u1", FamilyID: "f1"},
		ExpiresAt:       loginNow().Add(RefreshTokenLifetime),
	}
}

func TestDecideRefresh_RSEC17_FreshTokenIsRotated(t *testing.T) {
	if got := DecideRefresh(activeSession(), loginNow()); got != RefreshRotate {
		t.Fatalf("decision = %v", got)
	}
}

func TestDecideRefresh_RSEC18_ReplayWithinTenSecondsGetsTheSamePair(t *testing.T) {
	session := activeSession()
	session.RotatedAt = loginNow()

	atEdge := DecideRefresh(session, loginNow().Add(10*time.Second))
	pastEdge := DecideRefresh(session, loginNow().Add(10*time.Second+time.Millisecond))

	if atEdge != RefreshReplay || pastEdge != RefreshReuseDetected {
		t.Fatalf("at edge = %v, past edge = %v", atEdge, pastEdge)
	}
}

func TestDecideRefresh_RSEC17_RevokedOrExpiredTokensAreRejected(t *testing.T) {
	revoked := activeSession()
	revoked.RevokedAt = loginNow().Add(-time.Minute)
	expired := activeSession()
	expired.ExpiresAt = loginNow()

	if DecideRefresh(revoked, loginNow()) != RefreshRejected || DecideRefresh(expired, loginNow()) != RefreshRejected {
		t.Fatal("revoked and expired sessions must be rejected")
	}
}

func TestRotate_RSEC14_SuccessorSlidesTheThirtyDayWindow(t *testing.T) {
	later := loginNow().Add(20 * 24 * time.Hour)

	successor := activeSession().Rotate("s2", []byte("next"), later)

	if successor.ID != "s2" || successor.FamilyID != "f1" || successor.UserID != "u1" ||
		!successor.ExpiresAt.Equal(later.Add(RefreshTokenLifetime)) {
		t.Fatalf("successor = %+v", successor)
	}
}
