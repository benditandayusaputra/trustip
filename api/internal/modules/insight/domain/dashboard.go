package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

const (
	dueSoonDays          = 3
	confirmationWindow   = 48 * time.Hour
	maxDashboardInsights = 5
	businessDateLayout   = time.DateOnly
)

type DashboardQuery struct {
	OrganizationID string
	Today          string
	DueSoonUntil   string
	TodayStart     time.Time
	PendingSince   time.Time
}

func NewDashboardQuery(organizationID string, now time.Time) DashboardQuery {
	startOfToday := startOfDay(now)
	return DashboardQuery{
		OrganizationID: organizationID,
		Today:          startOfToday.Format(businessDateLayout),
		DueSoonUntil:   startOfToday.AddDate(0, 0, dueSoonDays).Format(businessDateLayout),
		TodayStart:     startOfToday,
		PendingSince:   now.Add(-confirmationWindow),
	}
}

type Tally struct {
	Count  int
	Amount money.Rupiah
}

type Figures struct {
	ActivePartners int
	StockUnits     int
	Unpaid         Tally
	Overdue        Tally
	DueSoon        Tally
}

type Tasks struct {
	DropsToConfirm    int
	CountsToConfirm   int
	PaymentsToReview  int
	PaymentsRejected  int
	ProposalsAwaiting int
}

type UncountedPartner struct {
	PartnershipID   string
	CounterpartName string
	StockUnits      int
}

type Dashboard struct {
	Figures    Figures
	Tasks      Tasks
	QuickCount []UncountedPartner
}

type SnapshotQuery struct {
	OrganizationID string
	Now            time.Time
	SalesSince     time.Time
}

func NewSnapshotQuery(organizationID string, now time.Time) SnapshotQuery {
	return SnapshotQuery{
		OrganizationID: organizationID, Now: now,
		SalesSince: startOfDay(now).AddDate(0, 0, 1-salesWindowDays),
	}
}

func Top(insights []Insight) []Insight {
	return insights[:min(len(insights), maxDashboardInsights)]
}

func startOfDay(at time.Time) time.Time {
	local := at.In(period.Jakarta())
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, period.Jakarta())
}
