package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
)

func fieldCodes(t *testing.T, err error) map[string]string {
	t.Helper()
	appErr, ok := errors.AsType[*apperror.Error](err)
	if !ok {
		t.Fatalf("err = %v, want a validation error", err)
	}
	codes := map[string]string{}
	for _, problem := range appErr.Fields() {
		codes[problem.Field] = problem.Code
	}
	return codes
}

func TestNewQuery_FRRPT02_RangeIsJakartaDaysWithAnExclusiveEnd(t *testing.T) {
	query, err := NewQuery(Filter{From: "2026-11-01", To: "2026-11-30", GroupBy: "PRODUCT"})

	jakarta := period.Jakarta()
	if err != nil || query.GroupBy != ByProduct ||
		!query.From.Equal(time.Date(2026, 11, 1, 0, 0, 0, 0, jakarta)) ||
		!query.To.Equal(time.Date(2026, 12, 1, 0, 0, 0, 0, jakarta)) {
		t.Fatalf("query = %+v, err = %v", query, err)
	}
}

func TestNewQuery_BRRPT04_RangeIsAtMost366Days(t *testing.T) {
	_, longest := NewQuery(Filter{From: "2027-01-01", To: "2028-01-01"})
	_, tooLong := NewQuery(Filter{From: "2027-01-01", To: "2028-01-02"})

	if longest != nil || fieldCodes(t, tooLong)["to"] != "OUT_OF_RANGE" {
		t.Fatalf("366 days = %v, 367 days = %v", longest, tooLong)
	}
}

func TestNewQuery_RefusesMissingOrBackwardRanges(t *testing.T) {
	cases := map[string]Filter{
		"from":     {To: "2026-11-30"},
		"to":       {From: "2026-11-30", To: "2026-11-01"},
		"group_by": {From: "2026-11-01", To: "2026-11-30", GroupBy: "WEEK"},
	}
	for name, filter := range cases {
		_, err := NewQuery(filter)
		if _, found := fieldCodes(t, err)[name]; !found {
			t.Fatalf("%s: codes = %v", name, fieldCodes(t, err))
		}
	}
}

func TestSummarize_FRRPT04_AddsRowsAndWorksOutPercentReturned(t *testing.T) {
	rows := []Row{
		{Key: "a", Figures: Figures{SoldQty: 18, ReturnedQty: 2, ReceivedQty: 20, Split: pricing.Split{
			Gross: 36000, StoreShare: 4500, ConsignorShare: 31500,
		}}},
		{Key: "b", Figures: Figures{SoldQty: 8, ReturnedQty: 1, ReceivedQty: 10, Split: pricing.Split{
			Gross: 16000, StoreShare: 2000, ConsignorShare: 14000,
		}}},
	}

	summary := Summarize(rows)

	if summary.SoldQty != 26 || summary.ReturnedQty != 3 || summary.ReceivedQty != 30 ||
		summary.Split != (pricing.Split{Gross: 52000, StoreShare: 6500, ConsignorShare: 45500}) ||
		summary.ReturnedBPS() != 1000 {
		t.Fatalf("summary = %+v, returned = %d", summary, summary.ReturnedBPS())
	}
}

func TestReturnedBPS_NothingReceivedIsZeroAndThirdsRoundHalfUp(t *testing.T) {
	if got := (Figures{ReturnedQty: 3}).ReturnedBPS(); got != 0 {
		t.Fatalf("nothing received = %d", got)
	}
	if got := (Figures{ReturnedQty: 1, ReceivedQty: 3}).ReturnedBPS(); got != 3333 {
		t.Fatalf("one in three = %d", got)
	}
	if got := (Figures{ReturnedQty: 2, ReceivedQty: 3}).ReturnedBPS(); got != 6667 {
		t.Fatalf("two in three = %d", got)
	}
}

func TestFillDays_FRRPT05_EveryDayOfTheRangeAppearsInOrder(t *testing.T) {
	query, err := NewQuery(Filter{From: "2026-11-01", To: "2026-11-03", GroupBy: "DAY"})
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	days := FillDays(query, []Row{{Key: "2026-11-02", Label: "2026-11-02", Figures: Figures{SoldQty: 5}}})

	if len(days) != 3 || days[0].Key != "2026-11-01" || days[1].SoldQty != 5 || days[2].Key != "2026-11-03" ||
		days[2].SoldQty != 0 {
		t.Fatalf("days = %+v", days)
	}
}
