package domain

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
)

func day(month time.Month, date int) time.Time {
	return time.Date(2026, month, date, 0, 0, 0, 0, time.UTC)
}

func aProduct() Product {
	return Product{
		ID: "bread", Name: "Roti Manis", ShelfLife: stock.ShelfLifeOf(30),
		Batches: []Batch{
			{ID: "b1", Expiry: stock.ExpiresOn(day(11, 12)), ReceivedAt: day(11, 1).Add(2 * time.Hour), Balance: 5},
			{ID: "b2", Expiry: stock.ExpiresOn(day(12, 20)), ReceivedAt: day(11, 8), Balance: 12},
		},
	}
}

func TestViewOn_FRSTK01_BalanceAgeAndFlagPerBatch(t *testing.T) {
	jakarta := time.FixedZone("WIB", 7*60*60)

	view := aProduct().ViewOn(time.Date(2026, 11, 10, 20, 0, 0, 0, jakarta))

	if view.Balance != 17 || len(view.BatchViews) != 2 {
		t.Fatalf("view = %+v", view)
	}
	first, second := view.BatchViews[0], view.BatchViews[1]
	if first.AgeDays != 9 || first.Flag != stock.NearExpiry || second.AgeDays != 2 || second.Flag != stock.Safe {
		t.Fatalf("batches = %+v", view.BatchViews)
	}
}

func TestViewOn_FRSTK05_ProductCarriesItsWorstBatchFlag(t *testing.T) {
	product := aProduct()
	product.Batches[1].Expiry = stock.ExpiresOn(day(11, 9))

	view := product.ViewOn(day(11, 10))

	if view.Flag != stock.Expired {
		t.Fatalf("flag = %s, want EXPIRED", view.Flag)
	}
}

func TestViewOn_ProductWithoutExpiryIsSafe(t *testing.T) {
	product := Product{ID: "sambal", ShelfLife: stock.UnknownShelfLife(), Batches: []Batch{
		{ID: "s1", Expiry: stock.NoExpiry(), ReceivedAt: day(11, 1), Balance: 3},
	}}

	view := product.ViewOn(day(11, 10))

	if view.Flag != stock.Safe || view.NearestExpiry != "" {
		t.Fatalf("view = %+v", view)
	}
}

func TestViewOn_NearestExpiryIsTheEarliestDate(t *testing.T) {
	view := aProduct().ViewOn(day(11, 10))

	if view.NearestExpiry != "2026-11-12" {
		t.Fatalf("nearest = %q", view.NearestExpiry)
	}
}
