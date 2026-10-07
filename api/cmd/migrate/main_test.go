package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
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
	err := run(t.Context(), &bytes.Buffer{}, []string{"redo"}, temporaryDatabase(t))
	if !errors.Is(err, errUsage) {
		t.Fatalf("err = %v, want %v", err, errUsage)
	}
}

func TestRun_UpDownUp_InstallsAndRemovesExtensions(t *testing.T) {
	databaseURL := temporaryDatabase(t)

	migrate(t, databaseURL, "up")
	afterUp := installedExtensions(t, databaseURL)
	migrate(t, databaseURL, "down")
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
	databaseURL := temporaryDatabase(t)
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

func installedExtensions(t *testing.T, databaseURL string) []string {
	t.Helper()
	queries := sqlc.New(connect(t, databaseURL))
	names, err := queries.ListInstalledExtensions(t.Context(), []string{"pg_trgm", "postgis"})
	if err != nil {
		t.Fatalf("list installed extensions: %v", err)
	}
	return names
}

func temporaryDatabase(t *testing.T) string {
	t.Helper()
	baseURL := testDatabaseURL(t)
	name := "trustip_test_" + randomSuffix(t)
	admin := connect(t, baseURL)
	identifier := pgx.Identifier{name}.Sanitize()
	if _, err := admin.Exec(t.Context(), "CREATE DATABASE "+identifier); err != nil {
		t.Fatalf("create database: %v", err)
	}
	t.Cleanup(func() {
		ctx := context.WithoutCancel(t.Context())
		if _, err := admin.Exec(ctx, "DROP DATABASE "+identifier+" WITH (FORCE)"); err != nil {
			t.Errorf("drop database: %v", err)
		}
	})
	return withDatabaseName(t, baseURL, name)
}

func testDatabaseURL(t *testing.T) string {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is empty, run tests through make test")
	}
	if isManagedDatabase(databaseURL) {
		t.Fatal("tests must never run against the managed production database")
	}
	return databaseURL
}

func connect(t *testing.T, databaseURL string) *pgx.Conn {
	t.Helper()
	conn, err := pgx.Connect(t.Context(), databaseURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(context.WithoutCancel(t.Context())); err != nil {
			t.Errorf("close: %v", err)
		}
	})
	return conn
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
	buffer := make([]byte, 6)
	if _, err := rand.Read(buffer); err != nil {
		t.Fatalf("random suffix: %v", err)
	}
	return hex.EncodeToString(buffer)
}
