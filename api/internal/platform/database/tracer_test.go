package database

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/logger"
)

const namedQuery = "-- name: ListDrops :many\nSELECT id FROM drops WHERE organization_id = $1"

func traceQuery(t *testing.T, elapsed time.Duration) string {
	t.Helper()
	var logs bytes.Buffer
	manual := clock.NewManual(time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC))
	tracer := newSlowQueryTracer(logger.New(&logs, slog.LevelInfo), manual, slowQueryThreshold)

	start := pgx.TraceQueryStartData{SQL: namedQuery, Args: []any{"org-secret-id"}}
	ctx := tracer.TraceQueryStart(t.Context(), nil, start)
	manual.Advance(elapsed)
	tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{})

	return logs.String()
}

func TestSlowQueryTracer_QueryOverThreshold_LogsNameWithoutArguments(t *testing.T) {
	logs := traceQuery(t, 250*time.Millisecond)

	if !strings.Contains(logs, `"query":"ListDrops"`) || !strings.Contains(logs, `"duration_ms":250`) {
		t.Fatalf("logs = %s", logs)
	}
	if strings.Contains(logs, "org-secret-id") || strings.Contains(logs, "SELECT") {
		t.Fatalf("query text or arguments leaked: %s", logs)
	}
}

func TestSlowQueryTracer_QueryUnderThreshold_IsNotLogged(t *testing.T) {
	if logs := traceQuery(t, 199*time.Millisecond); logs != "" {
		t.Fatalf("logs = %s", logs)
	}
}

func TestSlowQueryTracer_QueryAtThreshold_IsLogged(t *testing.T) {
	if logs := traceQuery(t, 200*time.Millisecond); logs == "" {
		t.Fatal("query at the 200 ms threshold was not logged")
	}
}

func TestQueryName_UnnamedSQL_ReturnsUnnamed(t *testing.T) {
	if got := queryName("SELECT 1"); got != "unnamed" {
		t.Fatalf("queryName = %q", got)
	}
}
