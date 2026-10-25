package storage_test

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/platform/storage"
)

var keyFormat = regexp.MustCompile(`^[0-9a-f]{32}\.jpg$`)

func TestPut_RSEC54_KeyIsRandomAndKeepsOnlyTheExtension(t *testing.T) {
	disk := storage.NewDisk(t.TempDir())

	first, err := disk.Put(t.Context(), storage.Public, []byte("one"), "jpg")
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	second, err := disk.Put(t.Context(), storage.Public, []byte("two"), "jpg")
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	if !keyFormat.MatchString(first) || first == second {
		t.Fatalf("keys = %q, %q", first, second)
	}
}

func TestRead_ReturnsWhatWasPutInTheSameArea(t *testing.T) {
	disk := storage.NewDisk(t.TempDir())
	key, err := disk.Put(t.Context(), storage.Private, []byte("qris"), "png")
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	data, err := disk.Read(storage.Private, key)

	if err != nil || string(data) != "qris" {
		t.Fatalf("data = %q, err = %v", data, err)
	}
}

func TestRead_RSEC72_PrivateFileIsNotInThePublicArea(t *testing.T) {
	disk := storage.NewDisk(t.TempDir())
	key, err := disk.Put(t.Context(), storage.Private, []byte("qris"), "png")
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	if _, err := disk.Read(storage.Public, key); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestRead_KeyThatIsNotAGeneratedKeyIsNotFound(t *testing.T) {
	root := t.TempDir()
	secret := filepath.Join(root, "secret.txt")
	if err := os.WriteFile(secret, []byte("rahasia"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	disk := storage.NewDisk(root)

	for _, key := range []string{"../secret.txt", "..%2Fsecret.txt", "secret.txt", "", "ABCDEF.jpg"} {
		if _, err := disk.Read(storage.Public, key); !errors.Is(err, storage.ErrNotFound) {
			t.Errorf("%q: err = %v", key, err)
		}
	}
}

func TestDelete_RemovesTheFileAndToleratesAMissingOne(t *testing.T) {
	disk := storage.NewDisk(t.TempDir())
	key, err := disk.Put(t.Context(), storage.Public, []byte("logo"), "jpg")
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	first := disk.Delete(storage.Public, key)
	second := disk.Delete(storage.Public, key)

	if _, err := disk.Read(storage.Public, key); first != nil || second != nil || !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("first = %v, second = %v, read = %v", first, second, err)
	}
}

func TestDelete_KeyThatIsNotAGeneratedKeyIsRefused(t *testing.T) {
	disk := storage.NewDisk(t.TempDir())

	if err := disk.Delete(storage.Public, "../secret.txt"); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}
