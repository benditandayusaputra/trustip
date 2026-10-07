package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Users) GetAccount(ctx context.Context, userID string) (domain.Account, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetAccount(readCtx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Account{}, apperror.NotFound()
	}
	if err != nil {
		return domain.Account{}, fmt.Errorf("get account %s: %w", userID, err)
	}
	return domain.Account{ID: row.ID, Email: row.Email, FullName: row.FullName}, nil
}
