package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/benditandayusaputra/trustip/api/internal/db/migrations"
)

const managedDatabaseHost = "neon.tech"

var (
	errUsage         = errors.New("usage: migrate up|down|status")
	errMissingURL    = errors.New("DATABASE_MIGRATE_URL is empty")
	errDownOnManaged = errors.New("refusing to migrate down on the managed production database")
)

func main() {
	err := run(context.Background(), os.Stdout, os.Args[1:], os.Getenv("DATABASE_MIGRATE_URL"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer, args []string, databaseURL string) error {
	if len(args) != 1 {
		return errUsage
	}
	if databaseURL == "" {
		return errMissingURL
	}
	if args[0] == "down" && isManagedDatabase(databaseURL) {
		return errDownOnManaged
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	return errors.Join(execute(ctx, out, db, args[0]), db.Close())
}

func isManagedDatabase(databaseURL string) bool {
	return strings.Contains(strings.ToLower(databaseURL), managedDatabaseHost)
}

func execute(ctx context.Context, out io.Writer, db *sql.DB, command string) error {
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.Files)
	if err != nil {
		return fmt.Errorf("prepare migrations: %w", err)
	}
	switch command {
	case "up":
		results, err := provider.Up(ctx)
		return reportResults(out, results, err)
	case "down":
		result, err := provider.Down(ctx)
		return reportResults(out, []*goose.MigrationResult{result}, err)
	case "status":
		statuses, err := provider.Status(ctx)
		return reportStatus(out, statuses, err)
	}
	return errUsage
}

func reportResults(out io.Writer, results []*goose.MigrationResult, err error) error {
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	for _, result := range results {
		if _, err := fmt.Fprintln(out, result.String()); err != nil {
			return fmt.Errorf("write result: %w", err)
		}
	}
	return nil
}

func reportStatus(out io.Writer, statuses []*goose.MigrationStatus, err error) error {
	if err != nil {
		return fmt.Errorf("read migration status: %w", err)
	}
	for _, status := range statuses {
		line := fmt.Sprintf("%05d %s %s", status.Source.Version, status.State, status.Source.Path)
		if _, err := fmt.Fprintln(out, line); err != nil {
			return fmt.Errorf("write status: %w", err)
		}
	}
	return nil
}
