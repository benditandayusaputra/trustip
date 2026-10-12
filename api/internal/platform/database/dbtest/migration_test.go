package dbtest

import (
	"context"
	"database/sql/driver"
	"errors"
	"io"
	"testing"

	"github.com/pressly/goose/v3"
)

type migrationScenario struct {
	name     string
	failures []error
	want     error
	calls    int
}

func TestMigrate_RTEST01_RetriesOnlyLostConnections(t *testing.T) {
	permanent := errors.New("invalid migration")
	wrapped := errors.Join(errors.New("partial migration"), driver.ErrBadConn)
	scenarios := []migrationScenario{
		{"bad connection recovers", []error{wrapped, nil}, nil, 2},
		{"unexpected EOF recovers", []error{io.ErrUnexpectedEOF, nil}, nil, 2},
		{"EOF recovers", []error{io.EOF, nil}, nil, 2},
		{"permanent error stops", []error{permanent, nil}, permanent, 1},
		{"retry is bounded", []error{driver.ErrBadConn, driver.ErrBadConn}, driver.ErrBadConn, 2},
		{"cancellation stops", []error{context.Canceled, nil}, context.Canceled, 1},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			requireMigrationResult(t, scenario)
		})
	}
}

func requireMigrationResult(t *testing.T, scenario migrationScenario) {
	t.Helper()
	calls := 0
	err := migrate(t.Context(), func(context.Context) ([]*goose.MigrationResult, error) {
		calls++
		return nil, scenario.failures[calls-1]
	})
	if !errors.Is(err, scenario.want) || calls != scenario.calls {
		t.Fatalf("migration error = %v, calls = %d; want error = %v, calls = %d",
			err, calls, scenario.want, scenario.calls)
	}
}
