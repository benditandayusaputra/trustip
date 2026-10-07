package database

import (
	"context"
	"log/slog"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
)

var queryNamePattern = regexp.MustCompile(`^-- name: (\w+)`)

type queryStartKey struct{}

type queryStart struct {
	at   time.Time
	name string
}

type slowQueryTracer struct {
	log       *slog.Logger
	clock     Clock
	threshold time.Duration
}

func newSlowQueryTracer(log *slog.Logger, clock Clock, threshold time.Duration) *slowQueryTracer {
	return &slowQueryTracer{log: log, clock: clock, threshold: threshold}
}

func (t *slowQueryTracer) TraceQueryStart(
	ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData,
) context.Context {
	return context.WithValue(ctx, queryStartKey{}, queryStart{at: t.clock.Now(), name: queryName(data.SQL)})
}

func (t *slowQueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryEndData) {
	start, ok := ctx.Value(queryStartKey{}).(queryStart)
	if !ok {
		return
	}
	elapsed := t.clock.Now().Sub(start.at)
	if elapsed < t.threshold {
		return
	}
	t.log.WarnContext(ctx, "slow query", "query", start.name, "duration_ms", elapsed.Milliseconds())
}

func queryName(sql string) string {
	match := queryNamePattern.FindStringSubmatch(sql)
	if match == nil {
		return "unnamed"
	}
	return match[1]
}
