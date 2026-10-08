package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

const hoursPerDay = 24

func daysBetween(from, until time.Time) int {
	return int(calendarDay(until).Sub(calendarDay(from)).Hours() / hoursPerDay)
}

func calendarDay(at time.Time) time.Time {
	local := at.In(period.Jakarta())
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}
