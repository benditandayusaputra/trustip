package repository_test

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/repository"
)

func TestGetListingFacts_FRORG08_ReflectsTheStoredProfile(t *testing.T) {
	given := foundedStore(t)
	given.file(firstPhotoID, "STORE_PHOTO")
	profiles := repository.NewProfiles(given.pool)
	change := storeChange([]string{breadCategory}, []string{firstPhotoID})
	if err := profiles.UpdateStoreProfile(t.Context(), change); err != nil {
		t.Fatalf("store profile: %v", err)
	}

	facts, err := profiles.GetListingFacts(t.Context(), organizationID)

	want := domain.ListingFacts{
		Type: actor.Store, HasName: true, HasCity: true, HasStoreType: true, AcceptedCategories: 1, Photos: 1,
	}
	if err != nil || facts != want {
		t.Fatalf("facts = %+v, err = %v", facts, err)
	}
}

func TestSetListed_FRORG06_ShowsUpInTheSettings(t *testing.T) {
	given := foundedStore(t)
	profiles := repository.NewProfiles(given.pool)

	if err := profiles.SetListed(t.Context(), domain.ListingChange{
		OrganizationID: organizationID, IsListed: true, At: createdAt(),
	}); err != nil {
		t.Fatalf("set listed: %v", err)
	}

	settings, err := profiles.GetSettings(t.Context(), organizationID)
	if err != nil || !settings.Listing.IsListed || len(settings.Listing.Completeness.Items) == 0 {
		t.Fatalf("listing = %+v, err = %v", settings.Listing, err)
	}
}
