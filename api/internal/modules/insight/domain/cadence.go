package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

const (
	dailyRhythmsWithoutCountDays  = 2
	slowerRhythmsWithoutCountDays = 14
)

type PartnershipCadence struct {
	ID              string
	CounterpartName string
	Rhythm          terms.Rhythm
	StockUnits      int
	UncountedSince  time.Time
}

func findNotCounted(partnership PartnershipCadence, today time.Time) []Insight {
	daysUncounted := daysBetween(partnership.UncountedSince, today)
	if partnership.StockUnits <= 0 || daysUncounted < uncountedLimitDays(partnership.Rhythm) {
		return nil
	}
	return []Insight{{
		Kind: KindNotCounted, Level: LevelLow, Days: daysUncounted, Units: partnership.StockUnits,
		Subject: Subject{PartnershipID: partnership.ID, CounterpartName: partnership.CounterpartName},
	}}
}

func uncountedLimitDays(rhythm terms.Rhythm) int {
	if rhythm == terms.RhythmDaily {
		return dailyRhythmsWithoutCountDays
	}
	return slowerRhythmsWithoutCountDays
}
