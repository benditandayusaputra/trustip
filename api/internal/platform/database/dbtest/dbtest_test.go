package dbtest_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

func TestTemporaryDatabase_RDB03_RemovesDatabaseAfterTestCleanup(t *testing.T) {
	admin := dbtest.Connect(t, dbtest.URL(t))
	var name string
	t.Run("temporary database", func(t *testing.T) {
		parsed, err := url.Parse(dbtest.TemporaryDatabase(t))
		if err != nil {
			t.Fatal(err)
		}
		name = strings.TrimPrefix(parsed.Path, "/")
		if !strings.HasPrefix(name, dbtest.TemporaryPrefix) {
			t.Fatalf("unsafe temporary database name: %s", name)
		}
		if !databaseExists(t, admin, name) {
			t.Fatal("temporary database was not created")
		}
	})
	if databaseExists(t, admin, name) {
		t.Fatal("temporary database remains after cleanup")
	}
}

func databaseExists(t *testing.T, conn *pgx.Conn, name string) bool {
	t.Helper()
	var exists bool
	if err := conn.QueryRow(t.Context(),
		"SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	return exists
}
