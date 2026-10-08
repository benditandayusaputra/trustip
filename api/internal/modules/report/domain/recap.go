package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

type DailyQuery struct {
	OrganizationID string
	Day            string
	From           time.Time
	To             time.Time
}

func NewDailyQuery(rawDay string, now time.Time) (DailyQuery, error) {
	local := now.In(period.Jakarta())
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, period.Jakarta())
	if rawDay != "" {
		parsed, problem := parseDay("date", rawDay)
		if problem != (field.Problem{}) {
			return DailyQuery{}, apperror.Validation(field.Collect(problem)...)
		}
		day = parsed
	}
	return DailyQuery{Day: day.Format(dateLayout), From: day, To: day.AddDate(0, 0, 1)}, nil
}

type RecapRow struct {
	PartnerID string
	Name      string
	Figures
	Paid money.Rupiah
}

func SummarizeRecap(rows []RecapRow) RecapRow {
	var total RecapRow
	for _, row := range rows {
		total.Figures = total.add(row.Figures)
		total.Paid += row.Paid
	}
	return total
}

type DailyRecap struct {
	Query DailyQuery
	Rows  []RecapRow
	Total RecapRow
}

type Tally struct {
	Count  int
	Amount money.Rupiah
}

func (t Tally) plus(other Tally) Tally {
	return Tally{Count: t.Count + other.Count, Amount: t.Amount + other.Amount}
}

type SettlementRow struct {
	PartnerID string
	Name      string
	Issued    Tally
	Paid      Tally
	Unpaid    Tally
	Overdue   Tally
}

func SummarizeSettlements(rows []SettlementRow) SettlementRow {
	var total SettlementRow
	for _, row := range rows {
		total.Issued = total.Issued.plus(row.Issued)
		total.Paid = total.Paid.plus(row.Paid)
		total.Unpaid = total.Unpaid.plus(row.Unpaid)
		total.Overdue = total.Overdue.plus(row.Overdue)
	}
	return total
}

type SettlementQuery struct {
	Query
	Today string
}

type SettlementReport struct {
	Query SettlementQuery
	Rows  []SettlementRow
	Total SettlementRow
}
