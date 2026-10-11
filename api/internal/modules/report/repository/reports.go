package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/modules/report/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

type Reports struct {
	pool *database.Pool
}

func NewReports(pool *database.Pool) *Reports {
	return &Reports{pool: pool}
}

type salesQuery func(ctx context.Context, queries *sqlc.Queries, params sqlc.SalesByPartnerParams) (
	[]sqlc.SalesByPartnerRow, error)

func salesQueries() map[domain.Grouping]salesQuery {
	return map[domain.Grouping]salesQuery{
		domain.ByPartner: func(ctx context.Context, queries *sqlc.Queries, params sqlc.SalesByPartnerParams) (
			[]sqlc.SalesByPartnerRow, error,
		) {
			return queries.SalesByPartner(ctx, params)
		},
		domain.ByProduct: func(ctx context.Context, queries *sqlc.Queries, params sqlc.SalesByPartnerParams) (
			[]sqlc.SalesByPartnerRow, error,
		) {
			rows, err := queries.SalesByProduct(ctx, sqlc.SalesByProductParams(params))
			return convertRows(rows, func(row sqlc.SalesByProductRow) sqlc.SalesByPartnerRow {
				return sqlc.SalesByPartnerRow(row)
			}), err
		},
		domain.ByDay: func(ctx context.Context, queries *sqlc.Queries, params sqlc.SalesByPartnerParams) (
			[]sqlc.SalesByPartnerRow, error,
		) {
			rows, err := queries.SalesByDay(ctx, sqlc.SalesByDayParams(params))
			return convertRows(rows, func(row sqlc.SalesByDayRow) sqlc.SalesByPartnerRow {
				return sqlc.SalesByPartnerRow(row)
			}), err
		},
	}
}

func convertRows[Row any](rows []Row, convert func(Row) sqlc.SalesByPartnerRow) []sqlc.SalesByPartnerRow {
	converted := make([]sqlc.SalesByPartnerRow, 0, len(rows))
	for _, row := range rows {
		converted = append(converted, convert(row))
	}
	return converted
}

func (r *Reports) SalesRows(ctx context.Context, query domain.Query) ([]domain.Row, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := salesQueries()[query.GroupBy](readCtx, sqlc.New(r.pool.Querier(ctx)), sqlc.SalesByPartnerParams{
		OrganizationID: query.OrganizationID, FromAt: query.From, ToAt: query.To,
		PartnerID: query.PartnerID, ProductID: query.ProductID,
	})
	if err != nil {
		return nil, fmt.Errorf("report sales of %s by %s: %w", query.OrganizationID, query.GroupBy, err)
	}
	found := make([]domain.Row, 0, len(rows))
	for _, row := range rows {
		found = append(found, domain.Row{
			Key: row.Key, Label: row.Label,
			Product: domain.ProductPreview{PhotoKey: row.PhotoKey, Slug: row.ConsignorSlug, SKU: row.Sku},
			Figures: domain.Figures{
				SoldQty: int(row.SoldQty), ReturnedQty: int(row.ReturnedQty), ReceivedQty: int(row.ReceivedQty),
				Split: pricing.Split{
					Gross: money.Rupiah(row.GrossAmount), StoreShare: money.Rupiah(row.StoreShareAmount),
					ConsignorShare: money.Rupiah(row.ConsignorShareAmount),
				},
			},
		})
	}
	return found, nil
}
