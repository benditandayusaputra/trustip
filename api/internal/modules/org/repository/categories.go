package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Regions) ListCategories(ctx context.Context) ([]domain.Category, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListCategories(readCtx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	categories := make([]domain.Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, domain.Category{ID: row.ID, Name: row.Name, Slug: row.Slug})
	}
	return categories, nil
}
