package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Drops) LockDrop(ctx context.Context, key domain.DropKey) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	_, err := sqlc.New(r.pool.Querier(ctx)).LockDrop(writeCtx, sqlc.LockDropParams{
		ID: key.DropID, OrganizationID: key.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.OutsideScope()
	}
	if err != nil {
		return fmt.Errorf("lock drop %s: %w", key.DropID, err)
	}
	return nil
}

func (r *Drops) IsStockUsed(ctx context.Context, dropID string) (bool, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	outflows, err := sqlc.New(r.pool.Querier(ctx)).CountDropOutflows(readCtx, dropID)
	if err != nil {
		return false, fmt.Errorf("count outflows of drop %s: %w", dropID, err)
	}
	return outflows > 0, nil
}

func (r *Drops) SaveDecision(ctx context.Context, decision domain.Decision, reversal []domain.Movement) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	changed, err := saveStatus(writeCtx, queries, decision)
	if err != nil {
		return fmt.Errorf("save %s of drop %s: %w", decision.Status, decision.DropID, err)
	}
	if changed == 0 {
		return domain.ErrNotPending
	}
	if len(reversal) == 0 {
		return nil
	}
	if err := queries.AddStockMovements(writeCtx, movementParams(reversal)); err != nil {
		return fmt.Errorf("reverse stock of drop %s: %w", decision.DropID, err)
	}
	return nil
}

func saveStatus(ctx context.Context, queries *sqlc.Queries, decision domain.Decision) (int64, error) {
	if decision.Status == domain.Cancelled {
		changed, err := queries.CancelDrop(ctx, sqlc.CancelDropParams{
			ID: decision.DropID, CancelledBy: decision.By, CancelledAt: decision.At,
		})
		return changed, wrapQuery(err)
	}
	changed, err := queries.ConfirmDrop(ctx, sqlc.ConfirmDropParams{
		ID: decision.DropID, ConfirmedBy: decision.By, ConfirmedAt: decision.At,
	})
	return changed, wrapQuery(err)
}

func wrapQuery(err error) error {
	if err != nil {
		return fmt.Errorf("update drop status: %w", err)
	}
	return nil
}
