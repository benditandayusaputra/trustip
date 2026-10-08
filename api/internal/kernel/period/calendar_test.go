package period_test

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

func TestDisplayDate_ReadsAsAnIndonesianDateInJakarta(t *testing.T) {
	lateEvening := time.Date(2026, 10, 25, 18, 30, 0, 0, time.UTC)

	if got := period.DisplayDate(lateEvening); got != "26 Oktober 2026" {
		t.Fatalf("DisplayDate = %q", got)
	}
}

func TestBusinessDate_FollowsTheJakartaCalendar(t *testing.T) {
	if got := period.BusinessDate(time.Date(2026, 10, 25, 17, 0, 0, 0, time.UTC)); got != "2026-10-26" {
		t.Fatalf("BusinessDate = %q", got)
	}
}
