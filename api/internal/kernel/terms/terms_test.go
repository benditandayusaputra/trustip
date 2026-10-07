package terms

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
)

func TestCheck_FRAGR01_GeneralTermsAreKeptWhenValid(t *testing.T) {
	general, problems := Check(Input{SettlementSchedule: "MONTHLY", PaymentTermsDays: 30, CountRhythm: "EACH_VISIT"})

	found := field.Collect(problems...)
	if len(found) != 0 || general.Schedule != ScheduleMonthly || general.Rhythm != RhythmEachVisit {
		t.Fatalf("general = %+v, problems = %v", general, found)
	}
}

func TestCheck_PaymentTermsUpToThirtyDays(t *testing.T) {
	input := Input{SettlementSchedule: "WEEKLY", PaymentTermsDays: MaxPaymentTermsDays + 1, CountRhythm: "DAILY"}

	_, problems := Check(input)

	found := field.Collect(problems...)
	if len(found) != 1 || found[0].Field != "terms.payment_terms_days" {
		t.Fatalf("problems = %v", found)
	}
}
