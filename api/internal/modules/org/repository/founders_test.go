package repository_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/repository"
)

func firstFounder() domain.FoundingIDs {
	return domain.FoundingIDs{OrganizationID: organizationID, MembershipID: membershipID, FounderID: userID}
}

func secondFounder() domain.FoundingIDs {
	return domain.FoundingIDs{
		OrganizationID: "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a80",
		MembershipID:   "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a81",
		FounderID:      otherUserID,
	}
}

func TestCreateOrganization_BRAUTH02_StoresOrganizationProfileAndOwner(t *testing.T) {
	given := newWorld(t)
	given.user(userID, "darmi@contoh.id")
	founders := repository.NewFounders(given.pool)

	cityName, err := founders.CreateOrganization(t.Context(), founding(t, firstFounder(), "kripik-mbok-darmi"))
	if err != nil {
		t.Fatalf("create organization: %v", err)
	}

	membership, found, err := repository.NewMemberships(given.pool).FindMembership(t.Context(), userID)
	want := actor.Membership{OrganizationID: organizationID, OrganizationType: actor.Consignor, Role: actor.Owner}
	if err != nil || !found || membership != want || cityName != "Kota Semarang" {
		t.Fatalf("membership = %+v, city = %q, err = %v", membership, cityName, err)
	}
}

func TestCreateOrganization_UnknownCityIsReported(t *testing.T) {
	given := newWorld(t)
	given.user(userID, "darmi@contoh.id")
	unknownCity := founding(t, firstFounder(), "kripik-mbok-darmi")
	unknownCity.Organization.CityCode = "99.99"

	_, err := repository.NewFounders(given.pool).CreateOrganization(t.Context(), unknownCity)

	if !errors.Is(err, domain.ErrUnknownCity) {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateOrganization_BRORG02_TakenSlugIsReported(t *testing.T) {
	given := newWorld(t)
	given.user(userID, "darmi@contoh.id")
	given.user(otherUserID, "ani@contoh.id")
	founders := repository.NewFounders(given.pool)
	if _, err := founders.CreateOrganization(t.Context(), founding(t, firstFounder(), "kripik")); err != nil {
		t.Fatalf("first: %v", err)
	}

	_, err := founders.CreateOrganization(t.Context(), founding(t, secondFounder(), "kripik"))

	if !errors.Is(err, domain.ErrSlugTaken) {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateOrganization_SecondOrganizationOfOneUserIsReported(t *testing.T) {
	given := newWorld(t)
	given.user(userID, "darmi@contoh.id")
	founders := repository.NewFounders(given.pool)
	if _, err := founders.CreateOrganization(t.Context(), founding(t, firstFounder(), "kripik")); err != nil {
		t.Fatalf("first: %v", err)
	}
	again := secondFounder()
	again.FounderID = userID

	_, err := founders.CreateOrganization(t.Context(), founding(t, again, "kripik-2"))

	if !errors.Is(err, domain.ErrAlreadyMember) {
		t.Fatalf("err = %v", err)
	}
}

func TestTakenSlugs_BRORG02_ListsTheBaseAndItsNumberedVariantsOnly(t *testing.T) {
	given := newWorld(t)
	given.user(userID, "darmi@contoh.id")
	given.user(otherUserID, "ani@contoh.id")
	founders := repository.NewFounders(given.pool)
	slugs := map[domain.FoundingIDs]domain.Slug{firstFounder(): "kripik", secondFounder(): "kripik-pedas"}
	for ids, slug := range slugs {
		if _, err := founders.CreateOrganization(t.Context(), founding(t, ids, slug)); err != nil {
			t.Fatalf("create %s: %v", slug, err)
		}
	}

	taken, err := founders.TakenSlugs(t.Context(), "kripik")

	if err != nil || !slices.Equal(taken, []domain.Slug{"kripik"}) {
		t.Fatalf("taken = %v, err = %v", taken, err)
	}
}
