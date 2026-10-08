package domain

import "time"

const salesWindowDays = 14

type SalesHistory struct {
	SoldLast14Days int
	FirstDropAt    time.Time
}

type dailyRate struct {
	sold int
	days int
}

func rateOf(sales SalesHistory, today time.Time) dailyRate {
	days := min(max(daysBetween(sales.FirstDropAt, today), 1), salesWindowDays)
	return dailyRate{sold: sales.SoldLast14Days, days: days}
}

func (r dailyRate) isSelling() bool {
	return r.sold > 0
}

func (r dailyRate) isStockLastingFewerThan(stockUnits, days int) bool {
	return stockUnits*r.days < days*r.sold
}

func (r dailyRate) daysOfStock(stockUnits int) int {
	return stockUnits * r.days / r.sold
}

func (r dailyRate) unsoldWithin(stockUnits, days int) int {
	excessTimesDays := stockUnits*r.days - r.sold*days
	if excessTimesDays <= 0 {
		return 0
	}
	return (excessTimesDays + r.days - 1) / r.days
}
