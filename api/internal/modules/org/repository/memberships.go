package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

type Memberships struct {
	pool *database.Pool
}

func NewMemberships(pool *database.Pool) *Memberships {
	return &Memberships{pool: pool}
}

func (r *Memberships) FindMembership(ctx context.Context, userID string) (actor.Membership, bool, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).FindActiveMembership(readCtx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return actor.Membership{}, false, nil
	}
	if err != nil {
		return actor.Membership{}, false, fmt.Errorf("find membership of %s: %w", userID, err)
	}
	return actor.Membership{
		OrganizationID:   row.OrganizationID,
		OrganizationType: actor.OrganizationType(row.OrganizationType),
		Role:             actor.Role(row.Role),
	}, true, nil
}
