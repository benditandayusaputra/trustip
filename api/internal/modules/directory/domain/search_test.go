package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	snacksID = "01a1161e-0a49-71f7-a49a-3cf81042e3b0"
	storeA   = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a01"
	storeB   = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a02"
	storeC   = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a03"
)

func fieldCodes(t *testing.T, err error) map[string]string {
	t.Helper()
	appErr, ok := errors.AsType[*apperror.Error](err)
	if !ok {
		t.Fatalf("err = %v, want a validation error", err)
	}
	codes := map[string]string{}
	for _, problem := range appErr.Fields() {
		codes[problem.Field] = problem.Code
	}
	return codes
}

func TestNewStoreSearch_FRDIR03_FRDIR05_DefaultsToNearestWithinTenKilometres(t *testing.T) {
	search, err := NewStoreSearch(StoreSearchForm{})

	if err != nil || search.Sort != Nearest || search.RadiusKm != 10 || search.Origin.Kind != OwnLocation ||
		!search.UsesOwnCategories {
		t.Fatalf("search = %+v, err = %v", search, err)
	}
}

func TestNewStoreSearch_FRDIR05_RadiusChoices(t *testing.T) {
	cases := map[string]int{"2": 2, "5": 5, "10": 10, "25": 25, "all": EveryDistance}
	for raw, want := range cases {
		search, err := NewStoreSearch(StoreSearchForm{Radius: raw})
		if err != nil || search.RadiusKm != want {
			t.Errorf("radius %q = %d, err = %v", raw, search.RadiusKm, err)
		}
	}
	for _, raw := range []string{"3", "0", "-5", "jauh"} {
		_, err := NewStoreSearch(StoreSearchForm{Radius: raw})
		if fieldCodes(t, err)["radius_km"] != "UNKNOWN_CHOICE" {
			t.Errorf("radius %q accepted", raw)
		}
	}
}

func TestNewStoreSearch_FRDIR06_FiltersAreValidatedTogether(t *testing.T) {
	_, err := NewStoreSearch(StoreSearchForm{
		Sort: "CHEAPEST", StoreType: "BAR", PricingModel: "RENT", MaxCommissionBPS: "5001",
		CategoryIDs: []string{"camilan"},
	})

	codes := fieldCodes(t, err)
	if codes["sort"] != "UNKNOWN_CHOICE" || codes["store_type"] != "UNKNOWN_CHOICE" ||
		codes["pricing_model"] != "UNKNOWN_CHOICE" || codes["max_commission_bps"] != "OUT_OF_RANGE" ||
		codes["category_id"] != "UNKNOWN_CHOICE" {
		t.Fatalf("codes = %v", codes)
	}
}

func TestNewStoreSearch_FRDIR06_MaxCommissionBoundaries(t *testing.T) {
	for raw, wantOK := range map[string]bool{"0": true, "5000": true, "5001": false, "-1": false} {
		search, err := NewStoreSearch(StoreSearchForm{MaxCommissionBPS: raw})
		if (err == nil) != wantOK || (wantOK && !search.Filters.HasMaxCommission) {
			t.Errorf("max commission %q: search = %+v, err = %v", raw, search.Filters, err)
		}
	}
}

func TestNewStoreSearch_BRDIR06_AnyCategoryClearsTheDefault(t *testing.T) {
	search, err := NewStoreSearch(StoreSearchForm{CategoryIDs: []string{"any"}})

	if err != nil || search.UsesOwnCategories || len(search.Filters.CategoryIDs) != 0 {
		t.Fatalf("search = %+v, err = %v", search, err)
	}
}

func TestNewStoreSearch_FRDIR04_OriginChoices(t *testing.T) {
	here, hereErr := NewStoreSearch(StoreSearchForm{Origin: OriginForm{
		Kind: "HERE", Latitude: "-7.0", Longitude: "110.4",
	}})
	city, cityErr := NewStoreSearch(StoreSearchForm{Origin: OriginForm{Kind: "CITY", CityCode: "33.74"}})
	_, badPoint := NewStoreSearch(StoreSearchForm{Origin: OriginForm{Kind: "HERE", Latitude: "91", Longitude: "0"}})
	_, badCity := NewStoreSearch(StoreSearchForm{Origin: OriginForm{Kind: "CITY", CityCode: "semarang"}})

	if hereErr != nil || here.Origin.Point.Longitude != 110.4 || cityErr != nil || city.Origin.CityCode != "33.74" {
		t.Fatalf("here = %+v %v, city = %+v %v", here.Origin, hereErr, city.Origin, cityErr)
	}
	if fieldCodes(t, badPoint)["origin"] != "OUT_OF_RANGE" || fieldCodes(t, badCity)["city_code"] != "UNKNOWN_CHOICE" {
		t.Fatalf("badPoint = %v, badCity = %v", badPoint, badCity)
	}
}

func TestResolveFor_FRDIR04_OriginFollowsTheChoice(t *testing.T) {
	viewer := Viewer{Location: Point{-7, 110}, HasLocation: true, City: Point{-6.9, 110.4}, HasCity: true}

	own, _ := OriginRequest{Kind: OwnLocation}.ResolveFor(viewer)
	city, _ := OriginRequest{Kind: CityCentre, CityCode: "33.74"}.ResolveFor(viewer)
	here, _ := OriginRequest{Kind: CurrentPosition, Point: Point{-7.1, 110.1}}.ResolveFor(Viewer{})

	if own.Point != viewer.Location || city.Point != viewer.City || !here.IsKnown || here.Point.Latitude != -7.1 {
		t.Fatalf("own = %+v, city = %+v, here = %+v", own, city, here)
	}
}

