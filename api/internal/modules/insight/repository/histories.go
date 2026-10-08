package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/insight/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

type Histories struct {
	pool *database.Pool
}

func NewHistories(pool *database.Pool) *Histories {
	return &Histories{pool: pool}
}

func (r *Histories) ListDropHistories(
	ctx context.Context, query domain.HistoryQuery,
) ([]domain.ProductHistory, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	if err := findPartnership(readCtx, queries, query); err != nil {
		return nil, err
	}
	products, err := queries.ListDropSuggestionProducts(readCtx, sqlc.ListDropSuggestionProductsParams{
		PartnershipID: query.PartnershipID, OrganizationID: query.OrganizationID,
		SalesSince: query.SalesSince, Now: query.Now,
	})
	if err != nil {
		return nil, fmt.Errorf("list drop suggestion products of %s: %w", query.PartnershipID, err)
	}
	sales, err := queries.ListSameWeekdaySales(readCtx, sqlc.ListSameWeekdaySalesParams{
		PartnershipID: query.PartnershipID, OrganizationID: query.OrganizationID,
		IsoWeekday: int64(query.ISOWeekday), MaxOccurrences: int64(query.MaxOccurrences),
	})
	if err != nil {
		return nil, fmt.Errorf("list same weekday sales of %s: %w", query.PartnershipID, err)
	}
	return historiesOf(products, salesByProduct(sales)), nil
}

func findPartnership(ctx context.Context, queries *sqlc.Queries, query domain.HistoryQuery) error {
	_, err := queries.FindInsightPartnership(ctx, sqlc.FindInsightPartnershipParams{
		PartnershipID: query.PartnershipID, OrganizationID: query.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.OutsideScope()
	}
	if err != nil {
		return fmt.Errorf("find partnership %s: %w", query.PartnershipID, err)
	}
	return nil
}

func salesByProduct(rows []sqlc.ListSameWeekdaySalesRow) map[string][]domain.DaySales {
	grouped := map[string][]domain.DaySales{}
	for _, row := range rows {
		grouped[row.ProductID] = append(grouped[row.ProductID], domain.DaySales{
			Sold: int(row.Sold), SoldOut: row.SoldOut,
		})
	}
	return grouped
}

func historiesOf(
	rows []sqlc.ListDropSuggestionProductsRow, sales map[string][]domain.DaySales,
) []domain.ProductHistory {
	histories := make([]domain.ProductHistory, 0, len(rows))
	for _, row := range rows {
		histories = append(histories, domain.ProductHistory{ProductID: row.ProductID, History: domain.DropHistory{
			Rhythm:           terms.Rhythm(row.CountRhythm),
			SameWeekdaySales: sales[row.ProductID],
			Sales:            domain.SalesHistory{SoldLast14Days: int(row.SoldSince), FirstDropAt: row.FirstDropAt},
			StockUnits:       int(row.StockUnits),
			ShelfLife:        shelfLifeOf(row.ShelfLifeDays),
		}})
	}
	return histories
}

func shelfLifeOf(days int64) stock.ShelfLife {
	if days <= 0 {
		return stock.UnknownShelfLife()
	}
	return stock.ShelfLifeOf(int(days))
}
