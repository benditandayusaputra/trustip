package repository_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const savedUserID = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"

func withSavedUser(t *testing.T) (*database.Pool, *repository.Users) {
	t.Helper()
	pool := dbtest.MigratedPool(t)
	users := repository.NewUsers(pool)
	if err := users.CreateUser(t.Context(), sampleUser(t, savedUserID, "sari@contoh.id")); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return pool, users
}

func TestFindCredentials_UnknownEmail_IsNotFoundWithoutError(t *testing.T) {
	_, users := withSavedUser(t)

	_, found, err := users.FindCredentials(t.Context(), "asing@contoh.id")

	if err != nil || found {
		t.Fatalf("found = %v, err = %v", found, err)
	}
}

func TestFindCredentials_KnownEmail_CarriesHashAndAttempts(t *testing.T) {
	_, users := withSavedUser(t)

	credentials, found, err := users.FindCredentials(t.Context(), "sari@contoh.id")

	if err != nil || !found || credentials.UserID != savedUserID || credentials.Password.Hash != storedHash ||
		credentials.Attempts != (domain.LoginAttempts{}) {
		t.Fatalf("credentials = %+v, found = %v, err = %v", credentials, found, err)
	}
}

func TestSaveLoginAttempts_LockUntilIsReadBack(t *testing.T) {
	pool, users := withSavedUser(t)
	lockedUntil := time.Date(2026, 10, 19, 8, 15, 0, 0, time.UTC)

	err := pool.WithinTx(t.Context(), func(ctx context.Context) error {
		if _, err := users.LockLoginAttempts(ctx, savedUserID); err != nil {
			return fmt.Errorf("lock attempts: %w", err)
		}
		return users.SaveLoginAttempts(ctx, savedUserID, domain.LoginAttempts{LockedUntil: lockedUntil})
	})
	if err != nil {
		t.Fatalf("save attempts: %v", err)
	}

	credentials, _, err := users.FindCredentials(t.Context(), "sari@contoh.id")
	if err != nil || !credentials.Attempts.LockedUntil.Equal(lockedUntil) || credentials.Attempts.FailedCount != 0 {
		t.Fatalf("attempts = %+v, err = %v", credentials.Attempts, err)
	}
}

func TestSavePassword_ReplacesHashAndPepperVersion(t *testing.T) {
	_, users := withSavedUser(t)
	rehashed := domain.StoredPassword{Hash: "$argon2id$v=19$m=20480,t=3,p=1$c2FsdA$aGFzaA", PepperVersion: "v2"}

	if err := users.SavePassword(t.Context(), savedUserID, rehashed); err != nil {
		t.Fatalf("save password: %v", err)
	}

	credentials, _, err := users.FindCredentials(t.Context(), "sari@contoh.id")
	if err != nil || credentials.Password != rehashed {
		t.Fatalf("password = %+v, err = %v", credentials.Password, err)
	}
}

func TestCreateSession_RSEC17_StoresOnlyTheTokenHash(t *testing.T) {
	pool, _ := withSavedUser(t)
	now := time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
	identity := domain.SessionIdentity{
		ID:       "0192f0c4-7b1e-7c3a-9d2e-000000000011",
		UserID:   savedUserID,
		FamilyID: "0192f0c4-7b1e-7c3a-9d2e-000000000012",
	}
	sessions := repository.NewSessions(pool)

	first := sessions.CreateSession(t.Context(), domain.NewSession(identity, bytes.Repeat([]byte{1}, 32), now))
	identity.ID = "0192f0c4-7b1e-7c3a-9d2e-000000000013"
	duplicateHash := sessions.CreateSession(t.Context(), domain.NewSession(identity, bytes.Repeat([]byte{1}, 32), now))

	if first != nil || duplicateHash == nil {
		t.Fatalf("first = %v, duplicate hash = %v", first, duplicateHash)
	}
}
