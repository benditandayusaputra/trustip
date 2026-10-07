package domain

import (
	"math"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const metersPerDegree = 111_320

func tembalang() Coordinates {
	return Coordinates{Latitude: -7.0546, Longitude: 110.4386}
}

func TestPublicLocation_BRORG03_StoreShowsItsExactLocation(t *testing.T) {
	if got := PublicLocation(actor.Store, tembalang()); got != tembalang() {
		t.Fatalf("public = %+v", got)
	}
}

func TestPublicLocation_BRORG03_ConsignorIsSnappedToAFiveHundredMetreGrid(t *testing.T) {
	got := PublicLocation(actor.Consignor, tembalang())

	want := Coordinates{Latitude: -7.056, Longitude: 110.439}
	if math.Abs(got.Latitude-want.Latitude) > 1e-9 || math.Abs(got.Longitude-want.Longitude) > 1e-9 {
		t.Fatalf("public = %+v, want %+v", got, want)
	}
}

func TestPublicLocation_BRORG03_NeighboursInOneGridCellShareOnePublicPoint(t *testing.T) {
	nearby := Coordinates{Latitude: tembalang().Latitude - 0.0005, Longitude: tembalang().Longitude + 0.0005}

	if PublicLocation(actor.Consignor, tembalang()) != PublicLocation(actor.Consignor, nearby) {
		t.Fatal("two homes 70 m apart must not be told apart by their public point")
	}
}

func TestPublicLocation_BRORG03_ConsignorMovesAtMostHalfAGridCell(t *testing.T) {
	exact := Coordinates{Latitude: -6.200123, Longitude: 106.816666}

	public := PublicLocation(actor.Consignor, exact)

	halfCellMetres := publicGridDegrees / 2 * metersPerDegree
	if shift := math.Abs(public.Latitude-exact.Latitude) * metersPerDegree; shift > halfCellMetres {
		t.Fatalf("latitude moved %.0f m", shift)
	}
	if shift := math.Abs(public.Longitude-exact.Longitude) * metersPerDegree; shift > halfCellMetres {
		t.Fatalf("longitude moved %.0f m", shift)
	}
}

func TestNewCoordinates_RSEC46_RangeBoundaries(t *testing.T) {
	cases := []struct {
		latitude, longitude float64
		valid               bool
	}{
		{-90, -180, true},
		{90, 180, true},
		{-90.000001, 0, false},
		{90.000001, 0, false},
		{0, -180.000001, false},
		{0, 180.000001, false},
	}
	for _, tc := range cases {
		_, err := NewCoordinates(tc.latitude, tc.longitude)

		if (err == nil) != tc.valid {
			t.Errorf("(%v, %v): err = %v", tc.latitude, tc.longitude, err)
		}
	}
}

func TestNewOnboarding_FRORG03_LocationIsOptionalButValidatedWhenGiven(t *testing.T) {
	form := validForm()
	form.Location = LocationInput{Given: true, Latitude: 95, Longitude: 110}

	_, err := NewOnboarding(form)

	if got := fieldCodes(t, err)["location"]; got != string(apperror.FieldOutOfRange) {
		t.Fatalf("location = %q", got)
	}
}

func TestNewOnboarding_BRORG03_ConsignorLocationGetsARoundedPublicPoint(t *testing.T) {
	form := validForm()
	form.Type = string(actor.Consignor)
	form.Location = LocationInput{Given: true, Latitude: tembalang().Latitude, Longitude: tembalang().Longitude}

	onboarding, err := NewOnboarding(form)
	if err != nil {
		t.Fatalf("new onboarding: %v", err)
	}

	if !onboarding.HasLocation || onboarding.Location.Exact != tembalang() ||
		onboarding.Location.Public != PublicLocation(actor.Consignor, tembalang()) {
		t.Fatalf("location = %+v", onboarding.Location)
	}
}
