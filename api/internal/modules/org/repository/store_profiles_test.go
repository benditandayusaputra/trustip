package repository_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/repository"
)

const (
	breadCategory = "01a1161e-0a30-752d-ab9b-7397976f0dba"
	drinkCategory = "01a1161e-0a92-70d6-9954-74cb98b0b1ac"
	firstPhotoID  = "0192f0c4-aa00-7e66-a1bb-4d2f3e5d6a91"
	secondPhotoID = "0192f0c4-aa00-7e66-a1bb-4d2f3e5d6a92"
)

func foundedStore(t *testing.T) world {
	t.Helper()
	given := newWorld(t)
	given.user(userID, "tini@contoh.id")
	store := founding(t, firstFounder(), "warung")
	store.Organization.Type = actor.Store
	if _, err := repository.NewFounders(given.pool).CreateOrganization(t.Context(), store); err != nil {
		t.Fatalf("found: %v", err)
	}
	return given
}

func storeChange(categories, photos []string) domain.StoreProfileChange {
	return domain.StoreProfileChange{OrganizationID: organizationID, At: createdAt(), Profile: domain.StoreProfile{
		StoreType: "BAKERY", CategoryIDs: categories, PhotoFileIDs: photos,
		Acceptance:   domain.Acceptance{AcceptsCommission: true},
		Commission:   domain.CommissionRange{Given: true, MinBPS: 1000, MaxBPS: 2000},
		Conditions:   domain.StoreConditions{RequiresHalal: true},
		OpeningHours: "07.00 sampai 21.00",
	}}
}

func TestUpdateStoreProfile_FRORG04_SavesProfileCategoriesAndPhotosInOrder(t *testing.T) {
	given := foundedStore(t)
	given.file(firstPhotoID, "STORE_PHOTO")
	given.file(secondPhotoID, "STORE_PHOTO")
	profiles := repository.NewProfiles(given.pool)
	photos := []string{secondPhotoID, firstPhotoID}

	change := storeChange([]string{drinkCategory, breadCategory}, photos)
	if err := profiles.UpdateStoreProfile(t.Context(), change); err != nil {
		t.Fatalf("update: %v", err)
	}

	settings, err := profiles.GetSettings(t.Context(), organizationID)
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	store := settings.Store.Profile
	if store.StoreType != "BAKERY" || store.Commission.MaxBPS != 2000 || !store.Conditions.RequiresHalal ||
		!slices.Equal(store.CategoryIDs, []string{breadCategory, drinkCategory}) ||
		!slices.Equal(store.PhotoFileIDs, photos) || len(settings.Store.Photos) != 2 {
		t.Fatalf("store = %+v", settings.Store)
	}
}

func TestUpdateStoreProfile_UnknownCategoryIsReported(t *testing.T) {
	given := foundedStore(t)

	err := repository.NewProfiles(given.pool).UpdateStoreProfile(t.Context(),
		storeChange([]string{"01a1161e-0000-7000-8000-000000000000"}, nil))

	if !errors.Is(err, domain.ErrUnknownCategory) {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateStoreProfile_PhotoMustBeAStorePhotoOfTheSameStore(t *testing.T) {
	given := foundedStore(t)
	given.file(firstPhotoID, "LOGO")

	err := repository.NewProfiles(given.pool).UpdateStoreProfile(t.Context(), storeChange(nil, []string{firstPhotoID}))

	if !errors.Is(err, domain.ErrUnknownPhoto) {
		t.Fatalf("err = %v", err)
	}
}

func TestListCategories_TwelveCategoriesInTheirOrder(t *testing.T) {
	categories, err := repository.NewRegions(newWorld(t).pool).ListCategories(t.Context())

	if err != nil || len(categories) != 12 ||
		categories[0].Slug != "roti-dan-kue-basah" || categories[11].Slug != "lainnya" {
		t.Fatalf("categories = %+v, err = %v", categories, err)
	}
}
