package domain

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
)

func consignorPartner(status Status) Partnership {
	return Partnership{ID: "k1", Status: status, Counterpart: Counterpart{
		ID: "c1", Type: actor.Consignor, Name: "Roti Pak Anton", Address: "Jl. Mawar 3", HasLocation: true,
		Exact:  Coordinates{Latitude: -6.99123, Longitude: 110.42345},
		Public: Coordinates{Latitude: -6.99, Longitude: 110.42},
	}}
}

func TestView_BRORG04_ActivePartnerSeesTheConsignorAddressAndExactLocation(t *testing.T) {
	view := consignorPartner(Active).View()

	if view.Counterpart.Address != "Jl. Mawar 3" || view.Location.Latitude != -6.99123 {
		t.Fatalf("view = %+v", view)
	}
}

func TestView_BRORG04_EndedPartnershipFallsBackToThePublicLocation(t *testing.T) {
	view := consignorPartner(Ended).View()

	if view.Counterpart.Address != "" || view.Location.Latitude != -6.99 {
		t.Fatalf("view = %+v", view)
	}
}

func TestView_StoreLocationIsAlwaysExact(t *testing.T) {
	partnership := consignorPartner(Ended)
	partnership.Counterpart.Type = actor.Store

	if view := partnership.View(); view.Location.Longitude != 110.42345 || view.Counterpart.Address == "" {
		t.Fatalf("view = %+v", view)
	}
}

func TestParseStatus_FRPRT05_EndedPartnershipsCanBeListed(t *testing.T) {
	for raw, valid := range map[string]bool{"": true, "ACTIVE": true, "ENDED": true, "PAUSED": false} {
		if _, err := ParseStatus(raw); (err == nil) != valid {
			t.Errorf("%q: err = %v", raw, err)
		}
	}
}
