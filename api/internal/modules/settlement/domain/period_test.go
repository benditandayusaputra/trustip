package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

func jakarta(wallClock string) time.Time {
	if len(wallClock) == len(time.DateOnly) {
		wallClock += " 00:00"
	}
	at, err := time.ParseInLocation("2006-01-02 15:04", wallClock, period.Jakarta())
	if err != nil {
		return time.Time{}
	}
	return at
}

func TestPeriodContaining_BRSTL07_WeekRunsFromMondayMidnightInJakarta(t *testing.T) {
	cases := []struct {
		name string
		at   time.Time
		want Period
	}{
		{"monday at midnight", jakarta("2026-11-09"), Period{jakarta("2026-11-09"), jakarta("2026-11-16")}},
		{"sunday 23.59", jakarta("2026-11-15 23:59"), Period{jakarta("2026-11-09"), jakarta("2026-11-16")}},
		{
			"sunday 17.30 UTC is already monday in Jakarta",
			time.Date(2026, 11, 15, 17, 30, 0, 0, time.UTC),
			Period{jakarta("2026-11-16"), jakarta("2026-11-23")},
		},
		{"week across a month", jakarta("2026-12-01 09:00"), Period{jakarta("2026-11-30"), jakarta("2026-12-07")}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := PeriodContaining(terms.ScheduleWeekly, tc.at)
			if err != nil || !got.Equal(tc.want) {
				t.Fatalf("period = %+v, err = %v, want %+v", got, err, tc.want)
			}
		})
	}
}

func TestPeriodContaining_BRSTL07_MonthRunsFromTheFirstToTheEnd(t *testing.T) {
	cases := []struct {
		name string
		at   time.Time
		want Period
	}{
		{"first at midnight", jakarta("2026-11-01"), Period{jakarta("2026-11-01"), jakarta("2026-12-01")}},
		{"last day 23.59", jakarta("2026-11-30 23:59"), Period{jakarta("2026-11-01"), jakarta("2026-12-01")}},
		{"leap february", jakarta("2028-02-29 12:00"), Period{jakarta("2028-02-01"), jakarta("2028-03-01")}},
		{
			"31 december 17.00 UTC is january in Jakarta",
			time.Date(2026, 12, 31, 17, 0, 0, 0, time.UTC),
			Period{jakarta("2027-01-01"), jakarta("2027-02-01")},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := PeriodContaining(terms.ScheduleMonthly, tc.at)
			if err != nil || !got.Equal(tc.want) {
				t.Fatalf("period = %+v, err = %v, want %+v", got, err, tc.want)
			}
		})
	}
}

func TestPeriodContaining_OnCountHasNoCalendarPeriod(t *testing.T) {
	_, err := PeriodContaining(terms.ScheduleOnCount, jakarta("2026-11-09"))
	if !errors.Is(err, ErrNoCalendarPeriod) {
		t.Fatalf("err = %v", err)
	}
}

func TestPeriod_BRSTL02_OverlapIsHalfOpen(t *testing.T) {
	week := Period{jakarta("2026-11-09"), jakarta("2026-11-16")}
	next := Period{jakarta("2026-11-16"), jakarta("2026-11-23")}
	inside := Period{jakarta("2026-11-15"), jakarta("2026-11-17")}
	if week.Overlaps(next) || !week.Overlaps(inside) || !next.Overlaps(inside) {
		t.Fatalf("overlaps = %v %v %v", week.Overlaps(next), week.Overlaps(inside), next.Overlaps(inside))
	}
	if !week.Contains(jakarta("2026-11-09")) || week.Contains(jakarta("2026-11-16")) {
		t.Fatal("contains must include the start and exclude the end")
	}
}
