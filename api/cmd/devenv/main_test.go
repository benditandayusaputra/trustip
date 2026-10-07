package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_FreshDirectory_CreatesBothFilesWithSharedEdgeKey(t *testing.T) {
	root := newProjectDir(t)

	if err := run(&bytes.Buffer{}, root); err != nil {
		t.Fatalf("run: %v", err)
	}

	apiEnv := readEnv(t, filepath.Join(root, "api", ".env"))
	appEnv := readEnv(t, filepath.Join(root, "app", ".env"))
	if apiEnv["EDGE_HMAC_KEY_V1"] == "" || apiEnv["EDGE_HMAC_KEY_V1"] != appEnv["EDGE_HMAC_KEY_V1"] {
		t.Fatal("edge key must be present and identical in api and app")
	}
}

func TestNewSecrets_EachSecretIsDistinct(t *testing.T) {
	secrets, err := newSecrets(rand.Reader)
	if err != nil {
		t.Fatalf("newSecrets: %v", err)
	}

	values := []string{secrets.jwt, secrets.pepper, secrets.edge, secrets.seal}
	seen := map[string]bool{}
	for _, value := range values {
		if value == "" || seen[value] {
			t.Fatal("secrets must be non-empty and distinct")
		}
		seen[value] = true
	}
}

func TestNewSecrets_ShortRandomSource_ReturnsError(t *testing.T) {
	if _, err := newSecrets(bytes.NewReader(make([]byte, secretBytes))); err == nil {
		t.Fatal("expected error when random source runs out")
	}
}

func TestRun_ExistingFile_IsNotOverwritten(t *testing.T) {
	root := newProjectDir(t)
	existing := filepath.Join(root, "api", ".env")
	if err := os.WriteFile(existing, []byte("APP_ENV=custom\n"), envPermissions); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := run(&bytes.Buffer{}, root); err != nil {
		t.Fatalf("run: %v", err)
	}

	if readEnv(t, existing)["APP_ENV"] != "custom" {
		t.Fatal("existing api/.env was overwritten")
	}
}

func TestRun_ApiEnv_ListsEveryExampleVariable(t *testing.T) {
	root := newProjectDir(t)

	if err := run(&bytes.Buffer{}, root); err != nil {
		t.Fatalf("run: %v", err)
	}

	for _, side := range []string{"api", "app"} {
		example := readEnv(t, filepath.Join("..", "..", "..", side, ".env.example"))
		generated := readEnv(t, filepath.Join(root, side, ".env"))
		for name := range example {
			if _, ok := generated[name]; !ok {
				t.Errorf("%s/.env is missing %s", side, name)
			}
		}
	}
}

func newProjectDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, side := range []string{"api", "app"} {
		if err := os.Mkdir(filepath.Join(root, side), 0o750); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	return root
}

func readEnv(t *testing.T, path string) map[string]string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	values := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		name, value, _ := strings.Cut(scanner.Text(), "=")
		values[name] = value
	}
	return values
}
