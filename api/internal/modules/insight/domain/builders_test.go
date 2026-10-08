package domain

import (
	"strconv"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

func today() time.Time {
	return time.Date(2026, 11, 10, 10, 0, 0, 0, period.Jakarta())
}

func daysFromToday(days int) time.Time {
	return today().AddDate(0, 0, days)
}

type productBuilder struct {
	product ProductStock
}

func aProduct() productBuilder {
	return productBuilder{product: ProductStock{
		PartnershipID:   "partnership-1",
		CounterpartName: "Toko Berkah",
		ProductID:       "product-roti",
		ProductName:     "Roti Sobek",
		ShelfLife:       stock.ShelfLifeOf(30),
		RetailPrice:     money.Rupiah(2000),
		Sales:           SalesHistory{FirstDropAt: daysFromToday(-60)},
	}}
}

func (b productBuilder) withShelfLife(days int) productBuilder {
	b.product.ShelfLife = stock.ShelfLifeOf(days)
	return b
}

func (b productBuilder) withSoldLast14Days(units int) productBuilder {
	b.product.Sales.SoldLast14Days = units
	return b
}

func (b productBuilder) withFirstDropDaysAgo(days int) productBuilder {
	b.product.Sales.FirstDropAt = daysFromToday(-days)
	return b
}

func (b productBuilder) withBatchExpiringIn(days, balance int) productBuilder {
	code := "B" + strconv.Itoa(len(b.product.Batches)+1)
	b.product.Batches = append(b.product.Batches, Batch{
		ID: "batch-" + code, Code: code, Expiry: stock.ExpiresOn(daysFromToday(days)), Balance: balance,
	})
	return b
}

func (b productBuilder) withBatchWithoutExpiry(balance int) productBuilder {
	b.product.Batches = append(b.product.Batches, Batch{ID: "batch-open", Code: "B0", Balance: balance})
	return b
}

func (b productBuilder) build() ProductStock {
	return b.product
}

func aSettlement(dueInDays int, outstanding int64) OpenSettlement {
	return OpenSettlement{
		ID: "settlement-1", Number: "ST-202611-0001", PartnershipID: "partnership-1",
		CounterpartName: "Toko Berkah", DueOn: daysFromToday(dueInDays), Outstanding: money.Rupiah(outstanding),
	}
}

func aPartnership(rhythm terms.Rhythm, uncountedDays, stockUnits int) PartnershipCadence {
	return PartnershipCadence{
		ID: "partnership-1", CounterpartName: "Toko Berkah", Rhythm: rhythm,
		StockUnits: stockUnits, UncountedSince: daysFromToday(-uncountedDays),
	}
}

func aCandidate(score int) Candidate {
	return Candidate{OrganizationID: "org-9", Name: "Warung Sejahtera", Score: score}
}

func storeSnapshot() Snapshot {
	return Snapshot{Side: actor.Store}
}

func consignorSnapshot() Snapshot {
	return Snapshot{Side: actor.Consignor}
}

func onlyInsight(t *testing.T, insights []Insight) Insight {
	t.Helper()
	if len(insights) != 1 {
		t.Fatalf("got %d insights, want exactly 1: %+v", len(insights), insights)
	}
	return insights[0]
}

func expectNone(t *testing.T, insights []Insight) {
	t.Helper()
	if len(insights) != 0 {
		t.Fatalf("got %d insights, want none: %+v", len(insights), insights)
	}
}
