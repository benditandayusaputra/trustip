package domain

import (
	"errors"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

const daysPerWeek = 7

var ErrNoCalendarPeriod = errors.New("the schedule issues on count and has no calendar period")

type Period struct {
	Start time.Time
	End   time.Time
}

func (p Period) Equal(other Period) bool {
	return p.Start.Equal(other.Start) && p.End.Equal(other.End)
}

func (p Period) Contains(at time.Time) bool {
	return !at.Before(p.Start) && at.Before(p.End)
}

func (p Period) Overlaps(other Period) bool {
	return p.Start.Before(other.End) && other.Start.Before(p.End)
}

type calendarRule func(at time.Time) Period

func calendarRules() map[terms.Schedule]calendarRule {
	return map[terms.Schedule]calendarRule{
		terms.ScheduleWeekly:  weekContaining,
		terms.ScheduleMonthly: monthContaining,
	}
}

func PeriodContaining(schedule terms.Schedule, at time.Time) (Period, error) {
	rule, isCalendar := calendarRules()[schedule]
	if !isCalendar {
		return Period{}, ErrNoCalendarPeriod
	}
	return rule(at), nil
}

func weekContaining(at time.Time) Period {
	day := startOfDay(at)
	daysSinceMonday := (int(day.Weekday()) + daysPerWeek - 1) % daysPerWeek
	start := day.AddDate(0, 0, -daysSinceMonday)
	return Period{Start: start, End: start.AddDate(0, 0, daysPerWeek)}
}

func monthContaining(at time.Time) Period {
	local := at.In(period.Jakarta())
	start := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, period.Jakarta())
	return Period{Start: start, End: start.AddDate(0, 1, 0)}
}

func startOfDay(at time.Time) time.Time {
	local := at.In(period.Jakarta())
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, period.Jakarta())
}
