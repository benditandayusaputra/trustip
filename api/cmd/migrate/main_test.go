package main

import (
	"bytes"
	"errors"
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/db/migrations"
	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

func TestRun_NoArguments_ReturnsUsage(t *testing.T) {
	err := run(t.Context(), &bytes.Buffer{}, nil, "postgres://localhost/trustip")
	if !errors.Is(err, errUsage) {
		t.Fatalf("err = %v, want %v", err, errUsage)
	}
}

func TestRun_EmptyURL_ReturnsMissingURL(t *testing.T) {
	err := run(t.Context(), &bytes.Buffer{}, []string{"up"}, "")
	if !errors.Is(err, errMissingURL) {
		t.Fatalf("err = %v, want %v", err, errMissingURL)
	}
}

func TestRun_DownOnManagedDatabase_IsRefused(t *testing.T) {
	err := run(t.Context(), &bytes.Buffer{}, []string{"down"}, "postgres://u@ep-x.ap-southeast-1.aws.neon.tech/db")
	if !errors.Is(err, errDownOnManaged) {
		t.Fatalf("err = %v, want %v", err, errDownOnManaged)
	}
}

func TestRun_UnknownCommand_ReturnsUsage(t *testing.T) {
	err := run(t.Context(), &bytes.Buffer{}, []string{"redo"}, dbtest.TemporaryDatabase(t))
	if !errors.Is(err, errUsage) {
		t.Fatalf("err = %v, want %v", err, errUsage)
	}
}

func TestRun_UpAllDownUp_InstallsAndRemovesExtensions(t *testing.T) {
	databaseURL := dbtest.TemporaryDatabase(t)

	migrate(t, databaseURL, "up")
	afterUp := installedExtensions(t, databaseURL)
	for range migrationCount(t) {
		migrate(t, databaseURL, "down")
	}
	afterDown := installedExtensions(t, databaseURL)
	migrate(t, databaseURL, "up")

	if !slices.Equal(afterUp, []string{"pg_trgm", "postgis"}) {
		t.Fatalf("after up = %v, want pg_trgm and postgis", afterUp)
	}
	if len(afterDown) != 0 {
		t.Fatalf("after down = %v, want none", afterDown)
	}
}

func TestRun_Status_ListsFirstMigration(t *testing.T) {
	databaseURL := dbtest.TemporaryDatabase(t)
	var out bytes.Buffer

	if err := run(t.Context(), &out, []string{"status"}, databaseURL); err != nil {
		t.Fatalf("status: %v", err)
	}

	if !strings.Contains(out.String(), "00001 pending 00001_extensions.sql") {
		t.Fatalf("status output = %q", out.String())
	}
}

func migrate(t *testing.T, databaseURL, command string) {
	t.Helper()
	if err := run(t.Context(), &bytes.Buffer{}, []string{command}, databaseURL); err != nil {
		t.Fatalf("migrate %s: %v", command, err)
	}
}

func migrationCount(t *testing.T) int {
	t.Helper()
	files, err := fs.Glob(migrations.Files, "*.sql")
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}
	return len(files)
}

func installedExtensions(t *testing.T, databaseURL string) []string {
	t.Helper()
	queries := sqlc.New(dbtest.Connect(t, databaseURL))
	names, err := queries.ListInstalledExtensions(t.Context(), []string{"pg_trgm", "postgis"})
	if err != nil {
		t.Fatalf("list installed extensions: %v", err)
	}
	return names
}
