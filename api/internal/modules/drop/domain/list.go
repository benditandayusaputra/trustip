package domain

import (
	"regexp"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

var idFormat = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func earliestListed() time.Time {
	return time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
}

func latestListed() time.Time {
	return time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)
}

type ListFilter struct {
	PartnershipID string
	Status        string
	From          string
	To            string
}

type ListQuery struct {
	OrganizationID string
	PartnershipID  string
	Status         Status
	From           time.Time
	To             time.Time
	ConfirmCutoff  time.Time
}

func NewListQuery(filter ListFilter, now time.Time) (ListQuery, error) {
	start, startProblem := parseDay("from", filter.From, earliestListed())
	end, endProblem := parseDay("to", filter.To, latestListed())
	if filter.To != "" {
		end = end.AddDate(0, 0, 1)
	}
	problems := field.Collect(
		startProblem, endProblem, findRangeProblem(start, end),
		field.CheckChoice("status", filter.Status, []string{string(Recorded), string(Confirmed), string(Cancelled)}),
		findPartnershipProblem(filter.PartnershipID),
	)
	if len(problems) > 0 {
		return ListQuery{}, apperror.Validation(problems...)
	}
	return ListQuery{
		PartnershipID: filter.PartnershipID, Status: Status(filter.Status), From: start, To: end,
		ConfirmCutoff: now.Add(-ConfirmWindow),
	}, nil
}

func parseDay(name, raw string, fallback time.Time) (time.Time, field.Problem) {
	if raw == "" {
		return fallback, field.Problem{}
	}
	day, err := time.ParseInLocation(time.DateOnly, raw, period.Jakarta())
	if err != nil {
		return fallback, field.Invalid(name, apperror.FieldInvalidDate)
	}
	return day, field.Problem{}
}

func findRangeProblem(from, to time.Time) field.Problem {
	if !from.Before(to) {
		return field.Invalid("to", apperror.FieldOutOfRange)
	}
	return field.Problem{}
}

func findPartnershipProblem(partnershipID string) field.Problem {
	if partnershipID == "" || idFormat.MatchString(partnershipID) {
		return field.Problem{}
	}
	return field.Invalid("partnership_id", apperror.FieldUnknownChoice)
}

type PageQuery struct {
	ListQuery
	After pagination.Position
	Limit int
}

type Summary struct {
	ID              string
	Number          string
	PartnershipID   string
	Status          Status
	OccurredAt      time.Time
	RecordedOffline bool
	CounterpartID   string
	CounterpartName string
	CounterpartType string
	TotalQty        int
	LineCount       int
}
