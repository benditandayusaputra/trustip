package domain

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

func TestNewDashboardQuery_DSH_DueSoonMeansWithinThreeJakartaDays(t *testing.T) {
	lateNightUTC := time.Date(2026, 11, 10, 18, 30, 0, 0, time.UTC)

	got := NewDashboardQuery("org-1", lateNightUTC)

	if got.Today != "2026-11-11" || got.DueSoonUntil != "2026-11-14" {
		t.Fatalf("query = %+v, want today 2026-11-11 and due soon until 2026-11-14", got)
	}
	if !got.TodayStart.Equal(time.Date(2026, 11, 11, 0, 0, 0, 0, period.Jakarta())) {
		t.Fatalf("today start = %s", got.TodayStart)
	}
}

func TestNewDashboardQuery_BRDRP08_PendingMeansReceivedWithinFortyEightHours(t *testing.T) {
	now := today()

	if got := NewDashboardQuery("org-1", now).PendingSince; !got.Equal(now.Add(-48 * time.Hour)) {
		t.Fatalf("pending since = %s, want 48 hours before now", got)
	}
}

func TestNewDashboardQuery_BRNED03_NeedsExpiringWithinThreeDays(t *testing.T) {
	now := today()

	if got := NewDashboardQuery("org-1", now).NeedsExpiring; !got.Equal(now.Add(72 * time.Hour)) {
		t.Fatalf("needs expiring until = %s, want 3 days after now", got)
	}
}

func TestTop_DSH_KeepsAtMostFiveInsights(t *testing.T) {
	insights := make([]Insight, 7)

	if got := len(Top(insights)); got != 5 {
		t.Fatalf("top = %d, want 5", got)
	}
	if got := len(Top(insights[:2])); got != 2 {
		t.Fatalf("top of two = %d, want 2", got)
	}
}
