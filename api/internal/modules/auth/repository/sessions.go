package repository

import (
	"context"
	"fmt"

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
