package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type scanFunc func(source []byte) ([]int, error)

var errUnknownFileType = errors.New("unknown file type, classify it in cmd/nocomment")

func scannerFor(path string) (scanFunc, error) {
	name := filepath.Base(path)
	extension := strings.ToLower(filepath.Ext(name))
	if scan, ok := scannerByName(name); ok {
		return scan, nil
	}
	if scan, ok := scannerByExtension(extension); ok {
		return scan, nil
	}
	if isExemptExtension(extension) {
		return skipScan, nil
	}
	return nil, fmt.Errorf("%s: %w", path, errUnknownFileType)
}

func scannerByName(name string) (scanFunc, bool) {
	switch name {
	case "Makefile", "Dockerfile", "Caddyfile", ".gitignore", ".editorconfig", ".dockerignore",
		".gitattributes", ".npmrc", ".prettierignore", ".env.example", "robots.txt":
		return infallible(scanHashComments), true
	}
	return nil, false
}

func scannerByExtension(extension string) (scanFunc, bool) {
	switch extension {
	case ".go":
		return scanGoComments, true
	case ".sql":
		return infallible(scanSQLComments), true
	case ".css":
		return infallible(scanCSSComments), true
	case ".html", ".svg":
		return infallible(scanMarkupComments), true
	case ".svelte":
		return infallible(scanSvelteStyleComments), true
	case ".yaml", ".yml", ".toml", ".sh":
		return infallible(scanHashComments), true
	}
	return nil, false
}

func isExemptExtension(extension string) bool {
	switch extension {
	case ".md", ".json", ".webmanifest", ".mod", ".sum", ".txt",
		".ts", ".js", ".mjs", ".cjs",
		".png", ".ico", ".jpg", ".jpeg", ".webp", ".woff2":
		return true
	}
	return false
}

func skipScan([]byte) ([]int, error) {
	return []int{}, nil
}

func infallible(scan func(source []byte) []int) scanFunc {
	return func(source []byte) ([]int, error) {
		return scan(source), nil
	}
}
