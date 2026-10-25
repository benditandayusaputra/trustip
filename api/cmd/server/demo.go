package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/config"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const (
	demoCommand   = "demo"
	demoResetHour = 3
	demoAttempts  = 3
)

func demoSecretsOf(cfg config.Config) demoSecrets {
	return demoSecrets{peppers: cfg.PasswordPeppers, sealKeys: cfg.SealKeys, fileDir: cfg.FileStorageDir}
}

func runDemoCommand(cfg config.Config, base *slog.Logger) error {
	pool, err := openDatabase(cfg, base)
	if err != nil {
		return err
	}
	defer pool.Close()
	return resetDemo(pool, demoSecretsOf(cfg))(context.Background(), clock.System{}.Now())
}

func resetDemo(pool *database.Pool, secrets demoSecrets) func(ctx context.Context, now time.Time) error {
	return func(ctx context.Context, now time.Time) error {
		var failures []error
		for range demoAttempts {
			seeder, err := newDemoSeeder(pool, secrets, now)
			if err != nil {
				return err
			}
			if err = seeder.Run(ctx); err == nil {
				return nil
			}
			failures = append(failures, err)
		}
		return fmt.Errorf("seed demo data: %w", errors.Join(failures...))
	}
}
