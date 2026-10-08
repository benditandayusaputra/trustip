package domain

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
)

func TestNewConsignorSearch_FRDIR07_FiltersAndBoundaries(t *testing.T) {
	search, err := NewConsignorSearch(ConsignorSearchForm{
		MinPrice: "0", MaxPrice: "100000000", MinShelfLifeDays: "730", HasHalal: true, DeliversToViewer: true,
	})

	if err != nil || !search.Filters.MaxPrice.IsGiven || search.Filters.MinShelfLifeDays.Value != 730 ||
		!search.Filters.HasHalal || !search.Filters.DeliversToViewer {
		t.Fatalf("search = %+v, err = %v", search.Filters, err)
	}
}

func TestNewConsignorSearch_FRDIR07_OutOfRangeBoundsAreRejectedTogether(t *testing.T) {
	_, err := NewConsignorSearch(ConsignorSearchForm{
		Search:   SearchForm{Radius: "7"},
		MinPrice: "-1", MaxPrice: "100000001", MinShelfLifeDays: "0",
	})

	codes := fieldCodes(t, err)
	if codes["min_price"] != "OUT_OF_RANGE" || codes["max_price"] != "OUT_OF_RANGE" ||
		codes["min_shelf_life_days"] != "OUT_OF_RANGE" || codes["radius_km"] != "UNKNOWN_CHOICE" {
		t.Fatalf("codes = %v", codes)
	}
}

func TestCheckViewer_FRDIR07_DeliveryFilterNeedsTheStoreLocation(t *testing.T) {
	search, _ := NewConsignorSearch(ConsignorSearchForm{DeliversToViewer: true})

	if fieldCodes(t, search.CheckViewer(Viewer{}))["location"] != "REQUIRED" {
		t.Fatal("delivery filter accepted without a location")
	}
	if err := search.CheckViewer(Viewer{HasLocation: true}); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func storeViewer() Viewer {
	return Viewer{Type: actor.Store, Store: StoreFacts{
		StoreType: "CANTEEN", AcceptedCategoryIDs: []string{snacksID}, NeedCategoryIDs: []string{},
	}}
}

func consignorTraits() ConsignorTraits {
	return ConsignorTraits{
		DeliveryRadiusKm: 10, PreferredStoreTypes: []string{"CANTEEN"},
		Reputation:   ConsignorReputation{ActivePartners: 3},
		LastActiveAt: sentAt().Add(-24 * time.Hour),
	}
}

func TestMatchConsignor_FRDIR10_StoreSeesHowWellTheConsignorFits(t *testing.T) {
	card := ConsignorCard{
		ID: storeA, Distance: KnownDistance(5000), Traits: consignorTraits(),
		Categories: []Category{{ID: snacksID, Name: "Kue kering dan camilan"}},
	}

	matched := storeViewer().MatchConsignor(card, sentAt())
	unmatched := Viewer{Type: actor.Consignor}.MatchConsignor(card, sentAt())

	if !matched.HasMatch || matched.Match.Score != 20+20+10+10+10 || unmatched.HasMatch {
		t.Fatalf("matched = %+v, unmatched = %v", matched.Match, unmatched.HasMatch)
	}
}

func TestMatchProduct_FRDIR10_ProductCategoryDecidesTheCategoryPart(t *testing.T) {
	card := ProductCard{
		ID: storeB, Distance: KnownDistance(10000), Traits: consignorTraits(),
		Category: Category{ID: "other", Name: "Minuman"},
	}

	matched := storeViewer().MatchProduct(card, sentAt())

	if !matched.HasMatch || matched.Match.Score != 0+0+10+10+10 {
		t.Fatalf("match = %+v", matched.Match)
	}
}

func TestRankByMatch_WorksForProductCards(t *testing.T) {
	ranked := RankByMatch([]ProductCard{
		{ID: storeA, Match: Match{Score: 10}}, {ID: storeB, Match: Match{Score: 90}},
	})

	if ranked[0].ID != storeB || ranked[0].SortPosition().ID != "" {
		t.Fatalf("ranked = %+v", ranked)
	}
}
