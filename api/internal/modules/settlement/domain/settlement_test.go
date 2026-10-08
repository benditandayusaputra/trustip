package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

const (
	partnershipID = "0192f0c4-0000-7000-8000-000000000201"
	storeID       = "0192f0c4-0000-7000-8000-000000000401"
	consignorID   = "0192f0c4-0000-7000-8000-000000000402"
)

type sale struct {
	product string
	terms   pricing.Terms
	qty     int
	at      time.Time
}

func allocationsOf(sales ...sale) []Allocation {
	allocations := make([]Allocation, 0, len(sales))
	for index, sold := range sales {
		allocations = append(allocations, Allocation{
			ID: fmt.Sprintf("sale-%d", index+1), ProductID: sold.product, OccurredAt: sold.at, Qty: sold.qty,
			Terms: sold.terms, Split: sold.terms.SplitOf(sold.qty),
		})
	}
	return allocations
}

func sectionThirteen(at time.Time) []Allocation {
	return allocationsOf(
		sale{"keripik", pricing.Terms{Model: pricing.Commission, RetailPrice: 15000, CommissionBPS: 2000}, 27, at},
		sale{"kue-bawang", pricing.Terms{Model: pricing.Commission, RetailPrice: 22500, CommissionBPS: 1750}, 3, at},
		sale{"sambal", pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 35000, DepositPrice: 28000}, 10, at},
		sale{"roti", pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1750}, 18, at},
	)
}

func TestSummarize_BRSTL04_SectionThirteenExampleToTheRupiah(t *testing.T) {
	summary := Summarize(sectionThirteen(jakarta("2026-11-10 12:00")))

	want := []pricing.Split{
		{Gross: 405000, StoreShare: 81000, ConsignorShare: 324000},
		{Gross: 67500, StoreShare: 11813, ConsignorShare: 55687},
		{Gross: 350000, StoreShare: 70000, ConsignorShare: 280000},
		{Gross: 36000, StoreShare: 4500, ConsignorShare: 31500},
	}
	if len(summary.Lines) != len(want) {
		t.Fatalf("lines = %+v", summary.Lines)
	}
	for index, line := range summary.Lines {
		if line.Split != want[index] {
			t.Fatalf("line %d = %+v, want %+v", index, line.Split, want[index])
		}
	}
	total := pricing.Split{Gross: 858500, StoreShare: 167313, ConsignorShare: 691187}
	if summary.Qty != 58 || summary.Split != total {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestSummarize_FRSTL04_OneLinePerProductAndTerms(t *testing.T) {
	bread := pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1750}
	cheaper := pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1500}
	at := jakarta("2026-11-10 12:00")

	summary := Summarize(allocationsOf(
		sale{"roti", bread, 5, at}, sale{"roti", bread, 3, at}, sale{"roti", cheaper, 2, at},
	))

	if len(summary.Lines) != 2 || summary.Lines[0].Qty != 8 || summary.Lines[0].Split.ConsignorShare != 14000 ||
		summary.Lines[1].Qty != 2 || summary.Lines[1].Split.ConsignorShare != 3000 {
		t.Fatalf("lines = %+v", summary.Lines)
	}
}

func aLedger(schedule terms.Schedule, allocations []Allocation) Ledger {
	return Ledger{
		PartnershipID: partnershipID, ConsignorID: consignorID, StoreID: storeID, Schedule: schedule,
		PaymentTermsDays: 7, Unsettled: allocations,
	}
}

func TestDraftUpTo_FRSTL01_TakesEverySaleUpToTheCount(t *testing.T) {
	bread := pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1750}
	counted := jakarta("2026-11-10 20:15")
	ledger := aLedger(terms.ScheduleOnCount, allocationsOf(
		sale{"roti", bread, 4, counted.Add(-time.Hour)}, sale{"roti", bread, 18, counted},
		sale{"roti", bread, 1, counted.Add(time.Second)},
	))

	draft, hasSales := ledger.DraftUpTo(counted)

	want := Period{Start: counted.Add(-time.Hour), End: counted.Add(time.Second)}
	if !hasSales || !draft.Period.Equal(want) || len(draft.Allocations) != 2 || draft.Summary.Qty != 22 {
		t.Fatalf("draft = %+v", draft)
	}
}

func TestDraftUpTo_BRSTL02_StartsWhereTheLastSettlementEnded(t *testing.T) {
	bread := pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1750}
	counted := jakarta("2026-11-10 20:15")
	ledger := aLedger(terms.ScheduleOnCount, allocationsOf(sale{"roti", bread, 4, counted}))
	ledger.SettledUntil = jakarta("2026-11-09")

	draft, hasSales := ledger.DraftUpTo(counted)

	if !hasSales || !draft.Period.Start.Equal(ledger.SettledUntil) {
		t.Fatalf("draft = %+v", draft)
	}
}

func TestDraftUpTo_BRSTL03_NoSalesNoSettlement(t *testing.T) {
	_, hasSales := aLedger(terms.ScheduleOnCount, nil).DraftUpTo(jakarta("2026-11-10 20:15"))
	if hasSales {
		t.Fatal("a count without sales must not issue a settlement")
	}
}

