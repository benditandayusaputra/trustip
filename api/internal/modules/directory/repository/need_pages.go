package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Directory) GetPublicNeed(ctx context.Context, needID string) (domain.PublicNeed, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetPublicNeed(readCtx, needID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PublicNeed{}, apperror.NotFound()
	}
	if err != nil {
		return domain.PublicNeed{}, fmt.Errorf("get public need %s: %w", needID, err)
	}
	return domain.PublicNeed{
		ID: row.ID, Title: row.Title, CategoryName: row.CategoryName, Description: row.Description,
		EstQtyPerWeek: int(row.EstQtyPerWeek), MaxRetailPrice: row.MaxRetailPrice, PreferredModel: row.PreferredModel,
		OfferedCommission: domain.CommissionOffer{
			BPS: int(row.OfferedCommissionBps), IsGiven: row.HasOfferedCommission,
		},
		Status: row.Status, ExpiresAt: row.ExpiresAt,
		Store: domain.NeedPageStore{
			ID: row.StoreID, ProfileSlug: row.StoreProfileSlug, Name: row.StoreName, StoreType: row.StoreType,
			District: row.StoreDistrict, CityName: row.StoreCityName,
		},
	}, nil
}

func (r *Directory) getVisitor(ctx context.Context, cityCode string) (domain.Viewer, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetDirectoryVisitorCity(readCtx, cityCode)
	if err != nil {
		return domain.Viewer{}, fmt.Errorf("get city %s for a visitor: %w", cityCode, err)
	}
	return domain.Viewer{
		OrganizationID: domain.VisitorID,
		City:           domain.Point{Latitude: row.CityLatitude, Longitude: row.CityLongitude}, HasCity: row.HasCity,
	}, nil
}
