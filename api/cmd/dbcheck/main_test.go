package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

func TestReadEnvValue_QuotedValue_ReturnsUnquoted(t *testing.T) {
	path := writeEnvFile(t, "OTHER=1\nDATABASE_MIGRATE_URL=\"postgres://u@h/db?a=1&b=2\"\n")

	value, err := readEnvValue(path, "DATABASE_MIGRATE_URL")
	if err != nil {
		t.Fatalf("readEnvValue: %v", err)
	}

	if value != "postgres://u@h/db?a=1&b=2" {
		t.Fatalf("value = %q", value)
	}
}

func TestRunWithEnvFile_MissingVariable_ReturnsMissingURL(t *testing.T) {
	path := writeEnvFile(t, "JWT_SECRET_V1=x\n")

	err := runWithEnvFile(t.Context(), &bytes.Buffer{}, []string{path})

	if !errors.Is(err, errMissingURL) {
		t.Fatalf("err = %v, want %v", err, errMissingURL)
	}
}

func TestRunWithEnvFile_NoArguments_ReturnsUsage(t *testing.T) {
	if err := runWithEnvFile(t.Context(), &bytes.Buffer{}, nil); !errors.Is(err, errUsage) {
		t.Fatalf("err = %v, want %v", err, errUsage)
	}
}

func TestRun_UnreachableHost_HidesConnectionString(t *testing.T) {
	err := run(t.Context(), &bytes.Buffer{}, "postgres://secretuser@127.0.0.1:1/secretdb?connect_timeout=1")

	if !errors.Is(err, errConnect) || strings.Contains(err.Error(), "secret") {
		t.Fatalf("err = %v, want generic connect error", err)
	}
}

func TestRun_TestDatabase_ReportsVersionAndExtensions(t *testing.T) {
	databaseURL := dbtest.URL(t)
	var out bytes.Buffer

	if err := run(t.Context(), &out, databaseURL); err != nil {
		t.Fatalf("run: %v", err)
	}

	for _, want := range []string{"server_version ", "extension pg_trgm", "extension postgis"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output %q does not contain %q", out.String(), want)
		}
	}
}

func writeEnvFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env.production")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	return path
}
