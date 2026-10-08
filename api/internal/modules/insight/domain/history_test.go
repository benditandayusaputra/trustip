package domain

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

func TestNewHistoryQuery_INS09_ReadsFourteenJakartaDaysAndTheWeekday(t *testing.T) {
	lateTuesdayUTC := time.Date(2026, 11, 10, 18, 0, 0, 0, time.UTC)

	got := NewHistoryQuery("org-1", "partnership-1", lateTuesdayUTC)

	wantSince := time.Date(2026, 10, 29, 0, 0, 0, 0, period.Jakarta())
	if !got.SalesSince.Equal(wantSince) || got.ISOWeekday != 3 || got.MaxOccurrences != 4 {
		t.Fatalf("query = %+v, want sales since %s on ISO weekday 3 (Wednesday)", got, wantSince)
	}
}

func TestNewHistoryQuery_SundayIsISOWeekdaySeven(t *testing.T) {
	sunday := time.Date(2026, 11, 15, 9, 0, 0, 0, period.Jakarta())

	if got := NewHistoryQuery("org-1", "partnership-1", sunday).ISOWeekday; got != 7 {
		t.Fatalf("weekday = %d, want 7", got)
	}
}

func TestSuggestDrops_INS09_OneSuggestionPerProduct(t *testing.T) {
	steady := []DaySales{{Sold: 17}, {Sold: 17}, {Sold: 18}, {Sold: 16}}
	products := []ProductHistory{
		{ProductID: "roti", History: DropHistory{Rhythm: terms.RhythmDaily, SameWeekdaySales: steady}},
		{ProductID: "keripik", History: DropHistory{Rhythm: terms.RhythmDaily}},
	}

	got := SuggestDrops(products, today())

	want := []DropSuggestion{{ProductID: "roti", Qty: 19, HasEnoughData: true}, {ProductID: "keripik"}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("suggestions = %+v, want %+v", got, want)
	}
}
