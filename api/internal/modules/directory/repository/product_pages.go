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

func (r *Directory) GetProductPage(ctx context.Context, query domain.ProductPageQuery) (domain.ProductPage, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.GetPublicProduct(readCtx, sqlc.GetPublicProductParams{
		ConsignorSlug: query.ConsignorSlug, Sku: query.SKU,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ProductPage{}, apperror.NotFound()
	}
	if err != nil {
		return domain.ProductPage{}, fmt.Errorf("get public product %s/%s: %w", query.ConsignorSlug, query.SKU, err)
	}
	stores, err := queries.ListPublicProductStores(readCtx, sqlc.ListPublicProductStoresParams{
		OriginLongitude: query.Origin.Point.Longitude, OriginLatitude: query.Origin.Point.Latitude,
		HasOrigin: query.Origin.IsKnown, ConsignorOrgID: row.ConsignorID, ProductID: row.ID,
		RowLimit: domain.PublicStoreLimit,
	})
	if err != nil {
		return domain.ProductPage{}, fmt.Errorf("list stores selling %s: %w", row.ID, err)
	}
	page := toProductPage(row)
	page.Stores = make([]domain.AvailableStore, 0, len(stores))
	for _, store := range stores {
		page.Stores = append(page.Stores, toAvailableStore(store))
	}
	return page, nil
}

func toProductPage(row sqlc.GetPublicProductRow) domain.ProductPage {
	return domain.ProductPage{
		Product: domain.ProductDetail{
			ID: row.ID, SKU: row.Sku, Name: row.Name, Unit: row.Unit, Description: row.Description,
			Composition: row.Composition, NetContent: row.NetContent, CategoryName: row.CategoryName,
			SuggestedRetailPrice: row.SuggestedRetailPrice, PhotoStorageKey: row.PhotoStorageKey,
		},
		Maker: domain.ProductMaker{
			ProfileSlug: row.ConsignorProfileSlug, Name: row.ConsignorName, District: row.ConsignorDistrict,
			CityName: row.ConsignorCityName,
			Legality: domain.DeclaredLegality{
				BusinessIDNo: row.BusinessIDNo, DistributionPermitNo: row.DistributionPermitNo,
				HalalCertNo: row.HalalCertNo,
			},
		},
	}
}

func toAvailableStore(row sqlc.ListPublicProductStoresRow) domain.AvailableStore {
	return domain.AvailableStore{
		ProfileSlug: row.ProfileSlug, Name: row.Name, Address: row.Address, District: row.District,
		CityName: row.CityName,
		Location: domain.Place{
			Point: domain.Point{Latitude: row.Latitude, Longitude: row.Longitude}, IsKnown: row.HasLocation,
		},
		Distance: distanceOf(row.DistanceMeters),
	}
}
