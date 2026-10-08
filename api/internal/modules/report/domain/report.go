package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
)

const (
	MaxRangeDays = 366
	dateLayout   = time.DateOnly
	wholeBPS     = 10000
)

type Grouping string

const (
	ByPartner Grouping = "PARTNER"
	ByProduct Grouping = "PRODUCT"
	ByDay     Grouping = "DAY"
)

func Groupings() []string {
	return []string{string(ByPartner), string(ByProduct), string(ByDay)}
}

type Filter struct {
	From    string
	To      string
	GroupBy string
}

type Query struct {
	OrganizationID string
	From           time.Time
	To             time.Time
	GroupBy        Grouping
}

func NewQuery(filter Filter) (Query, error) {
	firstDay, fromProblem := parseDay("from", filter.From)
	lastDay, toProblem := parseDay("to", filter.To)
	grouping := Grouping(filter.GroupBy)
	if grouping == "" {
		grouping = ByPartner
	}
	problems := field.Collect(fromProblem, toProblem, field.CheckChoice("group_by", string(grouping), Groupings()))
	if fromProblem == (field.Problem{}) && toProblem == (field.Problem{}) {
		problems = append(problems, field.Collect(findRangeProblem(firstDay, lastDay))...)
	}
	if len(problems) > 0 {
		return Query{}, apperror.Validation(problems...)
	}
	return Query{From: firstDay, To: lastDay.AddDate(0, 0, 1), GroupBy: grouping}, nil
}

func parseDay(name, raw string) (time.Time, field.Problem) {
	if raw == "" {
		return time.Time{}, field.Invalid(name, apperror.FieldRequired)
	}
	day, err := time.ParseInLocation(dateLayout, raw, period.Jakarta())
	if err != nil {
		return time.Time{}, field.Invalid(name, apperror.FieldInvalidDate)
	}
	return day, field.Problem{}
}

func findRangeProblem(firstDay, lastDay time.Time) field.Problem {
	if lastDay.Before(firstDay) || !lastDay.Before(firstDay.AddDate(0, 0, MaxRangeDays)) {
		return field.Invalid("to", apperror.FieldOutOfRange)
	}
	return field.Problem{}
}

func (q Query) FirstDay() string {
	return q.From.Format(dateLayout)
}

func (q Query) LastDay() string {
	return q.To.AddDate(0, 0, -1).Format(dateLayout)
}

type Figures struct {
	SoldQty     int
	ReturnedQty int
	ReceivedQty int
	Split       pricing.Split
}

func (f Figures) ReturnedBPS() int {
	if f.ReceivedQty <= 0 {
		return 0
	}
	return (f.ReturnedQty*wholeBPS + f.ReceivedQty/2) / f.ReceivedQty
}

func (f Figures) add(other Figures) Figures {
	return Figures{
		SoldQty: f.SoldQty + other.SoldQty, ReturnedQty: f.ReturnedQty + other.ReturnedQty,
		ReceivedQty: f.ReceivedQty + other.ReceivedQty,
		Split:       pricing.Summarize([]pricing.Split{f.Split, other.Split}),
	}
}

type Row struct {
	Key   string
	Label string
	Figures
}

type Report struct {
	Query   Query
	Rows    []Row
	Summary Figures
}

func Summarize(rows []Row) Figures {
	var total Figures
	for _, row := range rows {
		total = total.add(row.Figures)
	}
	return total
}

func FillDays(query Query, rows []Row) []Row {
	byDay := make(map[string]Row, len(rows))
	for _, row := range rows {
		byDay[row.Key] = row
	}
	days := []Row{}
	for day := query.From; day.Before(query.To); day = day.AddDate(0, 0, 1) {
		key := day.Format(dateLayout)
		row, found := byDay[key]
		if !found {
			row = Row{Key: key, Label: key}
		}
		days = append(days, row)
	}
	return days
}
