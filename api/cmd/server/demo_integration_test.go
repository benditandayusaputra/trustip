package main

import (
	"os"
	"path/filepath"
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

func TestDemoSeed_Spec25_SeedsThirtyDaysThenResetLeavesNoDemoRows(t *testing.T) {
	pool := dbtest.MigratedPool(t)
	now := time.Date(2026, 10, 9, 3, 0, 0, 0, time.UTC)

	fileDir := t.TempDir()
	seed := resetDemo(pool, demoSecrets{peppers: testPepper{}, sealKeys: testSealKeys{}, fileDir: fileDir})
	if err := seed(t.Context(), now); err != nil {
		t.Fatalf("seed: %v", err)
	}
	seeded := readDemoShape(t, pool, now)
	photos, err := os.ReadDir(filepath.Join(fileDir, "public"))
	if err != nil {
		t.Fatalf("read stored photos: %v", err)
	}
	proofs, err := os.ReadDir(filepath.Join(fileDir, "private"))
	if err != nil {
		t.Fatalf("read stored proofs: %v", err)
	}
	if err := sqlc.New(pool.Querier(t.Context())).ResetDemoData(t.Context()); err != nil {
		t.Fatalf("reset: %v", err)
	}
	reset := readDemoShape(t, pool, now)

	checks := map[string]bool{
		"8 active partnerships": seeded.ActivePartnerships == 8, "1 ended partnership": seeded.EndedPartnerships == 1,
		"5 open needs": seeded.OpenNeeds == 5, "1 fulfilled need": seeded.FulfilledNeeds == 1,
		"2 open proposals": seeded.OpenProposals == 2, "1 declined proposal": seeded.DeclinedProposals == 1,
		"paid settlements": seeded.PaidSettlements > 0, "pending payments": seeded.PendingPayments > 0,
		"overdue settlements":          seeded.OverdueSettlements > 0,
		"unpaid but not overdue":       seeded.UnpaidSettlements > seeded.OverdueSettlements,
		"28 demo organizations":        seeded.DemoOrganizations == 28,
		"files stored once each":       len(proofs) > 0 && int64(len(photos)+len(proofs)) == seeded.DemoFiles,
		"reset removes every demo row": reset == sqlc.GetDemoShapeRow{},
	}
	for name, holds := range checks {
		if !holds {
			t.Errorf("%s: seeded = %+v, reset = %+v", name, seeded, reset)
		}
	}
}
