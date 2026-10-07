package repository_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	userID         = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"
	otherUserID    = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0f"
	organizationID = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c"
	membershipID   = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d"
	storedHash     = "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"
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
	err := w.queries.CreateOrganization(w.t.Context(), sqlc.CreateOrganizationParams{
		ID: organizationID, Type: "STORE", Name: "Warung Bu Tini", Slug: "warung-bu-tini",
		Whatsapp: "6281234567890", CreatedAt: createdAt(),
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

func TestFindMembership_MemberGetsOrganizationTypeAndRole(t *testing.T) {
	given := newWorld(t)
	given.user(userID, "sari@contoh.id")
	given.store()
	if err := given.membership(membershipID); err != nil {
		t.Fatalf("create membership: %v", err)
	}

	got, found, err := repository.NewMemberships(given.pool).FindMembership(t.Context(), userID)
	if err != nil {
		t.Fatalf("find membership: %v", err)
	}

	want := actor.Membership{OrganizationID: organizationID, OrganizationType: actor.Store, Role: actor.Staff}
	if !found || got != want {
		t.Fatalf("membership = %+v, found = %v", got, found)
	}
}

func TestFindMembership_FRAUTH05_UserWithoutOrganizationIsNotFound(t *testing.T) {
	given := newWorld(t)
	given.user(userID, "sari@contoh.id")

	_, found, err := repository.NewMemberships(given.pool).FindMembership(t.Context(), userID)

	if err != nil || found {
		t.Fatalf("found = %v, err = %v", found, err)
	}
}

func TestCreateMembership_OneOrganizationPerUser(t *testing.T) {
	given := newWorld(t)
	given.user(userID, "sari@contoh.id")
	given.store()
	if err := given.membership(membershipID); err != nil {
		t.Fatalf("first membership: %v", err)
	}

	err := given.membership("0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7e")

	if !database.IsUniqueViolation(err, "memberships_user_id_key") {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateOrganization_BRORG02_SlugIsUnique(t *testing.T) {
	given := newWorld(t)
	given.store()

	err := given.queries.CreateOrganization(t.Context(), sqlc.CreateOrganizationParams{
		ID: "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7f", Type: "CONSIGNOR", Name: "Warung Bu Tini",
		Slug: "warung-bu-tini", Whatsapp: "6281234567890", CreatedAt: createdAt(),
	})

	if !database.IsUniqueViolation(err, "organizations_slug_key") {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateOrganization_BRORG02_SlugWithCapitalsIsRejected(t *testing.T) {
	given := newWorld(t)

	err := given.queries.CreateOrganization(t.Context(), sqlc.CreateOrganizationParams{
		ID: organizationID, Type: "STORE", Name: "Warung Bu Tini", Slug: "Warung-Bu-Tini",
		Whatsapp: "6281234567890", CreatedAt: createdAt(),
	})

	if err == nil {
		t.Fatal("a slug with capitals must violate organizations_slug_format")
	}
}

func TestOtherUser_HasNoMembershipOfItsOwn(t *testing.T) {
	given := newWorld(t)
	given.user(userID, "sari@contoh.id")
	given.user(otherUserID, "tini@contoh.id")
	given.store()
	if err := given.membership(membershipID); err != nil {
		t.Fatalf("create membership: %v", err)
	}

	_, found, err := repository.NewMemberships(given.pool).FindMembership(t.Context(), otherUserID)

	if err != nil || found {
		t.Fatalf("found = %v, err = %v", found, err)
	}
}
