package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	secretBytes    = 32
	envPermissions = 0o600
	localDatabase  = "postgres://trustip@localhost:54329/trustip?sslmode=disable"
	appOrigin      = "http://localhost:5173"
	apiOrigin      = "http://localhost:8080"
	turnstileTest  = "1x0000000000000000000000000000000AA"
	turnstileSite  = "1x00000000000000000000AA"
)

type variable struct {
	name  string
	value string
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[len(os.Args)-1]
	}
	if err := run(os.Stdout, root); err != nil {
		fmt.Fprintln(os.Stderr, "devenv:", err)
		os.Exit(1)
	}
}

func run(out io.Writer, root string) error {
	secrets, err := newSecrets(rand.Reader)
	if err != nil {
		return err
	}
	files := map[string][]variable{
		filepath.Join(root, "api", ".env"): apiVariables(secrets),
		filepath.Join(root, "app", ".env"): appVariables(secrets),
	}
	for path, variables := range files {
		if err := writeIfAbsent(out, path, variables); err != nil {
			return err
		}
	}
	return nil
}

type devSecrets struct {
	jwt    string
	pepper string
	edge   string
	seal   string
}

func newSecrets(source io.Reader) (devSecrets, error) {
	var secrets devSecrets
	for _, target := range []*string{&secrets.jwt, &secrets.pepper, &secrets.edge, &secrets.seal} {
		value, err := randomSecret(source)
		if err != nil {
			return devSecrets{}, err
		}
		*target = value
	}
	return secrets, nil
}

func randomSecret(source io.Reader) (string, error) {
	buffer := make([]byte, secretBytes)
	if _, err := io.ReadFull(source, buffer); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buffer), nil
}

func apiVariables(secrets devSecrets) []variable {
	return []variable{
		{"APP_ENV", "development"},
		{"HTTP_ADDR", "127.0.0.1:8080"},
		{"PUBLIC_APP_ORIGIN", appOrigin},
		{"DATABASE_URL", localDatabase},
		{"DATABASE_MIGRATE_URL", localDatabase},
		{"JWT_SECRET_V1", secrets.jwt},
		{"JWT_ACTIVE_KEY_ID", "v1"},
		{"PASSWORD_PEPPER_V1", secrets.pepper},
		{"PASSWORD_PEPPER_ACTIVE", "v1"},
		{"EDGE_HMAC_KEY_V1", secrets.edge},
		{"EDGE_HMAC_ACTIVE", "v1"},
		{"SEAL_PRIVATE_KEY_V1", secrets.seal},
		{"SEAL_ACTIVE_KEY_ID", "v1"},
		{"TURNSTILE_SECRET_KEY", turnstileTest},
		{"TURNSTILE_EXPECTED_HOSTNAME", "localhost"},
		{"FILE_STORAGE_DIR", "./storage"},
		{"PUBLIC_FILES_BASE_URL", apiOrigin + "/files"},
		{"LOG_LEVEL", "debug"},
	}
}

func appVariables(secrets devSecrets) []variable {
	return []variable{
		{"API_INTERNAL_URL", "http://127.0.0.1:8080"},
		{"EDGE_HMAC_KEY_V1", secrets.edge},
		{"EDGE_HMAC_ACTIVE", "v1"},
		{"PUBLIC_APP_ORIGIN", appOrigin},
		{"PUBLIC_STREAM_URL", apiOrigin + "/api/v1/stream"},
		{"PUBLIC_FILES_BASE_URL", apiOrigin + "/files"},
		{"PUBLIC_TURNSTILE_SITE_KEY", turnstileSite},
	}
}

func writeIfAbsent(out io.Writer, path string, variables []variable) error {
	file, err := os.OpenFile(filepath.Clean(path), os.O_WRONLY|os.O_CREATE|os.O_EXCL, envPermissions)
	if errors.Is(err, fs.ErrExist) {
		return report(out, path, "kept")
	}
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	if _, err := io.WriteString(file, render(variables)); err != nil {
		return errors.Join(fmt.Errorf("write %s: %w", path, err), file.Close())
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	return report(out, path, "created")
}

func render(variables []variable) string {
	var builder strings.Builder
	for _, entry := range variables {
		builder.WriteString(entry.name + "=" + entry.value + "\n")
	}
	return builder.String()
}

func report(out io.Writer, path, state string) error {
	if _, err := fmt.Fprintf(out, "%s %s\n", state, path); err != nil {
		return fmt.Errorf("report %s: %w", path, err)
	}
	return nil
}
