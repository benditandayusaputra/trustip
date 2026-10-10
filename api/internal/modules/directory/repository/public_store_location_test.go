package repository_test

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

func TestStoreDirectoryReturnsPublicInsteadOfPrivateLocation(t *testing.T) {
	scene := newWorld(t)
	publicLatitude := originLatitude + kmInLatitude
	_, err := scene.queries.UpdateOrganizationProfile(t.Context(), sqlc.UpdateOrganizationProfileParams{
		ID: nearCanteen, Name: "Kantin Teknik", Whatsapp: "6281234567890", CityCode: semarangCode,
		HasLocation: true, Latitude: originLatitude, Longitude: originLongitude,
		PublicLatitude: publicLatitude, PublicLongitude: originLongitude, UpdatedAt: base(),
	})
	scene.must(err)
	cards := scene.list(domain.StoreQuery{Listing: firstOf(1)})
	requireIDs(t, cards, nearCanteen)
	if !cards[0].Location.IsKnown || cards[0].Location.Point.Latitude != publicLatitude {
		t.Fatalf("directory exposed wrong location: %#v", cards[0].Location)
	}
	_, err = scene.queries.UpdateOrganizationProfile(t.Context(), sqlc.UpdateOrganizationProfileParams{
		ID: nearCanteen, Name: "Kantin Teknik", Whatsapp: "6281234567890", CityCode: semarangCode, UpdatedAt: base(),
	})
	scene.must(err)
	cards = scene.list(domain.StoreQuery{Listing: nearest(), Filters: domain.StoreFilters{Text: "Kantin Teknik"}})
	requireIDs(t, cards, nearCanteen)
	if cards[0].Location.IsKnown {
		t.Fatal("missing public location must remain unknown")
	}
}
