package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

type Regions struct {
	pool *database.Pool
}

func NewRegions(pool *database.Pool) *Regions {
	return &Regions{pool: pool}
}

func (r *Regions) ListProvinces(ctx context.Context) ([]domain.Province, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListProvinces(readCtx)
	if err != nil {
		return nil, fmt.Errorf("list provinces: %w", err)
	}
	provinces := make([]domain.Province, 0, len(rows))
	for _, row := range rows {
		provinces = append(provinces, domain.Province{Code: row.ProvinceCode, Name: row.ProvinceName})
	}
	return provinces, nil
}

func (r *Regions) ListCities(ctx context.Context, provinceCode string) ([]domain.City, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListCitiesInProvince(readCtx, provinceCode)
	if err != nil {
		return nil, fmt.Errorf("list cities in %s: %w", provinceCode, err)
	}
	cities := make([]domain.City, 0, len(rows))
	for _, row := range rows {
		cities = append(cities, domain.City{Code: row.Code, Name: row.Name, ProvinceName: row.ProvinceName})
	}
	return cities, nil
}
