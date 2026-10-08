package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/stock/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const nilID = "00000000-0000-0000-0000-000000000000"

func (r *Stocks) GetCardProduct(ctx context.Context, key domain.CardKey) (domain.CardProduct, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetStockCardProduct(readCtx, sqlc.GetStockCardProductParams{
		PartnershipID: key.PartnershipID, ProductID: key.ProductID, OrganizationID: key.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CardProduct{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.CardProduct{}, fmt.Errorf("get product %s: %w", key.ProductID, err)
	}
	return domain.CardProduct{ID: row.ID, Name: row.Name, SKU: row.Sku, Unit: row.Unit}, nil
}

func (r *Stocks) ListCard(ctx context.Context, query domain.CardQuery) ([]domain.CardEntry, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	afterID := query.After.ID
	if query.After.IsStart() {
		afterID = nilID
	}
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListStockCard(readCtx, sqlc.ListStockCardParams{
		PartnershipID: query.PartnershipID, ProductID: query.ProductID, OrganizationID: query.OrganizationID,
		FromStart: query.After.IsStart(), AfterAt: query.After.At, AfterID: afterID, RowLimit: int64(query.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list stock card of %s: %w", query.ProductID, err)
	}
	entries := make([]domain.CardEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, domain.CardEntry{
			ID: row.ID, Type: row.Type, Qty: int(row.Qty), OccurredAt: row.OccurredAt, BatchCode: row.BatchCode,
			SourceType: row.SourceType, SourceID: row.SourceID, SourceNumber: row.SourceNumber,
			BalanceAfter: int(row.BalanceAfter),
		})
	}
	return entries, nil
}
