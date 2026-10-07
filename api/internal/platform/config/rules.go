package config

import (
	"log/slog"
	"strings"
)

const managedDatabaseHost = "neon.tech"

func (r *reader) checkEnvironmentRules(cfg Config) {
	if cfg.Environment == Production {
		r.checkProductionRules(cfg)
		return
	}
	if strings.Contains(strings.ToLower(cfg.DatabaseURL), managedDatabaseHost) {
		r.reject("DATABASE_URL", "must point to the local database outside production")
	}
}

func (r *reader) checkProductionRules(cfg Config) {
	if !strings.HasPrefix(cfg.PublicAppOrigin, "https://") {
		r.reject("PUBLIC_APP_ORIGIN", "must use https in production")
	}
	if isTurnstileTestKey(cfg.Turnstile.SecretKey) {
		r.reject("TURNSTILE_SECRET_KEY", "must not be a Turnstile test key in production")
	}
	if !hasVerifiedTLS(cfg.DatabaseURL) {
		r.reject("DATABASE_URL", "must use sslmode=require or stricter in production")
	}
	if cfg.LogLevel < slog.LevelInfo {
		r.reject("LOG_LEVEL", "must not be debug in production")
	}
}

func isTurnstileTestKey(secret string) bool {
	for _, prefix := range []string{"1x0000", "2x0000", "3x0000"} {
		if strings.HasPrefix(secret, prefix) {
			return true
		}
	}
	return false
}

func hasVerifiedTLS(databaseURL string) bool {
	for _, mode := range []string{"sslmode=require", "sslmode=verify-ca", "sslmode=verify-full"} {
		if strings.Contains(databaseURL, mode) {
			return true
		}
	}
	return false
}
