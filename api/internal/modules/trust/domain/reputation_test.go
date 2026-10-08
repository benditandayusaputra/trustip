package domain

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

func TestOnTimeBPS_BRTRS01_ShareOfDueSettlementsPaidOnTime(t *testing.T) {
	cases := []struct {
		name         string
		onTime, due  int
		wantBasisPts int
	}{
		{"all on time", 3, 3, 10000},
		{"two of three rounds half up", 2, 3, 6667},
		{"none due yet", 0, 0, 0},
		{"none on time", 0, 4, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := OnTimeBPS(tc.onTime, tc.due); got != tc.wantBasisPts {
				t.Fatalf("on time = %d, want %d", got, tc.wantBasisPts)
			}
		})
	}
}

func TestWindowAt_BRTRS01_LooksBack180JakartaDays(t *testing.T) {
	lateNightUTC := time.Date(2026, 11, 10, 18, 0, 0, 0, time.UTC)

	got := WindowAt(lateNightUTC)

	if got.Today != "2026-11-11" || got.Since != "2026-05-15" {
		t.Fatalf("window = %+v, want 2026-05-15 until 2026-11-11", got)
	}
}

func TestSummarize_BRTRS02_KeepsCountsWithoutMoney(t *testing.T) {
	facts := []Facts{{
		OrganizationID: "store-1", ActivePartners: 4, DueSettlements: 3, OnTimeSettlements: 3, Products: 0,
	}}

	got := Summarize(facts)

	want := Stats{OrganizationID: "store-1", ActivePartners: 4, OnTimeBPS: 10000, SettlementsCounted: 3}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("stats = %+v, want %+v", got, want)
	}
}

func TestWindowAt_UsesJakartaDates(t *testing.T) {
	morning := time.Date(2026, 11, 11, 6, 0, 0, 0, period.Jakarta())

	if got := WindowAt(morning).Today; got != "2026-11-11" {
		t.Fatalf("today = %s", got)
	}
}
