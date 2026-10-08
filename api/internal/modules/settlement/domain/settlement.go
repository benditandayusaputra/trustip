package domain

import (
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

const (
	EntityType   = "SETTLEMENT"
	NumberPrefix = "ST"
	hoursPerDay  = 24
)

type Status string

const (
	Issued Status = "ISSUED"
	Paid   Status = "PAID"
)

type Settlement struct {
	ID            string
	Number        string
	PartnershipID string
	ConsignorID   string
	StoreID       string
	Schedule      terms.Schedule
	Period        Period
	Status        Status
	Summary       Summary
	Allocations   []Allocation
	DueDate       time.Time
	IssuedAt      time.Time
}

type Issuing struct {
	ID       string
	Sequence int
	At       time.Time
}

func Number(at time.Time, sequence int) string {
	return fmt.Sprintf("%s-%s-%04d", NumberPrefix, period.MonthKey(at), sequence)
}

func (d Draft) Issue(issuing Issuing) Settlement {
	return Settlement{
		ID: issuing.ID, Number: Number(issuing.At, issuing.Sequence), PartnershipID: d.PartnershipID,
		ConsignorID: d.ConsignorID, StoreID: d.StoreID, Schedule: d.Schedule, Period: d.Period, Status: Issued,
		Summary: d.Summary, Allocations: d.Allocations,
		DueDate: startOfDay(issuing.At).AddDate(0, 0, d.PaymentTermsDays), IssuedAt: issuing.At,
	}
}

type Age string

const (
	AgePaid       Age = "PAID"
	AgeNotDue     Age = "NOT_DUE"
	AgeLate1To7   Age = "LATE_1_7"
	AgeLate8To30  Age = "LATE_8_30"
	AgeLateOver30 Age = "LATE_OVER_30"
)

const (
	lastDayOfFirstAge  = 7
	lastDayOfSecondAge = 30
)

type Standing struct {
	IsOverdue bool
	DaysLate  int
	Age       Age
}

func StandingOf(status Status, dueDate, now time.Time) Standing {
	if status == Paid {
		return Standing{Age: AgePaid}
	}
	daysLate := int(startOfDay(now).Sub(startOfDay(dueDate)).Round(time.Hour).Hours()) / hoursPerDay
	if daysLate <= 0 {
		return Standing{Age: AgeNotDue}
	}
	return Standing{IsOverdue: true, DaysLate: daysLate, Age: ageOf(daysLate)}
}

func ageOf(daysLate int) Age {
	if daysLate <= lastDayOfFirstAge {
		return AgeLate1To7
	}
	if daysLate <= lastDayOfSecondAge {
		return AgeLate8To30
	}
	return AgeLateOver30
}
