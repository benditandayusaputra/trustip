package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

type Area string

const (
	Public  Area = "public"
	Private Area = "private"

	keyBytes      = 16
	directoryMode = 0o750
	fileMode      = 0o640
)

var (
	ErrNotFound = errors.New("stored file not found")
	keyFormat   = regexp.MustCompile(`^[0-9a-f]{32}\.(jpg|png)$`)
)

type Disk struct {
	root string
}

func NewDisk(root string) *Disk {
	return &Disk{root: root}
}

func (d *Disk) Put(ctx context.Context, area Area, data []byte, extension string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("store file: %w", err)
	}
	random := make([]byte, keyBytes)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate storage key: %w", err)
	}
	key := hex.EncodeToString(random) + "." + extension
	directory := filepath.Join(d.root, string(area))
	if err := os.MkdirAll(directory, directoryMode); err != nil {
		return "", fmt.Errorf("create %s area: %w", area, err)
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return "", fmt.Errorf("open %s area: %w", area, err)
	}
	if err := errors.Join(root.WriteFile(key, data, fileMode), root.Close()); err != nil {
		return "", fmt.Errorf("write %s file: %w", area, err)
	}
	return key, nil
}

func (d *Disk) Read(area Area, key string) ([]byte, error) {
	if !keyFormat.MatchString(key) {
		return nil, ErrNotFound
	}
	data, err := d.readInside(area, key)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read %s file: %w", area, err)
	}
	return data, nil
}

func (d *Disk) readInside(area Area, key string) ([]byte, error) {
	root, err := os.OpenRoot(filepath.Join(d.root, string(area)))
	if err != nil {
		return nil, fmt.Errorf("open %s area: %w", area, err)
	}
	data, err := root.ReadFile(key)
	return data, errors.Join(err, root.Close())
}

func (d *Disk) Delete(area Area, key string) error {
	if !keyFormat.MatchString(key) {
		return ErrNotFound
	}
	root, err := os.OpenRoot(filepath.Join(d.root, string(area)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open %s area: %w", area, err)
	}
	err = root.Remove(key)
	if errors.Is(err, fs.ErrNotExist) {
		err = nil
	}
	if err = errors.Join(err, root.Close()); err != nil {
		return fmt.Errorf("delete %s file: %w", area, err)
	}
	return nil
}
