package dbtest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

const (
	managedDatabaseHost = "neon.tech"
	suffixBytes         = 6
)

func URL(t *testing.T) string {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is empty, run tests through make test")
	}
	if strings.Contains(strings.ToLower(databaseURL), managedDatabaseHost) {
		t.Fatal("tests must never run against the managed production database")
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
	name := "trustip_test_" + randomSuffix(t)
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
