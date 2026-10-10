package repository

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
)

func TestStoreCardMapsPublicCoordinates(t *testing.T) {
	row := sqlc.ListDirectoryStoresRow{HasLocation: true, Latitude: -6.99, Longitude: 110.42}
	card := toStoreCard(row)
	if !card.Location.IsKnown ||
		card.Location.Point.Latitude != row.Latitude || card.Location.Point.Longitude != row.Longitude {
		t.Fatalf("mapped location = %#v", card.Location)
	}
	row.HasLocation = false
	if toStoreCard(row).Location.IsKnown {
		t.Fatal("missing public location must stay unknown")
	}
}
