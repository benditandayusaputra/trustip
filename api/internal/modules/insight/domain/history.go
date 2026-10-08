package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

type HistoryQuery struct {
	OrganizationID string
	PartnershipID  string
	Now            time.Time
	SalesSince     time.Time
	ISOWeekday     int
	MaxOccurrences int
}

func NewHistoryQuery(organizationID, partnershipID string, now time.Time) HistoryQuery {
	today := now.In(period.Jakarta())
	startOfToday := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, period.Jakarta())
	return HistoryQuery{
		OrganizationID: organizationID, PartnershipID: partnershipID, Now: now,
		SalesSince:     startOfToday.AddDate(0, 0, 1-salesWindowDays),
		ISOWeekday:     isoWeekday(today.Weekday()),
		MaxOccurrences: maxSameWeekdayOccurrences,
	}
}

func isoWeekday(weekday time.Weekday) int {
	if weekday == time.Sunday {
		return int(time.Saturday) + 1
	}
	return int(weekday)
}

type ProductHistory struct {
	ProductID string
	History   DropHistory
}

type DropSuggestion struct {
	ProductID     string
	Qty           int
	HasEnoughData bool
}

func SuggestDrops(products []ProductHistory, now time.Time) []DropSuggestion {
	suggestions := make([]DropSuggestion, 0, len(products))
	for _, product := range products {
		qty, hasEnoughData := SuggestDropQty(product.History, now)
		suggestions = append(suggestions, DropSuggestion{
			ProductID: product.ProductID, Qty: qty, HasEnoughData: hasEnoughData,
		})
	}
	return suggestions
}
