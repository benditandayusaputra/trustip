package domain

import (
	"slices"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
)

const hoursPerDay = 24

type Batch struct {
	ID         string
	Code       string
	Expiry     stock.Expiry
	ReceivedAt time.Time
	Balance    int
}

type Product struct {
	ID          string
	Name        string
	SKU         string
	Unit        string
	ShelfLife   stock.ShelfLife
	Batches     []Batch
	Counterpart Counterpart
}

type Counterpart struct {
	PartnershipID string
	ID            string
	Name          string
	Type          string
}

type BatchView struct {
	Batch
	AgeDays int
	Flag    stock.Flag
}

type ProductView struct {
	Product
	Balance       int
	Flag          stock.Flag
	NearestExpiry string
	BatchViews    []BatchView
}

func (p Product) ViewOn(now time.Time) ProductView {
	today := now.In(period.Jakarta())
	view := ProductView{Product: p, Flag: stock.Safe, BatchViews: make([]BatchView, 0, len(p.Batches))}
	for _, batch := range p.Batches {
		flag := batch.Expiry.FlagOn(today, p.ShelfLife)
		age := ageDays(batch.ReceivedAt, today)
		view.BatchViews = append(view.BatchViews, BatchView{Batch: batch, AgeDays: age, Flag: flag})
		view.Balance += batch.Balance
		view.Flag = worse(view.Flag, flag)
		view.NearestExpiry = earlier(view.NearestExpiry, batch.Expiry)
	}
	return view
}

func ageDays(receivedAt, today time.Time) int {
	received := receivedAt.In(period.Jakarta())
	start := time.Date(received.Year(), received.Month(), received.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	return int(end.Sub(start).Hours() / hoursPerDay)
}

func worse(current, candidate stock.Flag) stock.Flag {
	bySeverity := []stock.Flag{stock.Safe, stock.NearExpiry, stock.Expired}
	if slices.Index(bySeverity, candidate) > slices.Index(bySeverity, current) {
		return candidate
	}
	return current
}

func earlier(current string, expiry stock.Expiry) string {
	date, isGiven := expiry.Date()
	if !isGiven {
		return current
	}
	text := date.Format(time.DateOnly)
	if current == "" || text < current {
		return text
	}
	return current
}

type CardEntry struct {
	ID           string
	Type         string
	Qty          int
	OccurredAt   time.Time
	BatchCode    string
	SourceType   string
	SourceID     string
	SourceNumber string
	BalanceAfter int
}
