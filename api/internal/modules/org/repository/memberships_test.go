package repository_test

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

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
