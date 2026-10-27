package repository_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/contact"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
)

const foreignOrganizationID = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c"

func profileChange(t *testing.T, avatarFileID, organizationID string) domain.AccountChange {
	t.Helper()
	whatsApp, err := contact.ParseWhatsAppNumber("0813 1111 2222")
	if err != nil {
		t.Fatalf("whatsapp: %v", err)
	}
	return domain.AccountChange{
		UserID:         savedUserID,
		OrganizationID: organizationID,
		Profile:        domain.ProfileChange{FullName: "Sari W", WhatsApp: whatsApp, AvatarFileID: avatarFileID},
		At:             revokedAt(),
	}
}

func TestUpdateProfile_SavesNameAndWhatsAppForTheAccount(t *testing.T) {
	_, users := withSavedUser(t)

	if err := users.UpdateProfile(t.Context(), profileChange(t, "", "")); err != nil {
		t.Fatalf("update profile: %v", err)
	}

	account, err := users.GetAccount(t.Context(), savedUserID)
	if err != nil || account.FullName != "Sari W" || account.WhatsApp != "6281311112222" || account.AvatarFileID != "" {
		t.Fatalf("account = %+v, err = %v", account, err)
	}
}

func TestUpdateProfile_RefusesAnAvatarThatTheUserDidNotUpload(t *testing.T) {
	_, users := withSavedUser(t)
	cases := []domain.AccountChange{
		profileChange(t, "0192f0c4-aa00-7e66-a1bb-4d2f3e5d6a99", ""),
		profileChange(t, "0192f0c4-aa00-7e66-a1bb-4d2f3e5d6a99", foreignOrganizationID),
	}
	for _, change := range cases {
		if err := users.UpdateProfile(t.Context(), change); !errors.Is(err, domain.ErrUnknownAvatar) {
			t.Fatalf("err = %v", err)
		}
	}
}

func TestLockPasswordOwner_ReturnsIdentityAndStoredHash(t *testing.T) {
	pool, users := withSavedUser(t)
	var owner domain.PasswordOwner
	err := pool.WithinTx(t.Context(), func(ctx context.Context) error {
		found, err := users.LockPasswordOwner(ctx, savedUserID)
		owner = found
		if err != nil {
			return fmt.Errorf("lock password owner: %w", err)
		}
		return nil
	})

	if err != nil || owner.Email != "sari@contoh.id" || owner.Password.Hash != storedHash {
		t.Fatalf("owner = %+v, err = %v", owner, err)
	}
}

func TestRevokeOtherSessions_KeepsTheCurrentFamilySignedIn(t *testing.T) {
	pool, _ := withSavedUser(t)
	store := storeSessions(t, pool,
		sessionAt("0192f0c4-7b1e-7c3a-9d2e-000000000051", "0192f0c4-7b1e-7c3a-9d2e-0000000000f1", 4),
		sessionAt("0192f0c4-7b1e-7c3a-9d2e-000000000052", "0192f0c4-7b1e-7c3a-9d2e-0000000000f2", 5),
		sessionAt("0192f0c4-7b1e-7c3a-9d2e-000000000053", "0192f0c4-7b1e-7c3a-9d2e-0000000000f3", 6),
	)

	families, err := store.RevokeOtherSessions(
		t.Context(), savedUserID, "0192f0c4-7b1e-7c3a-9d2e-0000000000f1", revokedAt(),
	)

	want := []string{"0192f0c4-7b1e-7c3a-9d2e-0000000000f2", "0192f0c4-7b1e-7c3a-9d2e-0000000000f3"}
	if err != nil || !slices.Equal(families, want) {
		t.Fatalf("families = %v, err = %v", families, err)
	}
	if !lockedSession(t, pool, 4).RevokedAt.IsZero() || lockedSession(t, pool, 5).RevokedAt.IsZero() {
		t.Fatal("the current family must stay signed in while the others end")
	}
}
