package domain

import (
	"slices"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
)

func readyStore() ListingFacts {
	return ListingFacts{
		Type: actor.Store, HasName: true, HasCity: true, HasLocation: true,
		HasStoreType: true, AcceptedCategories: 1,
	}
}

func TestCompletenessOf_BRORG05_StoreNeedsNameCityLocationTypeAndACategory(t *testing.T) {
	facts := ListingFacts{Type: actor.Store, HasName: true}

	completeness := CompletenessOf(facts)

	want := []Requirement{RequireCity, RequireLocation, RequireStoreType, RequireAcceptedCategory}
	if !slices.Equal(completeness.MissingRequired(), want) || completeness.CanBeListed() {
		t.Fatalf("missing = %v", completeness.MissingRequired())
	}
}

func TestCompletenessOf_BRORG05_ConsignorNeedsAnActiveProduct(t *testing.T) {
	facts := ListingFacts{Type: actor.Consignor, HasName: true, HasCity: true, HasLocation: true}

	completeness := CompletenessOf(facts)

	if !slices.Equal(completeness.MissingRequired(), []Requirement{RequireActiveProduct}) {
		t.Fatalf("missing = %v", completeness.MissingRequired())
	}
}

func TestCompletenessOf_BRORG05_ReadyStoreCanBeListedWithoutRecommendedItems(t *testing.T) {
	completeness := CompletenessOf(readyStore())

	if !completeness.CanBeListed() || completeness.Percent() == 100 {
		t.Fatalf("completeness = %+v", completeness)
	}
}

func TestCompletenessOf_FRORG08_PercentCountsEveryItem(t *testing.T) {
	full := readyStore()
	full.HasDescription, full.HasLogo, full.Photos = true, true, 1

	if got := CompletenessOf(full).Percent(); got != 100 {
		t.Fatalf("percent = %d", got)
	}
	if got := CompletenessOf(ListingFacts{Type: actor.Store}).Percent(); got != 0 {
		t.Fatalf("empty percent = %d", got)
	}
	if got := CompletenessOf(readyStore()).Percent(); got != 62 {
		t.Fatalf("ready percent = %d", got)
	}
}

func TestCompleteness_PercentOfNoItemsIsZero(t *testing.T) {
	if got := (Completeness{}).Percent(); got != 0 {
		t.Fatalf("percent = %d", got)
	}
}
