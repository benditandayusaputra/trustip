package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Users) FindCredentials(ctx context.Context, email string) (domain.Credentials, bool, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).FindUserByEmail(readCtx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Credentials{}, false, nil
	}
	if err != nil {
		return domain.Credentials{}, false, fmt.Errorf("find user by email: %w", err)
	}
	return domain.Credentials{
		UserID:   row.ID,
		Email:    row.Email,
		FullName: row.FullName,
		Password: domain.StoredPassword{Hash: row.PasswordHash, PepperVersion: row.PepperVersion},
		Attempts: toLoginAttempts(row.FailedLoginCount, row.LockedUntil),
	}, true, nil
}

func (r *Users) LockLoginAttempts(ctx context.Context, userID string) (domain.LoginAttempts, error) {
	row, err := sqlc.New(r.pool.Querier(ctx)).LockLoginAttempts(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.LoginAttempts{}, apperror.NotFound()
	}
	if err != nil {
		return domain.LoginAttempts{}, fmt.Errorf("lock login attempts of %s: %w", userID, err)
	}
	return toLoginAttempts(row.FailedLoginCount, row.LockedUntil), nil
}

func (r *Users) SaveLoginAttempts(ctx context.Context, userID string, attempts domain.LoginAttempts) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	err := sqlc.New(r.pool.Querier(ctx)).SaveLoginAttempts(writeCtx, sqlc.SaveLoginAttemptsParams{
		ID:               userID,
		FailedLoginCount: failedCountColumn(attempts.FailedCount),
		LockedUntil:      optionalTime(attempts.LockedUntil),
	})
	if err != nil {
		return fmt.Errorf("save login attempts of %s: %w", userID, err)
	}
	return nil
}

func (r *Users) SavePassword(ctx context.Context, userID string, password domain.StoredPassword) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	err := sqlc.New(r.pool.Querier(ctx)).SavePassword(writeCtx, sqlc.SavePasswordParams{
		ID:            userID,
		PasswordHash:  password.Hash,
		PepperVersion: password.PepperVersion,
	})
	if err != nil {
		return fmt.Errorf("save password of %s: %w", userID, err)
	}
	return nil
}

func toLoginAttempts(failedCount int32, lockedUntil pgtype.Timestamptz) domain.LoginAttempts {
	attempts := domain.LoginAttempts{FailedCount: int(failedCount)}
	if lockedUntil.Valid {
		attempts.LockedUntil = lockedUntil.Time
	}
	return attempts
}

func failedCountColumn(count int) int32 {
	if count < 0 || count > domain.MaxFailedLogins {
		return domain.MaxFailedLogins
	}
	return int32(count)
}

func optionalTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: !value.IsZero()}
}
