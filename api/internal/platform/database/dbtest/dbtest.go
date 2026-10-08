package dbtest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"io"
	"log/slog"
	"net/url"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/benditandayusaputra/trustip/api/internal/db/migrations"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/logger"
)

const (
	TemporaryPrefix = "trustip_test_"
	suffixBytes     = 6
)

func URL(t *testing.T) string {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is empty, run tests through make test")
	}
	return databaseURL
}

func Connect(t *testing.T, databaseURL string) *pgx.Conn {
	t.Helper()
	conn, err := pgx.Connect(t.Context(), databaseURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(context.WithoutCancel(t.Context())); err != nil {
			t.Errorf("close test connection: %v", err)
		}
	})
	return conn
}

func TemporaryDatabase(t *testing.T) string {
	t.Helper()
	baseURL := URL(t)
	name := TemporaryPrefix + randomSuffix(t)
	admin := Connect(t, baseURL)
	identifier := pgx.Identifier{name}.Sanitize()
	if _, err := admin.Exec(t.Context(), "CREATE DATABASE "+identifier); err != nil {
		t.Fatalf("create temporary database: %v", err)
	}
	t.Cleanup(func() {
		ctx := context.WithoutCancel(t.Context())
		if _, err := admin.Exec(ctx, "DROP DATABASE "+identifier+" WITH (FORCE)"); err != nil {
			t.Errorf("drop temporary database: %v", err)
		}
	})
	return withDatabaseName(t, baseURL, name)
}

func withDatabaseName(t *testing.T, databaseURL, name string) string {
	t.Helper()
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse database url: %v", err)
	}
	parsed.Path = "/" + name
	return parsed.String()
}

func randomSuffix(t *testing.T) string {
	t.Helper()
	buffer := make([]byte, suffixBytes)
	if _, err := rand.Read(buffer); err != nil {
		t.Fatalf("random suffix: %v", err)
	}
	return hex.EncodeToString(buffer)
}

func MigratedDatabase(t *testing.T) string {
	t.Helper()
	databaseURL := TemporaryDatabase(t)
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open temporary database: %v", err)
	}
	defer closeDB(t, db)
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.Files)
	if err != nil {
		t.Fatalf("prepare migrations: %v", err)
	}
	if _, err := provider.Up(t.Context()); err != nil {
		t.Fatalf("migrate temporary database: %v", err)
	}
	return databaseURL
}

func closeDB(t *testing.T, db *sql.DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Errorf("close migration connection: %v", err)
	}
}

func MigratedPool(t *testing.T) *database.Pool {
	t.Helper()
	return PoolFor(t, MigratedDatabase(t))
}

func PoolFor(t *testing.T, databaseURL string) *database.Pool {
	t.Helper()
	pool, err := database.Open(t.Context(), database.Options{
		URL:    databaseURL,
		Logger: logger.New(io.Discard, slog.LevelInfo),
		Clock:  clock.System{},
	})
	if err != nil {
		t.Fatalf("open migrated pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
