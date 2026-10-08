package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

const (
	ReputationWindowDays = 180
	wholeBasisPoints     = 10000
	businessDateLayout   = time.DateOnly
)

type Window struct {
	Since string
	Today string
}

func WindowAt(now time.Time) Window {
	today := now.In(period.Jakarta())
	return Window{
		Since: today.AddDate(0, 0, -ReputationWindowDays).Format(businessDateLayout),
		Today: today.Format(businessDateLayout),
	}
}

type Facts struct {
	OrganizationID    string
	ActivePartners    int
	DueSettlements    int
	OnTimeSettlements int
	Products          int
}

type Stats struct {
	OrganizationID     string
	ActivePartners     int
	OnTimeBPS          int
	SettlementsCounted int
	Products           int
}

func Summarize(facts []Facts) []Stats {
	stats := make([]Stats, 0, len(facts))
	for _, fact := range facts {
		stats = append(stats, Stats{
			OrganizationID: fact.OrganizationID, ActivePartners: fact.ActivePartners,
			OnTimeBPS:          OnTimeBPS(fact.OnTimeSettlements, fact.DueSettlements),
			SettlementsCounted: fact.DueSettlements, Products: fact.Products,
		})
	}
	return stats
}

func OnTimeBPS(onTime, due int) int {
	if due <= 0 {
		return 0
	}
	return (onTime*wholeBasisPoints + due/2) / due
}
