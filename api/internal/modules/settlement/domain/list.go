package domain

import (
	"regexp"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
)

const (
	earliestDue = "0001-01-01"
	latestDue   = "9999-12-31"
)

var idFormat = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type ListFilter struct {
	PartnershipID string
	Status        string
	Age           string
}

type ListQuery struct {
	OrganizationID string
	PartnershipID  string
	Status         string
	DueFrom        string
	DueTo          string
	After          pagination.Position
	Limit          int
}

type dueBound struct {
	daysFromToday int
	isBounded     bool
}

type dueRange struct {
	from dueBound
	to   dueBound
}

func (b dueBound) dateOr(today time.Time, unbounded string) string {
	if !b.isBounded {
		return unbounded
	}
	return today.AddDate(0, 0, b.daysFromToday).Format(dateLayout)
}

func lateAges() map[Age]dueRange {
	return map[Age]dueRange{
		AgeNotDue:     {from: dueBound{0, true}},
		AgeLate1To7:   {from: dueBound{-lastDayOfFirstAge, true}, to: dueBound{-1, true}},
		AgeLate8To30:  {from: dueBound{-lastDayOfSecondAge, true}, to: dueBound{-lastDayOfFirstAge - 1, true}},
		AgeLateOver30: {to: dueBound{-lastDayOfSecondAge - 1, true}},
	}
}

func NewListQuery(filter ListFilter, now time.Time) (ListQuery, error) {
	age := lateAges()[Age(filter.Age)]
	problems := field.Collect(
		findPartnershipProblem(filter.PartnershipID),
		findStatusProblem(filter.Status),
		findAgeProblem(filter),
	)
	if len(problems) > 0 {
		return ListQuery{}, apperror.Validation(problems...)
	}
	query := ListQuery{
		PartnershipID: filter.PartnershipID, Status: filter.Status, DueFrom: earliestDue, DueTo: latestDue,
	}
	if filter.Age == "" {
		return query, nil
	}
	today := startOfDay(now)
	query.Status = string(Issued)
	query.DueFrom = age.from.dateOr(today, earliestDue)
	query.DueTo = age.to.dateOr(today, latestDue)
	return query, nil
}

func findPartnershipProblem(partnershipID string) field.Problem {
	if partnershipID == "" || idFormat.MatchString(partnershipID) {
		return field.Problem{}
	}
	return field.Invalid("partnership_id", apperror.FieldUnknownChoice)
}

func findStatusProblem(status string) field.Problem {
	if status == "" {
		return field.Problem{}
	}
	return field.CheckChoice("status", status, []string{string(Issued), string(Paid)})
}

func findAgeProblem(filter ListFilter) field.Problem {
	if filter.Age == "" {
		return field.Problem{}
	}
	if _, isKnown := lateAges()[Age(filter.Age)]; !isKnown {
		return field.Invalid("age", apperror.FieldUnknownChoice)
	}
	if filter.Status == string(Paid) {
		return field.Invalid("age", apperror.FieldOutOfRange)
	}
	return field.Problem{}
}

type ListItem struct {
	ID              string
	Number          string
	PartnershipID   string
	Status          Status
	Schedule        string
	Period          Period
	DueDate         time.Time
	IssuedAt        time.Time
	SoldQty         int
	ConsignorShare  int64
	CounterpartID   string
	CounterpartName string
	CounterpartType string
	Standing        Standing
}
