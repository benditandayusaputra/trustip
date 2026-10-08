package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
)

const lowStockDays = 3

type Batch struct {
	ID      string
	Code    string
	Expiry  stock.Expiry
	Balance int
}

type ProductStock struct {
	PartnershipID   string
	CounterpartName string
	ProductID       string
	ProductName     string
	ShelfLife       stock.ShelfLife
	RetailPrice     money.Rupiah
	Sales           SalesHistory
	Batches         []Batch
}

func findExpiryInsights(product ProductStock, today time.Time) []Insight {
	insights := []Insight{}
	for _, batch := range product.Batches {
		if batch.Balance <= 0 {
			continue
		}
		switch batch.Expiry.FlagOn(today, product.ShelfLife) {
		case stock.Expired:
			insights = append(insights, expiredInsight(product, batch))
		case stock.NearExpiry:
			insights = append(insights, nearExpiryInsight(product, batch, today))
		case stock.Safe:
		}
	}
	return insights
}

func expiredInsight(product ProductStock, batch Batch) Insight {
	return Insight{
		Kind: KindExpired, Level: LevelHigh, Subject: product.batchSubject(batch),
		Units: batch.Balance, Amount: product.valueOf(batch.Balance),
	}
}

func nearExpiryInsight(product ProductStock, batch Batch, today time.Time) Insight {
	daysLeft, _ := batch.Expiry.DaysLeftOn(today)
	unsold := rateOf(product.Sales, today).unsoldWithin(batch.Balance, daysLeft)
	level := LevelMedium
	if unsold > 0 {
		level = LevelHigh
	}
	return Insight{
		Kind: KindNearExpiry, Level: level, Subject: product.batchSubject(batch),
		Units: unsold, Days: daysLeft, Amount: product.valueOf(unsold),
	}
}

func findLowStock(product ProductStock, today time.Time) []Insight {
	rate := rateOf(product.Sales, today)
	stockUnits := product.balance()
	if !rate.isSelling() || !rate.isStockLastingFewerThan(stockUnits, lowStockDays) {
		return nil
	}
	return []Insight{{
		Kind: KindLowStock, Level: LevelMedium, Subject: product.subject(),
		Units: stockUnits, Days: rate.daysOfStock(stockUnits),
	}}
}

func (p ProductStock) balance() int {
	total := 0
	for _, batch := range p.Batches {
		total += max(batch.Balance, 0)
	}
	return total
}

func (p ProductStock) valueOf(units int) money.Rupiah {
	return p.RetailPrice * money.Rupiah(units)
}

func (p ProductStock) subject() Subject {
	return Subject{
		PartnershipID: p.PartnershipID, CounterpartName: p.CounterpartName,
		ProductID: p.ProductID, ProductName: p.ProductName,
	}
}

func (p ProductStock) batchSubject(batch Batch) Subject {
	subject := p.subject()
	subject.BatchID = batch.ID
	subject.BatchCode = batch.Code
	return subject
}
