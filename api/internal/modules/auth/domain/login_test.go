package domain

import (
	"testing"
	"time"
)

func loginNow() time.Time {
	return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
}

func failTimes(count int) LoginAttempts {
	attempts := LoginAttempts{}
	for range count {
		attempts = attempts.AfterFailure(loginNow())
	}
	return attempts
}

func TestLoginAttempts_BRAUTH01_FourFailuresDoNotLock(t *testing.T) {
	attempts := failTimes(4)

	if attempts.IsLockedAt(loginNow()) || attempts.FailedCount != 4 {
		t.Fatalf("attempts = %+v", attempts)
	}
}

func TestLoginAttempts_BRAUTH01_FifthFailureLocksForFifteenMinutes(t *testing.T) {
	attempts := failTimes(5)

	if !attempts.IsLockedAt(loginNow()) || !attempts.LockedUntil.Equal(loginNow().Add(15*time.Minute)) {
		t.Fatalf("attempts = %+v", attempts)
	}
}

func TestLoginAttempts_BRAUTH01_LockEndsExactlyAfterFifteenMinutes(t *testing.T) {
	attempts := failTimes(5)

	stillLocked := attempts.IsLockedAt(loginNow().Add(15*time.Minute - time.Second))
	released := attempts.IsLockedAt(loginNow().Add(15 * time.Minute))

	if !stillLocked || released {
		t.Fatalf("still locked = %v, released = %v", stillLocked, released)
	}
}

func TestLoginAttempts_BRAUTH01_CountStartsOverAfterTheLock(t *testing.T) {
	afterLock := failTimes(5).AfterFailure(loginNow().Add(16 * time.Minute))

	if afterLock.FailedCount != 1 || afterLock.IsLockedAt(loginNow().Add(16*time.Minute)) {
		t.Fatalf("attempts = %+v", afterLock)
	}
}

func TestLoginAttempts_BRAUTH01_SuccessClearsTheStreak(t *testing.T) {
	attempts := failTimes(4).AfterSuccess().AfterFailure(loginNow())

	if attempts.FailedCount != 1 {
		t.Fatalf("attempts = %+v", attempts)
	}
}

func TestLoginAttempts_RetryAfterRoundsUpToWholeSeconds(t *testing.T) {
	attempts := failTimes(5)

	if got := attempts.SecondsUntilUnlock(loginNow().Add(500 * time.Millisecond)); got != 900 {
		t.Fatalf("seconds = %d", got)
	}
}

func TestNewSession_RSEC17_RefreshTokenLivesThirtyDays(t *testing.T) {
	session := NewSession(SessionIdentity{ID: "s1", UserID: "u1", FamilyID: "f1"}, []byte("hash"), loginNow())

	if !session.ExpiresAt.Equal(loginNow().Add(30*24*time.Hour)) || session.FamilyID != "f1" {
		t.Fatalf("session = %+v", session)
	}
}

func TestNewAccessClaims_RSEC16_ExpireAfterFifteenMinutes(t *testing.T) {
	claims := NewAccessClaims("u1", "f1", loginNow())

	expires := loginNow().Add(15 * time.Minute)
	if claims.Subject != "u1" || claims.SessionFamily != "f1" || !claims.ExpiresAt.Equal(expires) {
		t.Fatalf("claims = %+v", claims)
	}
}
