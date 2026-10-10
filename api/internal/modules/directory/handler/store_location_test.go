package handler

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

func TestStoreCardPublicLocation(t *testing.T) {
	known := domain.Place{Point: domain.Point{Latitude: -6.99, Longitude: 110.42}, IsKnown: true}
	for name, place := range map[string]domain.Place{"known": known, "unknown": {}, "zero": {IsKnown: true}} {
		t.Run(name, func(t *testing.T) {
			response := (presenter{}).storeCard(domain.StoreCard{Location: place})
			if !place.IsKnown {
				if response.Location != nil {
					t.Fatal("unknown location must be null")
				}
				return
			}
			want := placeResponse{Latitude: place.Point.Latitude, Longitude: place.Point.Longitude}
			if response.Location == nil || *response.Location != want {
				t.Fatalf("public location = %#v, want %#v", response.Location, want)
			}
		})
	}
}
