package domain

import (
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

type Ledger struct {
	PartnershipID    string
	ConsignorID      string
	StoreID          string
	StoreName        string
	Schedule         terms.Schedule
	PaymentTermsDays int
	SettledUntil     time.Time
	Unsettled        []Allocation
}

type Draft struct {
	PartnershipID    string
	ConsignorID      string
	StoreID          string
	Schedule         terms.Schedule
	PaymentTermsDays int
	Period           Period
	Allocations      []Allocation
	Summary          Summary
}

func (l Ledger) DraftUpTo(lastSaleAt time.Time) (Draft, bool) {
	end := lastSaleAt.Truncate(time.Second).Add(time.Second)
	included := l.salesBefore(end)
	if len(included) == 0 {
		return Draft{}, false
	}
	start := included[0].OccurredAt
	if !l.SettledUntil.IsZero() {
		start = l.SettledUntil
	}
	return l.draftOf(Period{Start: start, End: end})
}

func (l Ledger) DraftClosedPeriods(now time.Time) ([]Draft, error) {
	current, err := PeriodContaining(l.Schedule, now)
	if err != nil {
		return nil, fmt.Errorf("find the current period of %s: %w", l.PartnershipID, err)
	}
	drafts := []Draft{}
	for _, natural := range l.closedPeriods(current.Start) {
		if draft, hasSales := l.draftOf(l.clip(natural)); hasSales {
			drafts = append(drafts, draft)
		}
	}
	return drafts, nil
}

func (l Ledger) closedPeriods(currentStart time.Time) []Period {
	periods := []Period{}
	for _, allocation := range l.salesBefore(currentStart) {
		natural, err := PeriodContaining(l.Schedule, allocation.OccurredAt)
		if err != nil {
			continue
		}
		if len(periods) == 0 || !periods[len(periods)-1].Equal(natural) {
			periods = append(periods, natural)
		}
	}
	return periods
}

func (l Ledger) clip(natural Period) Period {
	if l.SettledUntil.After(natural.Start) {
		natural.Start = l.SettledUntil
	}
	return natural
}

func (l Ledger) salesBefore(end time.Time) []Allocation {
	included := []Allocation{}
	for _, allocation := range l.Unsettled {
		if allocation.OccurredAt.Before(end) {
			included = append(included, allocation)
		}
	}
	return included
}

func (l Ledger) draftOf(period Period) (Draft, bool) {
	included := []Allocation{}
	for _, allocation := range l.Unsettled {
		if period.Contains(allocation.OccurredAt) {
			included = append(included, allocation)
		}
	}
	if len(included) == 0 {
		return Draft{}, false
	}
	return Draft{
		PartnershipID: l.PartnershipID, ConsignorID: l.ConsignorID, StoreID: l.StoreID, Schedule: l.Schedule,
		PaymentTermsDays: l.PaymentTermsDays, Period: period, Allocations: included, Summary: Summarize(included),
	}, true
}
