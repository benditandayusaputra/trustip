package repository_test

import (
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/contact"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const storedHash = "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"

func sampleUser(t *testing.T, id, email string) domain.User {
	t.Helper()
	whatsApp, err := contact.ParseWhatsAppNumber("081234567890")
	if err != nil {
		t.Fatalf("whatsapp: %v", err)
	}
	return domain.User{
		ID:                id,
		Email:             email,
		FullName:          "Sari Wulandari",
		WhatsApp:          whatsApp,
		Password:          domain.StoredPassword{Hash: storedHash, PepperVersion: "v1"},
		PrivacyAcceptedAt: time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC),
	}
}

func TestCreateUser_StoresEveryField(t *testing.T) {
	pool := dbtest.MigratedPool(t)
	user := sampleUser(t, "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e", "sari@contoh.id")

	if err := repository.NewUsers(pool).CreateUser(t.Context(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	row, err := sqlc.New(pool.Querier(t.Context())).FindUserByEmail(t.Context(), "sari@contoh.id")
	if err != nil {
		t.Fatalf("find user: %v", err)
	}
	if row.ID != user.ID || row.Whatsapp != "6281234567890" || row.PasswordHash != user.Password.Hash ||
		row.PepperVersion != "v1" || !row.PrivacyAcceptedAt.Equal(user.PrivacyAcceptedAt) || row.FailedLoginCount != 0 {
		t.Fatalf("row = %+v", row)
	}
}

func TestCreateUser_SameEmailTwice_IsEmailTaken(t *testing.T) {
	pool := dbtest.MigratedPool(t)
	users := repository.NewUsers(pool)
	first := sampleUser(t, "0192f0c4-7b1e-7c3a-9d2e-000000000001", "sari@contoh.id")
	if err := users.CreateUser(t.Context(), first); err != nil {
		t.Fatalf("first user: %v", err)
	}

	err := users.CreateUser(t.Context(), sampleUser(t, "0192f0c4-7b1e-7c3a-9d2e-000000000002", "sari@contoh.id"))

	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateUser_UppercaseEmail_IsRejectedByTheDatabase(t *testing.T) {
	pool := dbtest.MigratedPool(t)

	mixedCase := sampleUser(t, "0192f0c4-7b1e-7c3a-9d2e-000000000003", "Sari@Contoh.ID")

	err := repository.NewUsers(pool).CreateUser(t.Context(), mixedCase)

	if err == nil || errors.Is(err, domain.ErrEmailTaken) {
		t.Fatalf("err = %v, want a check constraint failure", err)
	}
}
