package config_test

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/platform/config"
)

func validSecret(fill byte) string {
	key := make([]byte, 32)
	for i := range key {
		key[i] = fill + byte(i)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func validEnviron() map[string]string {
	return map[string]string{
		"APP_ENV":                     "development",
		"HTTP_ADDR":                   "127.0.0.1:8080",
		"PUBLIC_APP_ORIGIN":           "http://localhost:5173",
		"DATABASE_URL":                "postgres://trustip@localhost:54329/trustip?sslmode=disable",
		"JWT_SECRET_V1":               validSecret(1),
		"JWT_ACTIVE_KEY_ID":           "v1",
		"PASSWORD_PEPPER_V1":          validSecret(2),
		"PASSWORD_PEPPER_ACTIVE":      "v1",
		"EDGE_HMAC_KEY_V1":            validSecret(3),
		"EDGE_HMAC_ACTIVE":            "v1",
		"SEAL_PRIVATE_KEY_V1":         validSecret(4),
		"SEAL_ACTIVE_KEY_ID":          "v1",
		"TURNSTILE_SECRET_KEY":        "1x0000000000000000000000000000000AA",
		"TURNSTILE_EXPECTED_HOSTNAME": "localhost",
		"FILE_STORAGE_DIR":            "./storage",
		"PUBLIC_FILES_BASE_URL":       "http://localhost:8080/files",
		"LOG_LEVEL":                   "debug",
	}
}

func toEnviron(values map[string]string) []string {
	environ := make([]string, 0, len(values))
	for name, value := range values {
		environ = append(environ, name+"="+value)
	}
	return environ
}

func loadWith(t *testing.T, change func(map[string]string)) (config.Config, error) {
	t.Helper()
	values := validEnviron()
	change(values)
	cfg, err := config.Load(toEnviron(values))
	if err != nil {
		return cfg, fmt.Errorf("load test environment: %w", err)
	}
	return cfg, nil
}

func requireInvalid(t *testing.T, err error, mention string) {
	t.Helper()
	if !errors.Is(err, config.ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
	if !strings.Contains(err.Error(), mention) {
		t.Fatalf("err = %q, want mention of %s", err, mention)
	}
}

func TestLoad_EmptySecret_IsRejected(t *testing.T) {
	_, err := loadWith(t, func(values map[string]string) { values["JWT_SECRET_V1"] = "" })
	requireInvalid(t, err, "JWT_SECRET_V1")
}

func TestLoad_ShortSecret_IsRejected(t *testing.T) {
	short := base64.StdEncoding.EncodeToString([]byte("only-sixteen-byt"))
	_, err := loadWith(t, func(values map[string]string) { values["PASSWORD_PEPPER_V1"] = short })
	requireInvalid(t, err, "PASSWORD_PEPPER_V1")
}

func TestLoad_PlaceholderSecret_IsRejected(t *testing.T) {
	zeros := base64.StdEncoding.EncodeToString(make([]byte, 32))
	_, err := loadWith(t, func(values map[string]string) { values["EDGE_HMAC_KEY_V1"] = zeros })
	requireInvalid(t, err, "EDGE_HMAC_KEY_V1")
}

func TestLoad_NonBase64Secret_IsRejected(t *testing.T) {
	_, err := loadWith(t, func(values map[string]string) { values["JWT_SECRET_V1"] = "changeme" })
	requireInvalid(t, err, "JWT_SECRET_V1")
}

func TestLoad_SealSeedLongerThan32Bytes_IsRejected(t *testing.T) {
	long := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("ab", 20)))
	_, err := loadWith(t, func(values map[string]string) { values["SEAL_PRIVATE_KEY_V1"] = long })
	requireInvalid(t, err, "SEAL_PRIVATE_KEY_V1")
}

func TestLoad_ActiveKeyWithoutSecret_IsRejected(t *testing.T) {
	_, err := loadWith(t, func(values map[string]string) { values["JWT_ACTIVE_KEY_ID"] = "v2" })
	requireInvalid(t, err, "JWT_SECRET_V2")
}

func TestLoad_MalformedActiveKeyID_IsRejected(t *testing.T) {
	_, err := loadWith(t, func(values map[string]string) { values["EDGE_HMAC_ACTIVE"] = "latest" })
	requireInvalid(t, err, "EDGE_HMAC_ACTIVE")
}

func TestLoad_UnknownEnvironment_IsRejected(t *testing.T) {
	_, err := loadWith(t, func(values map[string]string) { values["APP_ENV"] = "staging" })
	requireInvalid(t, err, "APP_ENV")
}

func TestLoad_ManagedDatabaseOutsideProduction_IsRejected(t *testing.T) {
	_, err := loadWith(t, func(values map[string]string) {
		values["DATABASE_URL"] = "postgres://u@ep-x.ap-southeast-1.aws.neon.tech/db?sslmode=require"
	})
	requireInvalid(t, err, "DATABASE_URL")
}

func TestLoad_ProductionWithTurnstileTestKey_IsRejected(t *testing.T) {
	_, err := loadWith(t, func(values map[string]string) {
		values["APP_ENV"] = "production"
		values["PUBLIC_APP_ORIGIN"] = "https://trustip.id"
	})
	requireInvalid(t, err, "TURNSTILE_SECRET_KEY")
}

func TestLoad_ProductionWithPlainHTTPOrigin_IsRejected(t *testing.T) {
	_, err := loadWith(t, func(values map[string]string) {
		values["APP_ENV"] = "production"
		values["TURNSTILE_SECRET_KEY"] = "0x4AAAAAAAreal-secret-from-dashboard"
	})
	requireInvalid(t, err, "PUBLIC_APP_ORIGIN")
}

func TestLoad_SeveralProblems_AreAllReported(t *testing.T) {
	_, err := loadWith(t, func(values map[string]string) {
		values["JWT_SECRET_V1"] = ""
		values["HTTP_ADDR"] = ""
	})
	requireInvalid(t, err, "JWT_SECRET_V1")
	requireInvalid(t, err, "HTTP_ADDR")
}

func TestLoad_ErrorMessage_NeverContainsSecretValue(t *testing.T) {
	leaked := base64.StdEncoding.EncodeToString([]byte("sixteen-byte-key"))
	_, err := loadWith(t, func(values map[string]string) { values["JWT_SECRET_V1"] = leaked })
	if err == nil || strings.Contains(err.Error(), leaked) {
		t.Fatalf("err = %v, must exist and hide the value", err)
	}
}

func TestLoad_ValidEnvironment_ExposesActiveKeys(t *testing.T) {
	cfg, err := loadWith(t, func(map[string]string) {})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	id, key := cfg.JWTKeys.Active()
	if id != "v1" || base64.StdEncoding.EncodeToString(key) != validSecret(1) {
		t.Fatalf("active jwt key = %s", id)
	}
	if cfg.Environment != config.Development || cfg.LogLevel != slog.LevelDebug {
		t.Fatalf("environment = %s, log level = %s", cfg.Environment, cfg.LogLevel)
	}
}

func TestLoad_RotatedKeys_KeepOldKeyForVerification(t *testing.T) {
	cfg, err := loadWith(t, func(values map[string]string) {
		values["PASSWORD_PEPPER_V2"] = validSecret(9)
		values["PASSWORD_PEPPER_ACTIVE"] = "v2"
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	activeID, _ := cfg.PasswordPeppers.Active()
	_, hasOld := cfg.PasswordPeppers.Lookup("v1")
	if activeID != "v2" || !hasOld {
		t.Fatalf("active = %s, has v1 = %v", activeID, hasOld)
	}
}

func TestLoad_UnknownLogLevel_IsRejected(t *testing.T) {
	_, err := loadWith(t, func(values map[string]string) { values["LOG_LEVEL"] = "verbose" })
	requireInvalid(t, err, "LOG_LEVEL")
}

func TestLoad_HighBytePlaceholder_IsRejected(t *testing.T) {
	filled := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("\xff", 32)))
	_, err := loadWith(t, func(values map[string]string) { values["JWT_SECRET_V1"] = filled })
	requireInvalid(t, err, "JWT_SECRET_V1")
}

func TestLoad_ProductionWithDebugLogs_IsRejected(t *testing.T) {
	_, err := loadWith(t, func(values map[string]string) {
		values["APP_ENV"] = "production"
		values["LOG_LEVEL"] = "debug"
	})
	requireInvalid(t, err, "LOG_LEVEL")
}

func TestLoad_ProductionWithoutTLSDatabase_IsRejected(t *testing.T) {
	_, err := loadWith(t, func(values map[string]string) { values["APP_ENV"] = "production" })
	requireInvalid(t, err, "DATABASE_URL")
}

func TestLoad_ValidProduction_IsAccepted(t *testing.T) {
	_, err := loadWith(t, func(values map[string]string) {
		values["APP_ENV"] = "production"
		values["PUBLIC_APP_ORIGIN"] = "https://trustip.id"
		values["DATABASE_URL"] = "postgres://app@ep-x.ap-southeast-1.aws.neon.tech/db?sslmode=require"
		values["TURNSTILE_SECRET_KEY"] = "0x4AAAAAAAreal-secret-from-dashboard"
		values["LOG_LEVEL"] = "info"
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
}
