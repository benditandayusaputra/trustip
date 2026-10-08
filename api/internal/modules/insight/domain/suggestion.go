package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

const (
	maxSameWeekdayOccurrences = 4
	minSameWeekdayOccurrences = 2
	minRateDays               = 3
	coverDays                 = 7
	wholePercent              = 100
	steadyFactorPercent       = 110
	soldOutFactorPercent      = 125
	minDailySuggestion        = 1
)

type DaySales struct {
	Sold    int
	SoldOut bool
}

type DropHistory struct {
	Rhythm           terms.Rhythm
	SameWeekdaySales []DaySales
	Sales            SalesHistory
	StockUnits       int
	ShelfLife        stock.ShelfLife
}

func SuggestDropQty(history DropHistory, today time.Time) (int, bool) {
	if history.Rhythm == terms.RhythmDaily {
		return suggestForDailyRhythm(history.SameWeekdaySales)
	}
	return suggestFromDailyRate(history, rateOf(history.Sales, today))
}

func suggestForDailyRhythm(newestFirst []DaySales) (int, bool) {
	occurrences := newestFirst[:min(len(newestFirst), maxSameWeekdayOccurrences)]
	if len(occurrences) < minSameWeekdayOccurrences {
		return 0, false
	}
	totalSold := 0
	for _, occurrence := range occurrences {
		totalSold += occurrence.Sold
	}
	suggestion := ceilDiv(totalSold*factorPercent(occurrences), len(occurrences)*wholePercent)
	return max(suggestion, minDailySuggestion), true
}

func factorPercent(newestFirst []DaySales) int {
	if newestFirst[0].SoldOut && newestFirst[1].SoldOut {
		return soldOutFactorPercent
	}
	return steadyFactorPercent
}

func suggestFromDailyRate(history DropHistory, rate dailyRate) (int, bool) {
	if rate.days < minRateDays {
		return 0, false
	}
	needed := ceilDiv(rate.sold*coverDays*steadyFactorPercent, rate.days*wholePercent)
	suggestion := max(0, needed-history.StockUnits)
	if shelfLifeDays, given := history.ShelfLife.Days(); given {
		suggestion = min(suggestion, ceilDiv(rate.sold*shelfLifeDays, rate.days))
	}
	return suggestion, true
}

func ceilDiv(numerator, denominator int) int {
	return (numerator + denominator - 1) / denominator
}
