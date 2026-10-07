package repository_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/contact"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	userID         = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"
	otherUserID    = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0f"
	organizationID = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c"
	membershipID   = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d"
	storedHash     = "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"
	semarangCode   = "33.74"
)

func createdAt() time.Time {
	return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
}

type world struct {
	t       *testing.T
	pool    *database.Pool
	queries *sqlc.Queries
}

func newWorld(t *testing.T) world {
	t.Helper()
	pool := dbtest.MigratedPool(t)
	return world{t: t, pool: pool, queries: sqlc.New(pool.Querier(t.Context()))}
}

func (w world) user(id, email string) {
	w.t.Helper()
	err := w.queries.CreateUser(w.t.Context(), sqlc.CreateUserParams{
		ID: id, Email: email, PasswordHash: storedHash,
		PepperVersion: "v1", FullName: "Sari Wulandari", Whatsapp: "6281234567890", PrivacyAcceptedAt: createdAt(),
	})
	if err != nil {
		w.t.Fatalf("create user: %v", err)
	}
}

func (w world) store() {
	w.t.Helper()
	_, err := w.queries.CreateOrganization(w.t.Context(), sqlc.CreateOrganizationParams{
		ID: organizationID, Type: "STORE", Name: "Warung Bu Tini", Slug: "warung-bu-tini",
		Whatsapp: "6281234567890", CityCode: semarangCode, CreatedAt: createdAt(),
	})
	if err != nil {
		w.t.Fatalf("create organization: %v", err)
	}
}

func (w world) membership(id string) error {
	err := w.queries.CreateMembership(w.t.Context(), sqlc.CreateMembershipParams{
		ID: id, OrganizationID: organizationID, UserID: userID, Role: "STAFF", CreatedAt: createdAt(),
	})
	if err != nil {
		return fmt.Errorf("create membership %s: %w", id, err)
	}
	return nil
}

func founding(t *testing.T, ids domain.FoundingIDs, slug domain.Slug) domain.Founding {
	t.Helper()
	whatsApp, err := contact.ParseWhatsAppNumber("081234567890")
	if err != nil {
		t.Fatalf("whatsapp: %v", err)
	}
	onboarding := domain.Onboarding{
		Type: actor.Consignor, Name: "Kripik Mbok Darmi", WhatsApp: whatsApp, CityCode: semarangCode,
		District: "Tembalang",
	}
	return onboarding.Found(ids, slug, createdAt())
}
