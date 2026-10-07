package period

import (
	"testing"
	"time"
)

func TestMonthKey_BRSTL07_FollowsTheCalendarInJakarta(t *testing.T) {
	cases := map[time.Time]string{
		time.Date(2026, 10, 31, 16, 59, 0, 0, time.UTC): "202610",
		time.Date(2026, 10, 31, 17, 0, 0, 0, time.UTC):  "202611",
		time.Date(2026, 12, 31, 17, 0, 0, 0, time.UTC):  "202701",
	}
	for at, want := range cases {
		if got := MonthKey(at); got != want {
			t.Errorf("MonthKey(%s) = %s, want %s", at, got, want)
		}
	}
}
