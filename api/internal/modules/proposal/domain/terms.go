package domain

import (
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
)

const (
	MaxPaymentTermsDays = 30
	MaxNotesLength      = 500
)

type Schedule string

const (
	ScheduleOnCount Schedule = "ON_COUNT"
	ScheduleWeekly  Schedule = "WEEKLY"
	ScheduleMonthly Schedule = "MONTHLY"
)

type Rhythm string

const (
	RhythmDaily     Rhythm = "DAILY"
	RhythmEachVisit Rhythm = "EACH_VISIT"
	RhythmWeekly    Rhythm = "WEEKLY"
)

type TermsInput struct {
	SettlementSchedule string
	PaymentTermsDays   int
	CountRhythm        string
	Notes              string
}

type GeneralTerms struct {
	Schedule         Schedule
	PaymentTermsDays int
	Rhythm           Rhythm
	Notes            string
}

func checkTerms(input TermsInput) (GeneralTerms, []field.Problem) {
	problems := []field.Problem{
		findRequiredChoice("terms.settlement_schedule", input.SettlementSchedule,
			[]string{string(ScheduleOnCount), string(ScheduleWeekly), string(ScheduleMonthly)}),
		findRequiredChoice("terms.count_rhythm", input.CountRhythm,
			[]string{string(RhythmDaily), string(RhythmEachVisit), string(RhythmWeekly)}),
		findPaymentTermsProblem(input.PaymentTermsDays),
		field.CheckLength("terms.notes", input.Notes, field.Length{Max: MaxNotesLength}),
	}
	return GeneralTerms{
		Schedule: Schedule(input.SettlementSchedule), PaymentTermsDays: input.PaymentTermsDays,
		Rhythm: Rhythm(input.CountRhythm), Notes: input.Notes,
	}, problems
}

func findRequiredChoice(name, value string, choices []string) field.Problem {
	if value == "" {
		return field.Invalid(name, apperror.FieldRequired)
	}
	return field.CheckChoice(name, value, choices)
}

func findPaymentTermsProblem(days int) field.Problem {
	if days < 0 || days > MaxPaymentTermsDays {
		return field.Invalid("terms.payment_terms_days", apperror.FieldOutOfRange)
	}
	return field.Problem{}
}
