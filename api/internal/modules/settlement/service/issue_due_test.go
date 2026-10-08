package service_test

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/service"
)

func TestIssueDueSettlements_FRSTL02_IssuesTheEndedWeekUnderALock(t *testing.T) {
	run := newScenario(terms.ScheduleWeekly, 18, 4)
	monday := time.Date(2026, 11, 16, 0, 10, 0, 0, countedAt().Location())

	issued, err := service.NewIssueDueSettlements(run.deps).Execute(t.Context(), terms.ScheduleWeekly, monday)

	if err != nil || issued != 1 || run.settlements.locks != 1 || len(run.settlements.created) != 1 {
		t.Fatalf("issued = %d, err = %v, locks = %d", issued, err, run.settlements.locks)
	}
	key := run.settlements.dueKeys[0]
	if key.Schedule != "WEEKLY" || !key.Before.Equal(time.Date(2026, 11, 16, 0, 0, 0, 0, monday.Location())) {
		t.Fatalf("due key = %+v", key)
	}
	settlement := run.settlements.created[0]
	if settlement.Summary.Qty != 22 || settlement.Schedule != terms.ScheduleWeekly ||
		len(run.events.notices) != 2 {
		t.Fatalf("settlement = %+v", settlement)
	}
}

func TestIssueDueSettlements_SkipsAPartnershipThatChangedSchedule(t *testing.T) {
	run := newScenario(terms.ScheduleMonthly, 18)
	monday := time.Date(2026, 11, 16, 0, 10, 0, 0, countedAt().Location())

	issued, err := service.NewIssueDueSettlements(run.deps).Execute(t.Context(), terms.ScheduleWeekly, monday)

	if err != nil || issued != 0 || len(run.settlements.created) != 0 {
		t.Fatalf("issued = %d, err = %v", issued, err)
	}
}

func TestIssueDueSettlements_OnCountHasNoCalendarRun(t *testing.T) {
	run := newScenario(terms.ScheduleOnCount, 18)

	_, err := service.NewIssueDueSettlements(run.deps).Execute(t.Context(), terms.ScheduleOnCount, countedAt())

	if err == nil {
		t.Fatal("an on-count schedule has no calendar period to issue")
	}
}

func TestIssueOutstanding_BRPRT03_EndingIssuesTheRunningSalesWhateverTheSchedule(t *testing.T) {
	run := newScenario(terms.ScheduleMonthly, 18)

	settlement, isIssued, err := service.NewIssuer(run.deps).IssueOutstanding(t.Context(), partnershipID,
		countedAt().Add(time.Hour))

	if err != nil || !isIssued || settlement.Summary.Qty != 18 || settlement.Schedule != terms.ScheduleMonthly {
		t.Fatalf("settlement = %+v, err = %v", settlement, err)
	}
}