func TestResolveFor_UnknownCityIsAFieldError(t *testing.T) {
	_, err := OriginRequest{Kind: CityCentre, CityCode: "99.99"}.ResolveFor(Viewer{})

	if fieldCodes(t, err)["city_code"] != "UNKNOWN_CHOICE" {
		t.Fatalf("err = %v", err)
	}
}

func TestSortFor_BRDIR04_WithoutOriginNearestFallsBackToCity(t *testing.T) {
	search := StoreSearch{Sort: Nearest}
	newest := StoreSearch{Sort: Newest}

	if search.SortFor(Origin{}) != ByCity || search.SortFor(Origin{IsKnown: true}) != Nearest ||
		newest.SortFor(Origin{}) != Newest {
		t.Fatal("unexpected sort")
	}
}

func TestEffectiveRadius_BRDIR03_ExpandsUntilFiveResults(t *testing.T) {
	counts := map[int]int{2: 1, 5: 4, 10: 5, 25: 9}
	cases := []struct {
		name     string
		selected int
		counts   map[int]int
		want     int
	}{
		{"enough within 2 km stays", 2, map[int]int{2: 5}, 2},
		{"four within 5 km grows to 10", 5, counts, 10},
		{"2 km skips 5 km with four", 2, counts, 10},
		{"25 km with too few becomes every distance", 25, map[int]int{25: 4}, EveryDistance},
		{"every distance stays", EveryDistance, map[int]int{}, EveryDistance},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := EffectiveRadius(tc.selected, tc.counts); got != tc.want {
				t.Fatalf("EffectiveRadius = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSearchContext_BRDIR03_ExpandedOnlyWithAnOrigin(t *testing.T) {
	expanded := SearchContext{Origin: Origin{IsKnown: true}, SelectedRadiusKm: 5, EffectiveRadiusKm: 10}
	noOrigin := SearchContext{SelectedRadiusKm: 5, EffectiveRadiusKm: EveryDistance}

	if !expanded.IsExpanded() || noOrigin.IsExpanded() {
		t.Fatal("unexpected expansion")
	}
}

func scored(id string, score, meters int) StoreCard {
	return StoreCard{ID: id, Distance: KnownDistance(meters), Match: Match{Score: score}}
}

func TestRankByMatch_FRDIR08_HighestScoreThenNearest(t *testing.T) {
	ranked := RankByMatch([]StoreCard{scored(storeA, 60, 900), scored(storeB, 80, 3000), scored(storeC, 60, 400)})

	if ranked[0].ID != storeB || ranked[1].ID != storeC || ranked[2].ID != storeA {
		t.Fatalf("ranked = %v %v %v", ranked[0].ID, ranked[1].ID, ranked[2].ID)
	}
}

func TestPageAfter_BRDIR05_ContinuesAfterTheLastCard(t *testing.T) {
	ranked := []StoreCard{scored(storeA, 3, 0), scored(storeB, 2, 0), scored(storeC, 1, 0)}

	first, firstMore := PageAfter(ranked, "", 2)
	second, secondMore := PageAfter(ranked, storeB, 2)

	if len(first) != 2 || !firstMore || len(second) != 1 || second[0].ID != storeC || secondMore {
		t.Fatalf("first = %d %v, second = %d %v", len(first), firstMore, len(second), secondMore)
	}
}

func TestCursor_RoundTripsThePositionAndRadius(t *testing.T) {
	position := Position{Text: "Kota Semarang", Number: 1234.5, ID: storeA, RadiusKm: 25}

	parsed, err := ParseCursor(CursorOf(position))

	if err != nil || parsed != position {
		t.Fatalf("parsed = %+v, err = %v", parsed, err)
	}
}

func TestParseCursor_TamperedCursorIsAFieldError(t *testing.T) {
	tampered := []string{"!!!", "bm90LWpzb24", CursorOf(Position{ID: "x"}), CursorOf(Position{ID: storeA, RadiusKm: 7})}
	for _, raw := range tampered {
		if fieldCodes(t, mustFail(ParseCursor(raw)))["cursor"] != "UNKNOWN_CHOICE" {
			t.Errorf("%q accepted", raw)
		}
	}
}

func TestMatchStore_FRDIR10_OnlyConsignorsSeeAScore(t *testing.T) {
	card := StoreCard{
		ID: storeA, Distance: KnownDistance(5000), StoreType: "CANTEEN",
		AcceptedCategories: []Category{{ID: snacksID, Name: "Kue kering dan camilan"}},
	}
	consignor := Viewer{Type: actor.Consignor, Consignor: ConsignorFacts{
		Categories: []Category{{ID: snacksID, Name: "Kue kering dan camilan"}},
	}}

	matched := consignor.MatchStore(card, sentAt())
	unmatched := Viewer{Type: actor.Store}.MatchStore(card, sentAt())

	if !matched.HasMatch || matched.Match.Score != 32+20+5+5 || unmatched.HasMatch {
		t.Fatalf("matched = %+v, unmatched = %+v", matched.Match, unmatched.HasMatch)
	}
}

func TestViewer_BRDIR06_DefaultCategoriesAreTheConsignorsProductCategories(t *testing.T) {
	viewer := Viewer{Consignor: ConsignorFacts{Categories: []Category{{ID: snacksID}}}}

	if ids := viewer.DefaultCategoryIDs(); len(ids) != 1 || ids[0] != snacksID {
		t.Fatalf("ids = %v", ids)
	}
}

func mustFail[T any](_ T, err error) error {
	return err
}

func sentAt() time.Time {
	return time.Date(2026, 10, 20, 3, 0, 0, 0, time.UTC)
}
