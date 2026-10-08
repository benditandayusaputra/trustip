package repository_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	loadStoreCount    = 10_000
	loadSpread        = 0.5
	nearestBudget     = 200 * time.Millisecond
	measuredRuns      = 5
	spatialIndex      = "organizations_location_public_idx"
	semarangLatitude  = -6.99
	semarangLongitude = 110.42
)

type planLines struct {
	mu    sync.Mutex
	lines []string
}

func (p *planLines) collect(_ *pgconn.PgConn, notice *pgconn.Notice) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lines = append(p.lines, notice.Message)
}

func (p *planLines) text() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return strings.Join(p.lines, "\n")
}

func seededDirectory(t *testing.T) (string, *repository.Directory) {
	t.Helper()
	databaseURL := dbtest.MigratedDatabase(t)
	pool := dbtest.PoolFor(t, databaseURL)
	queries := sqlc.New(pool.Querier(t.Context()))
	err := queries.SeedListedStores(t.Context(), sqlc.SeedListedStoresParams{
		Longitude: semarangLongitude, Latitude: semarangLatitude, SpreadDegrees: loadSpread, StoreCount: loadStoreCount,
	})
	if err != nil {
		t.Fatalf("seed stores: %v", err)
	}
	if err := queries.AnalyzeOrganizations(t.Context()); err != nil {
		t.Fatalf("analyze: %v", err)
	}
	return databaseURL, repository.NewDirectory(pool)
}

func nearestPage(ctx context.Context, directory *repository.Directory) error {
	query := domain.StoreQuery{Listing: domain.Listing{
		ViewerID: consignorID, Sort: domain.Nearest, RadiusKm: domain.DefaultRadiusKm, Limit: 21,
		Origin: domain.Origin{
			Point: domain.Point{Latitude: semarangLatitude, Longitude: semarangLongitude}, IsKnown: true,
		},
	}}
	if _, err := directory.CountStoresWithin(ctx, query); err != nil {
		return fmt.Errorf("count nearest stores: %w", err)
	}
	if _, err := directory.ListStores(ctx, query); err != nil {
		return fmt.Errorf("list nearest stores: %w", err)
	}
	return nil
}

func TestNearestStores_NFR03_FirstPageOnTenThousandOrganizationsStaysUnderBudget(t *testing.T) {
	_, directory := seededDirectory(t)
	ctx := t.Context()
	if err := nearestPage(ctx, directory); err != nil {
		t.Fatalf("warm up: %v", err)
	}
	var failure error

	result := testing.Benchmark(func(b *testing.B) {
		for b.Loop() {
			if err := nearestPage(ctx, directory); err != nil {
				failure = err
			}
		}
	})

	perPage := time.Duration(result.NsPerOp())
	t.Logf("nearest first page on %d stores: %v per page over %d runs", loadStoreCount, perPage, result.N)
	if failure != nil || perPage >= nearestBudget {
		t.Fatalf("per page %v, budget %v, err = %v", perPage, nearestBudget, failure)
	}
}

func TestNearestStores_RPERF02_PlanUsesTheSpatialIndex(t *testing.T) {
	databaseURL, _ := seededDirectory(t)
	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	plan := &planLines{}
	config.OnNotice = plan.collect
	conn, err := pgx.ConnectConfig(t.Context(), config)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(context.WithoutCancel(t.Context())); err != nil {
			t.Errorf("close: %v", err)
		}
	})

	if err := sqlc.New(conn).ReportNearestStoresPlan(t.Context()); err != nil {
		t.Fatalf("explain: %v", err)
	}

	t.Logf("plan:\n%s", plan.text())
	if !strings.Contains(plan.text(), spatialIndex) {
		t.Fatalf("plan does not use %s:\n%s", spatialIndex, plan.text())
	}
}
