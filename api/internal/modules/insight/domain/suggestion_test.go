package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

type vectorDaySales struct {
	Sold    int  `json:"sold"`
	SoldOut bool `json:"sold_out"`
}

type suggestionVector struct {
	Name             string           `json:"name"`
	Rhythm           string           `json:"rhythm"`
	SameWeekdaySales []vectorDaySales `json:"same_weekday_sales"`
	SoldLast14Days   int              `json:"sold_last_14_days"`
	DaysSinceFirst   int              `json:"days_since_first_drop"`
	StockUnits       int              `json:"stock_units"`
	ShelfLifeDays    int              `json:"shelf_life_days"`
	Qty              int              `json:"qty"`
	HasEnoughData    bool             `json:"has_enough_data"`
}

func loadSuggestionVectors(t *testing.T) []suggestionVector {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "testdata", "vectors", "drop_suggestion.json"))
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}
	var file struct {
		Suggestions []suggestionVector `json:"suggestions"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("decode vectors: %v", err)
	}
	return file.Suggestions
}

func (v suggestionVector) history() DropHistory {
	shelfLife := stock.UnknownShelfLife()
	if v.ShelfLifeDays > 0 {
		shelfLife = stock.ShelfLifeOf(v.ShelfLifeDays)
	}
	sales := make([]DaySales, 0, len(v.SameWeekdaySales))
	for _, day := range v.SameWeekdaySales {
		sales = append(sales, DaySales(day))
	}
	return DropHistory{
		Rhythm:           terms.Rhythm(v.Rhythm),
		SameWeekdaySales: sales,
		Sales:            SalesHistory{SoldLast14Days: v.SoldLast14Days, FirstDropAt: daysFromToday(-v.DaysSinceFirst)},
		StockUnits:       v.StockUnits,
		ShelfLife:        shelfLife,
	}
}

func TestSuggestDropQty_INS09_SharedVectors(t *testing.T) {
	for _, vector := range loadSuggestionVectors(t) {
		t.Run(vector.Name, func(t *testing.T) {
			qty, hasEnoughData := SuggestDropQty(vector.history(), today())

			if qty != vector.Qty || hasEnoughData != vector.HasEnoughData {
				t.Fatalf("suggestion = %d, %v, want %d, %v", qty, hasEnoughData, vector.Qty, vector.HasEnoughData)
			}
		})
	}
}
