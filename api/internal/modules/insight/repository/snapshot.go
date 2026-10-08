package repository

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/insight/domain"
)

type stockSnapshot struct {
	products     []domain.ProductStock
	partnerships []domain.PartnershipCadence
}

func stockOf(rows []sqlc.ListInsightStockRow) stockSnapshot {
	snapshot := stockSnapshot{}
	for _, row := range rows {
		snapshot.addProduct(row)
		snapshot.addPartnership(row)
		if row.BatchID != "" {
			last := &snapshot.products[len(snapshot.products)-1]
			last.Batches = append(last.Batches, domain.Batch{
				ID: row.BatchID, Code: row.BatchCode, Expiry: expiryOf(row.ExpiryDate), Balance: int(row.Balance),
			})
			snapshot.partnerships[len(snapshot.partnerships)-1].StockUnits += int(row.Balance)
		}
	}
	return snapshot
}

func (s *stockSnapshot) addProduct(row sqlc.ListInsightStockRow) {
	last := len(s.products) - 1
	if last >= 0 && s.products[last].PartnershipID == row.PartnershipID && s.products[last].ProductID == row.ProductID {
		return
	}
	s.products = append(s.products, domain.ProductStock{
		PartnershipID: row.PartnershipID, CounterpartName: row.CounterpartName,
		ProductID: row.ProductID, ProductName: row.ProductName, ShelfLife: shelfLifeOf(row.ShelfLifeDays),
		RetailPrice: money.Rupiah(row.RetailPrice),
		Sales:       domain.SalesHistory{SoldLast14Days: int(row.SoldSince), FirstDropAt: row.FirstDropAt},
	})
}

func (s *stockSnapshot) addPartnership(row sqlc.ListInsightStockRow) {
	last := len(s.partnerships) - 1
	if last >= 0 && s.partnerships[last].ID == row.PartnershipID {
		return
	}
	s.partnerships = append(s.partnerships, domain.PartnershipCadence{
		ID: row.PartnershipID, CounterpartName: row.CounterpartName,
		Rhythm: terms.Rhythm(row.CountRhythm), UncountedSince: row.UncountedSince,
	})
}

func expiryOf(text string) stock.Expiry {
	date, err := time.Parse(time.DateOnly, text)
	if err != nil {
		return stock.NoExpiry()
	}
	return stock.ExpiresOn(date)
}
