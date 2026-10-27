package repository

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

type Sessions struct {
	pool *database.Pool
}

func NewSessions(pool *database.Pool) *Sessions {
	return &Sessions{pool: pool}
}

func (r *Sessions) CreateSession(ctx context.Context, session domain.Session) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	err := sqlc.New(r.pool.Querier(ctx)).CreateSession(writeCtx, sqlc.CreateSessionParams{
		ID:        session.ID,
		UserID:    session.UserID,
		FamilyID:  session.FamilyID,
		TokenHash: session.TokenHash,
		ExpiresAt: session.ExpiresAt,
		CreatedAt: session.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("insert session %s: %w", session.ID, err)
	}
	return nil
}

func (r *Sessions) LockSession(ctx context.Context, tokenHash []byte) (domain.StoredSession, bool, error) {
	row, err := sqlc.New(r.pool.Querier(ctx)).LockSessionByTokenHash(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.StoredSession{}, false, nil
	}
	if err != nil {
		return domain.StoredSession{}, false, fmt.Errorf("lock session by token hash: %w", err)
	}
	return domain.StoredSession{
		SessionIdentity: domain.SessionIdentity{ID: row.ID, UserID: row.UserID, FamilyID: row.FamilyID},
		ExpiresAt:       row.ExpiresAt,
		RotatedAt:       timeOrZero(row.RotatedAt),
		RevokedAt:       timeOrZero(row.RevokedAt),
	}, true, nil
}

func (r *Sessions) MarkRotated(ctx context.Context, sessionID string, at time.Time) error {
	err := sqlc.New(r.pool.Querier(ctx)).MarkSessionRotated(ctx, sqlc.MarkSessionRotatedParams{
		ID:        sessionID,
		RotatedAt: optionalTime(at),
	})
	if err != nil {
		return fmt.Errorf("mark session %s rotated: %w", sessionID, err)
	}
	return nil
}

func (r *Sessions) RevokeFamily(ctx context.Context, familyID string, at time.Time) error {
	err := sqlc.New(r.pool.Querier(ctx)).RevokeSessionFamily(ctx, sqlc.RevokeSessionFamilyParams{
		FamilyID:  familyID,
		RevokedAt: optionalTime(at),
	})
	if err != nil {
		return fmt.Errorf("revoke session family %s: %w", familyID, err)
	}
	return nil
}

func (r *Sessions) RevokeUser(ctx context.Context, userID string, at time.Time) ([]string, error) {
	families, err := sqlc.New(r.pool.Querier(ctx)).RevokeUserSessions(ctx, sqlc.RevokeUserSessionsParams{
		UserID:    userID,
		RevokedAt: optionalTime(at),
	})
	if err != nil {
		return nil, fmt.Errorf("revoke sessions of %s: %w", userID, err)
	}
	slices.Sort(families)
	return slices.Compact(families), nil
}

func (r *Sessions) RevokeOtherSessions(
	ctx context.Context, userID, keepFamily string, at time.Time,
) ([]string, error) {
	families, err := sqlc.New(r.pool.Querier(ctx)).RevokeOtherUserSessions(ctx, sqlc.RevokeOtherUserSessionsParams{
		UserID:       userID,
		KeepFamilyID: keepFamily,
		RevokedAt:    optionalTime(at),
	})
	if err != nil {
		return nil, fmt.Errorf("revoke other sessions of %s: %w", userID, err)
	}
	slices.Sort(families)
	return slices.Compact(families), nil
}

func timeOrZero(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}
