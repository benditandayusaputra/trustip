package database_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
	"github.com/benditandayusaputra/trustip/api/internal/platform/logger"
)

func openPool(t *testing.T, databaseURL string) *database.Pool {
	t.Helper()
	pool, err := database.Open(t.Context(), database.Options{
		URL:    databaseURL,
		Logger: logger.New(&bytes.Buffer{}, slog.LevelInfo),
		Clock:  clock.System{},
	})
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func transactionID(t *testing.T, ctx context.Context, pool *database.Pool) string {
	t.Helper()
	id, err := sqlc.New(pool.Querier(ctx)).GetCurrentTransactionID(ctx)
	if err != nil {
		t.Fatalf("read transaction id: %v", err)
	}
	return id
}

func TestOpen_UnparsableURL_HidesTheURL(t *testing.T) {
	_, err := database.Open(t.Context(), database.Options{URL: "postgres://user:secretpw@host:notaport/db"})

	if !errors.Is(err, database.ErrInvalidURL) || strings.Contains(err.Error(), "secretpw") {
		t.Fatalf("err = %v", err)
	}
}

func TestPing_LocalDatabase_Succeeds(t *testing.T) {
	if err := openPool(t, dbtest.URL(t)).Ping(t.Context()); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestPing_UnreachableDatabase_IsUnavailable(t *testing.T) {
	pool := openPool(t, "postgres://trustip@127.0.0.1:1/trustip?connect_timeout=1")

	err := pool.Ping(t.Context())

	if apperror.From(err).Code() != apperror.CodeServiceUnavailable {
		t.Fatalf("err = %v, want service unavailable", err)
	}
}

func TestWithinTx_QueriesInsideShareOneTransaction(t *testing.T) {
	pool := openPool(t, dbtest.URL(t))
	var first, second string

	err := pool.WithinTx(t.Context(), func(ctx context.Context) error {
		first = transactionID(t, ctx, pool)
		second = transactionID(t, ctx, pool)
		return nil
	})

	if err != nil || first != second {
		t.Fatalf("err = %v, ids = %s, %s", err, first, second)
	}
}

func TestQuerier_OutsideTransaction_UsesSeparateStatements(t *testing.T) {
	pool := openPool(t, dbtest.URL(t))

	first := transactionID(t, t.Context(), pool)
	second := transactionID(t, t.Context(), pool)

	if first == second {
		t.Fatalf("two statements outside a transaction shared id %s", first)
	}
}

func TestWithinTx_WorkFails_ReturnsWorkError(t *testing.T) {
	pool := openPool(t, dbtest.URL(t))
	workErr := errors.New("stock changed")

	err := pool.WithinTx(t.Context(), func(context.Context) error { return workErr })

	if !errors.Is(err, workErr) {
		t.Fatalf("err = %v, want %v", err, workErr)
	}
}

func TestWithinTx_Nested_JoinsOuterTransaction(t *testing.T) {
	pool := openPool(t, dbtest.URL(t))
	var outer, inner string

	err := pool.WithinTx(t.Context(), func(ctx context.Context) error {
		outer = transactionID(t, ctx, pool)
		return pool.WithinTx(ctx, func(nested context.Context) error {
			inner = transactionID(t, nested, pool)
			return nil
		})
	})

	if err != nil || outer != inner {
		t.Fatalf("err = %v, outer = %s, inner = %s", err, outer, inner)
	}
}

func TestOpen_SchemaOwnerWhenRefused_ReturnsErrSchemaOwner(t *testing.T) {
	_, err := database.Open(t.Context(), database.Options{
		URL:               dbtest.URL(t),
		Logger:            logger.New(&bytes.Buffer{}, slog.LevelInfo),
		Clock:             clock.System{},
		RefuseSchemaOwner: true,
	})

	if !errors.Is(err, database.ErrSchemaOwner) {
		t.Fatalf("err = %v, want ErrSchemaOwner", err)
	}
}

func TestAfterCommit_InsideTransaction_RunsOnlyAfterCommit(t *testing.T) {
	pool := openPool(t, dbtest.URL(t))
	ran := false

	err := pool.WithinTx(t.Context(), func(ctx context.Context) error {
		pool.AfterCommit(ctx, func() { ran = true })
		if ran {
			t.Error("hook ran before commit")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("within tx: %v", err)
	}

	if !ran {
		t.Fatal("hook did not run after commit")
	}
}

func TestAfterCommit_TransactionRolledBack_NeverRuns(t *testing.T) {
	pool := openPool(t, dbtest.URL(t))
	ran := false

	err := pool.WithinTx(t.Context(), func(ctx context.Context) error {
		pool.AfterCommit(ctx, func() { ran = true })
		return errors.New("stock changed")
	})

	if err == nil || ran {
		t.Fatalf("err = %v, ran = %t", err, ran)
	}
}

func TestAfterCommit_OutsideTransaction_RunsAtOnce(t *testing.T) {
	pool := openPool(t, dbtest.URL(t))
	ran := false

	pool.AfterCommit(t.Context(), func() { ran = true })

	if !ran {
		t.Fatal("hook did not run")
	}
}