func TestDraftClosedPeriods_FRSTL02_OneDraftPerEndedWeekAndNoneForTheCurrentWeek(t *testing.T) {
	bread := pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1750}
	ledger := aLedger(terms.ScheduleWeekly, allocationsOf(
		sale{"roti", bread, 3, jakarta("2026-11-03 10:00")},
		sale{"roti", bread, 5, jakarta("2026-11-13 10:00")},
		sale{"roti", bread, 2, jakarta("2026-11-15 23:59")},
		sale{"roti", bread, 7, jakarta("2026-11-16")},
	))

	drafts, err := ledger.DraftClosedPeriods(jakarta("2026-11-16 00:10"))

	if err != nil || len(drafts) != 2 {
		t.Fatalf("drafts = %+v, err = %v", drafts, err)
	}
	first, second := drafts[0], drafts[1]
	if !first.Period.Equal(Period{jakarta("2026-11-02"), jakarta("2026-11-09")}) || first.Summary.Qty != 3 {
		t.Fatalf("first = %+v", first)
	}
	secondWeek := Period{jakarta("2026-11-09"), jakarta("2026-11-16")}
	if !second.Period.Equal(secondWeek) || second.Summary.Qty != 7 {
		t.Fatalf("second = %+v", second)
	}
}

func TestDraftClosedPeriods_BRSTL02_ClipsAWeekThatAnEarlierSettlementPartlyCovered(t *testing.T) {
	bread := pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1750}
	ledger := aLedger(terms.ScheduleWeekly, allocationsOf(sale{"roti", bread, 5, jakarta("2026-11-13 10:00")}))
	ledger.SettledUntil = jakarta("2026-11-11 18:00")

	drafts, err := ledger.DraftClosedPeriods(jakarta("2026-11-16 00:10"))

	if err != nil || len(drafts) != 1 ||
		!drafts[0].Period.Equal(Period{jakarta("2026-11-11 18:00"), jakarta("2026-11-16")}) {
		t.Fatalf("drafts = %+v, err = %v", drafts, err)
	}
}

func TestDraftClosedPeriods_FRSTL02_MonthlyWaitsForTheMonthToEnd(t *testing.T) {
	bread := pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1750}
	ledger := aLedger(terms.ScheduleMonthly, allocationsOf(sale{"roti", bread, 5, jakarta("2026-11-30 23:00")}))

	before, beforeErr := ledger.DraftClosedPeriods(jakarta("2026-11-30 23:59"))
	after, afterErr := ledger.DraftClosedPeriods(jakarta("2026-12-01 00:10"))

	if beforeErr != nil || afterErr != nil || len(before) != 0 || len(after) != 1 ||
		!after[0].Period.Equal(Period{jakarta("2026-11-01"), jakarta("2026-12-01")}) {
		t.Fatalf("before = %+v, after = %+v", before, after)
	}
}

func TestIssue_FRSTL03_NumberAndDueDateFollowTheIssueDayInJakarta(t *testing.T) {
	draft, _ := aLedger(terms.ScheduleOnCount, sectionThirteen(jakarta("2026-11-30 20:00"))).
		DraftUpTo(jakarta("2026-11-30 20:00"))

	issuedAt := time.Date(2026, 11, 30, 17, 30, 0, 0, time.UTC)

	settlement := draft.Issue(Issuing{ID: "settlement-1", Sequence: 12, At: issuedAt})

	if settlement.Number != "ST-202612-0012" || settlement.Status != Issued ||
		!settlement.DueDate.Equal(jakarta("2026-12-08")) || len(settlement.Allocations) != 4 ||
		settlement.Summary.Split.ConsignorShare != 691187 {
		t.Fatalf("settlement = %+v", settlement)
	}
}

func late(days int, age Age) Standing {
	return Standing{IsOverdue: true, DaysLate: days, Age: age}
}

func TestStandingOf_BRSTL09_OverdueIsDerivedFromTheJakartaDate(t *testing.T) {
	due := jakarta("2026-11-10")
	cases := []struct {
		name   string
		status Status
		now    time.Time
		want   Standing
	}{
		{"on the due date", Issued, jakarta("2026-11-10 23:59"), Standing{Age: AgeNotDue}},
		{"one day late", Issued, jakarta("2026-11-11"), late(1, AgeLate1To7)},
		{"seven days late", Issued, jakarta("2026-11-17 08:00"), late(7, AgeLate1To7)},
		{"eight days late", Issued, jakarta("2026-11-18 08:00"), late(8, AgeLate8To30)},
		{"thirty days late", Issued, jakarta("2026-12-10 08:00"), late(30, AgeLate8To30)},
		{"thirty one days late", Issued, jakarta("2026-12-11 08:00"), late(31, AgeLateOver30)},
		{"paid late is not overdue", Paid, jakarta("2026-12-11 08:00"), Standing{Age: AgePaid}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := StandingOf(tc.status, due, tc.now); got != tc.want {
				t.Fatalf("standing = %+v, want %+v", got, tc.want)
			}
		})
	}
}
