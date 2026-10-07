package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
)

const checkTimeout = 15 * time.Second

var (
	errUsage      = errors.New("usage: dbcheck <env-file>")
	errMissingURL = errors.New("DATABASE_MIGRATE_URL is empty")
	errConnect    = errors.New("cannot connect to the database, check network access and credentials")
	errMissingExt = errors.New("required extensions are not available")
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	err := runWithEnvFile(ctx, os.Stdout, os.Args[1:])
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "dbcheck:", err)
		os.Exit(1)
	}
}

func runWithEnvFile(ctx context.Context, out io.Writer, args []string) error {
	if len(args) != 1 {
		return errUsage
	}
	databaseURL, err := readEnvValue(args[0], "DATABASE_MIGRATE_URL")
	if err != nil {
		return err
	}
	return run(ctx, out, databaseURL)
}

func readEnvValue(path, name string) (string, error) {
	file, err := os.OpenInRoot(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		return "", fmt.Errorf("open env file: %w", err)
	}
	value := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, found, ok := strings.Cut(scanner.Text(), "=")
		if ok && strings.TrimSpace(key) == name {
			value = strings.Trim(strings.TrimSpace(found), `"'`)
		}
	}
	return value, errors.Join(wrapScanError(scanner.Err()), file.Close())
}

func wrapScanError(err error) error {
	if err != nil {
		return fmt.Errorf("read env file: %w", err)
	}
	return nil
}

func run(ctx context.Context, out io.Writer, databaseURL string) error {
	if databaseURL == "" {
		return errMissingURL
	}
	conn, err := connectReadOnly(ctx, databaseURL)
	if err != nil {
		return err
	}
	return errors.Join(inspect(ctx, out, sqlc.New(conn)), conn.Close(ctx))
}

func connectReadOnly(ctx context.Context, databaseURL string) (*pgx.Conn, error) {
	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, errConnect
	}
	config.RuntimeParams["default_transaction_read_only"] = "on"
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, describeConnectError(err)
	}
	return conn, nil
}

func describeConnectError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return fmt.Errorf("%w (server code %s)", errConnect, pgErr.Code)
	}
	return errConnect
}

func inspect(ctx context.Context, out io.Writer, queries *sqlc.Queries) error {
	version, err := queries.GetServerVersion(ctx)
	if err != nil {
		return fmt.Errorf("read server version: %w", err)
	}
	extensions, err := queries.ListAvailableExtensions(ctx, requiredExtensions())
	if err != nil {
		return fmt.Errorf("read available extensions: %w", err)
	}
	if _, err := fmt.Fprintf(out, "server_version %s\n", version); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return reportExtensions(out, extensions)
}

func requiredExtensions() []string {
	return []string{"pg_trgm", "postgis"}
}

func reportExtensions(out io.Writer, extensions []sqlc.ListAvailableExtensionsRow) error {
	for _, extension := range extensions {
		line := fmt.Sprintf("extension %s %s available", extension.Name.String, extension.DefaultVersion.String)
		if _, err := fmt.Fprintln(out, line); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
	}
	if len(extensions) != len(requiredExtensions()) {
		return errMissingExt
	}
	return nil
}
