package config

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

const secretBytes = 32

type lengthRule int

const (
	minimumLength lengthRule = iota
	exactLength
)

type keyringSpec struct {
	secretPrefix string
	activeName   string
	length       lengthRule
}

var (
	keyID       = regexp.MustCompile(`^v[1-9][0-9]*$`)
	versionedID = regexp.MustCompile(`^V[1-9][0-9]*$`)
)

type reader struct {
	values   map[string]string
	problems []error
}

func newReader(environ []string) *reader {
	values := make(map[string]string, len(environ))
	for _, entry := range environ {
		name, value, _ := strings.Cut(entry, "=")
		values[name] = value
	}
	return &reader{values: values}
}

func (r *reader) reject(name, reason string) {
	r.problems = append(r.problems, fmt.Errorf("%w: %s %s", ErrInvalid, name, reason))
}

func (r *reader) err() error {
	return errors.Join(r.problems...)
}

func (r *reader) required(name string) string {
	value := strings.TrimSpace(r.values[name])
	if value == "" {
		r.reject(name, "is empty")
	}
	return value
}

func (r *reader) environment() Environment {
	environment := Environment(r.values["APP_ENV"])
	switch environment {
	case Development, Test, Production:
		return environment
	}
	r.reject("APP_ENV", "must be development, test, or production")
	return environment
}

func (r *reader) logLevel() slog.Level {
	var level slog.Level
	raw := r.values["LOG_LEVEL"]
	if raw == "" {
		return slog.LevelInfo
	}
	if err := level.UnmarshalText([]byte(raw)); err != nil {
		r.reject("LOG_LEVEL", "must be debug, info, warn, or error")
	}
	return level
}

func (r *reader) keyring(spec keyringSpec) Keyring {
	activeID := r.required(spec.activeName)
	if activeID != "" && !keyID.MatchString(activeID) {
		r.reject(spec.activeName, "must look like v1")
	}
	keys := r.versionedSecrets(spec)
	if _, ok := keys[activeID]; !ok && keyID.MatchString(activeID) {
		r.reject(spec.secretPrefix+strings.ToUpper(activeID), "is empty but named as the active key")
	}
	return Keyring{activeID: activeID, keys: keys}
}

func (r *reader) versionedSecrets(spec keyringSpec) map[string][]byte {
	keys := map[string][]byte{}
	for name, value := range r.values {
		version, found := strings.CutPrefix(name, spec.secretPrefix)
		if !found || !versionedID.MatchString(version) || value == "" {
			continue
		}
		if key, ok := r.decodeSecret(name, value, spec.length); ok {
			keys[strings.ToLower(version)] = key
		}
	}
	return keys
}

func (r *reader) decodeSecret(name, value string, rule lengthRule) ([]byte, bool) {
	key, err := base64.StdEncoding.DecodeString(value)
	switch {
	case err != nil:
		r.reject(name, "must be base64")
	case rule == exactLength && len(key) != secretBytes:
		r.reject(name, "must decode to exactly 32 bytes")
	case len(key) < secretBytes:
		r.reject(name, "must decode to at least 32 bytes")
	case isPlaceholder(key):
		r.reject(name, "still holds a placeholder value")
	default:
		return key, true
	}
	return nil, false
}

func isPlaceholder(key []byte) bool {
	return bytes.Count(key, key[:1]) == len(key)
}
