package domain

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

func TestNewDailyQuery_FRRPT01_DefaultsToTodayInJakarta(t *testing.T) {
	now := time.Date(2026, 11, 19, 18, 0, 0, 0, time.UTC)

	query, err := NewDailyQuery("", now)
	chosen, chosenErr := NewDailyQuery("2026-11-18", now)

	jakarta := period.Jakarta()
	if err != nil || query.Day != "2026-11-20" || !query.From.Equal(time.Date(2026, 11, 20, 0, 0, 0, 0, jakarta)) ||
		!query.To.Equal(time.Date(2026, 11, 21, 0, 0, 0, 0, jakarta)) {
		t.Fatalf("query = %+v, err = %v", query, err)
	}
	if chosenErr != nil || chosen.Day != "2026-11-18" {
		t.Fatalf("chosen = %+v, err = %v", chosen, chosenErr)
	}
	if _, badErr := NewDailyQuery("18-11-2026", now); fieldCodes(t, badErr)["date"] != "INVALID_DATE" {
		t.Fatalf("bad date = %v", badErr)
	}
}

func TestSummarizeRecap_FRRPT01_AddsEveryConsignor(t *testing.T) {
	rows := []RecapRow{
		{Name: "A", Figures: Figures{SoldQty: 18}, Paid: 31500},
		{Name: "B", Figures: Figures{SoldQty: 8}, Paid: 0},
		{Name: "C", Figures: Figures{ReceivedQty: 20}, Paid: 0},
	}

	total := SummarizeRecap(rows)

	if total.SoldQty != 26 || total.ReceivedQty != 20 || total.Paid != 31500 {
		t.Fatalf("total = %+v", total)
	}
}

func TestSettlementTally_FRRPT06_AddsCountsAndAmounts(t *testing.T) {
	rows := []SettlementRow{
		{Issued: Tally{Count: 2, Amount: 45500}, Paid: Tally{Count: 1, Amount: 31500}, Unpaid: Tally{
			Count:  1,
			Amount: 14000,
		}, Overdue: Tally{Count: 1, Amount: 14000}},
		{Issued: Tally{Count: 1, Amount: 1750}, Unpaid: Tally{Count: 1, Amount: 1750}},
	}

	total := SummarizeSettlements(rows)

	if total.Issued != (Tally{Count: 3, Amount: 47250}) || total.Paid != (Tally{Count: 1, Amount: 31500}) ||
		total.Unpaid != (Tally{Count: 2, Amount: 15750}) || total.Overdue != (Tally{Count: 1, Amount: 14000}) {
		t.Fatalf("total = %+v", total)
	}
}
