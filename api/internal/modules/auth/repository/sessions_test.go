package repository_test

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func sessionAt(id, familyID string, fill byte) domain.Session {
	identity := domain.SessionIdentity{ID: id, UserID: savedUserID, FamilyID: familyID}
	return domain.NewSession(identity, bytes.Repeat([]byte{fill}, 32), time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC))
}

func revokedAt() time.Time {
	return time.Date(2026, 10, 19, 9, 30, 0, 0, time.UTC)
}

func storeSessions(t *testing.T, pool *database.Pool, sessions ...domain.Session) *repository.Sessions {
	t.Helper()
	store := repository.NewSessions(pool)
	for _, session := range sessions {
		if err := store.CreateSession(t.Context(), session); err != nil {
			t.Fatalf("create session: %v", err)
		}
	}
	return store
}

func lockedSession(t *testing.T, pool *database.Pool, fill byte) domain.StoredSession {
	t.Helper()
	store := repository.NewSessions(pool)
	var stored domain.StoredSession
	err := pool.WithinTx(t.Context(), func(ctx context.Context) error {
		found, ok, err := store.LockSession(ctx, bytes.Repeat([]byte{fill}, 32))
		if err != nil || !ok {
			return fmt.Errorf("lock session: found %v: %w", ok, err)
		}
		stored = found
		return nil
	})
	if err != nil {
		t.Fatalf("lock: %v", err)
	}
	return stored
}

func TestLockSession_UnknownHash_IsNotFound(t *testing.T) {
	pool, _ := withSavedUser(t)

	_, found, err := repository.NewSessions(pool).LockSession(t.Context(), bytes.Repeat([]byte{9}, 32))

	if err != nil || found {
		t.Fatalf("found = %v, err = %v", found, err)
	}
}

func TestMarkRotated_IsReadBackOnTheNextLock(t *testing.T) {
	pool, _ := withSavedUser(t)
	only := sessionAt("0192f0c4-7b1e-7c3a-9d2e-000000000021", "0192f0c4-7b1e-7c3a-9d2e-0000000000f1", 1)
	store := storeSessions(t, pool, only)
	rotatedAt := time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)

	if err := store.MarkRotated(t.Context(), "0192f0c4-7b1e-7c3a-9d2e-000000000021", rotatedAt); err != nil {
		t.Fatalf("mark rotated: %v", err)
	}

	if got := lockedSession(t, pool, 1); !got.RotatedAt.Equal(rotatedAt) || !got.RevokedAt.IsZero() {
		t.Fatalf("session = %+v", got)
	}
}

func TestRevokeFamily_RSEC17_RevokesEverySessionOfThatFamilyOnly(t *testing.T) {
	pool, _ := withSavedUser(t)
	store := storeSessions(t, pool,
		sessionAt("0192f0c4-7b1e-7c3a-9d2e-000000000031", "0192f0c4-7b1e-7c3a-9d2e-0000000000f1", 1),
		sessionAt("0192f0c4-7b1e-7c3a-9d2e-000000000032", "0192f0c4-7b1e-7c3a-9d2e-0000000000f1", 2),
		sessionAt("0192f0c4-7b1e-7c3a-9d2e-000000000033", "0192f0c4-7b1e-7c3a-9d2e-0000000000f2", 3),
	)

	if err := store.RevokeFamily(t.Context(), "0192f0c4-7b1e-7c3a-9d2e-0000000000f1", revokedAt()); err != nil {
		t.Fatalf("revoke family: %v", err)
	}

	if lockedSession(t, pool, 1).RevokedAt.IsZero() || lockedSession(t, pool, 2).RevokedAt.IsZero() {
		t.Fatal("both sessions of the family must be revoked")
	}
	if !lockedSession(t, pool, 3).RevokedAt.IsZero() {
		t.Fatal("another family must stay active")
	}
}

func TestRevokeUser_RSEC19_ReturnsEachRevokedFamilyOnce(t *testing.T) {
	pool, _ := withSavedUser(t)
	store := storeSessions(t, pool,
		sessionAt("0192f0c4-7b1e-7c3a-9d2e-000000000041", "0192f0c4-7b1e-7c3a-9d2e-0000000000f1", 1),
		sessionAt("0192f0c4-7b1e-7c3a-9d2e-000000000042", "0192f0c4-7b1e-7c3a-9d2e-0000000000f1", 2),
		sessionAt("0192f0c4-7b1e-7c3a-9d2e-000000000043", "0192f0c4-7b1e-7c3a-9d2e-0000000000f2", 3),
	)

	families, err := store.RevokeUser(t.Context(), savedUserID, revokedAt())

	want := []string{"0192f0c4-7b1e-7c3a-9d2e-0000000000f1", "0192f0c4-7b1e-7c3a-9d2e-0000000000f2"}
	if err != nil || !slices.Equal(families, want) || lockedSession(t, pool, 3).RevokedAt.IsZero() {
		t.Fatalf("families = %v, err = %v", families, err)
	}
}
