package main

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

func readDemoShape(t *testing.T, pool *database.Pool, today time.Time) sqlc.GetDemoShapeRow {
	t.Helper()
	shape, err := sqlc.New(pool.Querier(t.Context())).GetDemoShape(t.Context(), today.Format(time.DateOnly))
	if err != nil {
		t.Fatalf("read demo shape: %v", err)
	}
	return shape
}

func TestDemoSeed_Spec25_ResetRebuildsThirtyDaysOfDemoHistory(t *testing.T) {
	pool := dbtest.MigratedPool(t)
	now := time.Date(2026, 10, 9, 3, 0, 0, 0, time.UTC)
	seed := resetDemo(pool, demoSecrets{peppers: testPepper{}, sealKeys: testSealKeys{}})

	if err := seed(t.Context(), now); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	first := readDemoShape(t, pool, now)
	if err := seed(t.Context(), now); err != nil {
		t.Fatalf("second seed: %v", err)
	}
	second := readDemoShape(t, pool, now)

	if first != second {
		t.Fatalf("first = %+v\nsecond = %+v", first, second)
	}
	checks := map[string]bool{
		"8 active partnerships": first.ActivePartnerships == 8, "1 ended partnership": first.EndedPartnerships == 1,
		"5 open needs": first.OpenNeeds == 5, "1 fulfilled need": first.FulfilledNeeds == 1,
		"2 open proposals": first.OpenProposals == 2, "1 declined proposal": first.DeclinedProposals == 1,
		"paid settlements": first.PaidSettlements > 0, "pending payments": first.PendingPayments > 0,
		"overdue settlements":    first.OverdueSettlements > 0,
		"unpaid but not overdue": first.UnpaidSettlements > first.OverdueSettlements,
		"28 demo organizations":  first.DemoOrganizations == 28,
	}
	for name, holds := range checks {
		if !holds {
			t.Errorf("%s: shape = %+v", name, first)
		}
	}
}
