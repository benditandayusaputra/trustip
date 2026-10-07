package config

import (
	"errors"
	"log/slog"
)

var ErrInvalid = errors.New("invalid configuration")

type Environment string

const (
	Development Environment = "development"
	Test        Environment = "test"
	Production  Environment = "production"
)

type Config struct {
	Environment        Environment
	HTTPAddr           string
	PublicAppOrigin    string
	DatabaseURL        string
	JWTKeys            Keyring
	PasswordPeppers    Keyring
	EdgeKeys           Keyring
	SealKeys           Keyring
	Turnstile          Turnstile
	FileStorageDir     string
	PublicFilesBaseURL string
	LogLevel           slog.Level
}

type Turnstile struct {
	SecretKey        string
	ExpectedHostname string
}

func Load(environ []string) (Config, error) {
	reader := newReader(environ)
	cfg := Config{
		Environment:     reader.environment(),
		HTTPAddr:        reader.required("HTTP_ADDR"),
		PublicAppOrigin: reader.required("PUBLIC_APP_ORIGIN"),
		DatabaseURL:     reader.required("DATABASE_URL"),
		JWTKeys:         reader.keyring(keyringSpec{"JWT_SECRET_", "JWT_ACTIVE_KEY_ID", minimumLength}),
		PasswordPeppers: reader.keyring(keyringSpec{"PASSWORD_PEPPER_", "PASSWORD_PEPPER_ACTIVE", minimumLength}),
		EdgeKeys:        reader.keyring(keyringSpec{"EDGE_HMAC_KEY_", "EDGE_HMAC_ACTIVE", minimumLength}),
		SealKeys:        reader.keyring(keyringSpec{"SEAL_PRIVATE_KEY_", "SEAL_ACTIVE_KEY_ID", exactLength}),
		Turnstile: Turnstile{
			SecretKey:        reader.required("TURNSTILE_SECRET_KEY"),
			ExpectedHostname: reader.required("TURNSTILE_EXPECTED_HOSTNAME"),
		},
		FileStorageDir:     reader.required("FILE_STORAGE_DIR"),
		PublicFilesBaseURL: reader.required("PUBLIC_FILES_BASE_URL"),
		LogLevel:           reader.logLevel(),
	}
	reader.checkEnvironmentRules(cfg)
	return cfg, reader.err()
}
