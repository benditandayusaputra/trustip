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

type ListFilter struct {
	PartnershipID string
	From          string
	To            string
}

type ListQuery struct {
	OrganizationID string
	PartnershipID  string
	From           time.Time
	To             time.Time
	ConfirmCutoff  time.Time
	After          pagination.Position
	Limit          int
}

func NewListQuery(filter ListFilter) (ListQuery, error) {
	start, startProblem := parseDay("from", filter.From, time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC))
	end, endProblem := parseDay("to", filter.To, time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC))
	if filter.To != "" {
		end = end.AddDate(0, 0, 1)
	}
	problems := field.Collect(startProblem, endProblem, findRangeProblem(start, end),
		findPartnershipProblem(filter.PartnershipID))
	if len(problems) > 0 {
		return ListQuery{}, apperror.Validation(problems...)
	}
	return ListQuery{PartnershipID: filter.PartnershipID, From: start, To: end}, nil
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

func findRangeProblem(start, end time.Time) field.Problem {
	if !start.Before(end) {
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

type Summary struct {
	ID              string
	Number          string
	PartnershipID   string
	Status          Status
	CountedAt       time.Time
	RecordedOffline bool
	CounterpartID   string
	CounterpartName string
	CounterpartType string
	SoldQty         int
	ReturnedQty     int
	ConsignorShare  int64
}
