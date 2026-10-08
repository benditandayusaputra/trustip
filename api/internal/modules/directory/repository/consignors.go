package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Directory) CountConsignorsWithin(ctx context.Context, query domain.ConsignorQuery) (map[int]int, error) {
	return r.countConsignors(ctx, query, false)
}

func (r *Directory) CountProductsWithin(ctx context.Context, query domain.ConsignorQuery) (map[int]int, error) {
	return r.countConsignors(ctx, query, true)
}

func (r *Directory) countConsignors(
	ctx context.Context, query domain.ConsignorQuery, seekingOnly bool,
) (map[int]int, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	list := consignorListParams(query)
	rows, err := sqlc.New(r.pool.Querier(ctx)).CountDirectoryConsignorsByRadius(readCtx,
		sqlc.CountDirectoryConsignorsByRadiusParams{
			RadiiMeters: radiusStepsInMeters(), OriginLongitude: list.OriginLongitude,
			OriginLatitude: list.OriginLatitude, ViewerLongitude: list.ViewerLongitude,
			ViewerLatitude: list.ViewerLatitude, HasPermit: list.HasPermit, HasHalal: list.HasHalal,
			CanSample: list.CanSample, DeliversToViewer: list.DeliversToViewer, SeekingOnly: seekingOnly,
			CategoryIds: list.CategoryIds, HasMinPrice: list.HasMinPrice, MinPrice: list.MinPrice,
			HasMaxPrice: list.HasMaxPrice, MaxPrice: list.MaxPrice, HasMinShelfLife: list.HasMinShelfLife,
			MinShelfLifeDays: list.MinShelfLifeDays, Query: list.Query,
		})
	if err != nil {
		return nil, fmt.Errorf("count directory consignors: %w", err)
	}
	counts := make(map[int]int, len(rows))
	for _, row := range rows {
		counts[int(row.RadiusMeters)/metersPerKm] = int(row.ConsignorCount)
	}
	return counts, nil
}

func (r *Directory) ListConsignors(ctx context.Context, query domain.ConsignorQuery) ([]domain.ConsignorCard, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListDirectoryConsignors(readCtx, consignorListParams(query))
	if err != nil {
		return nil, fmt.Errorf("list directory consignors: %w", err)
	}
	cards := make([]domain.ConsignorCard, 0, len(rows))
	for _, row := range rows {
		cards = append(cards, toConsignorCard(row))
	}
	return cards, nil
}

func (r *Directory) ListProducts(ctx context.Context, query domain.ConsignorQuery) ([]domain.ProductCard, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListDirectoryProducts(readCtx, productListParams(query))
	if err != nil {
		return nil, fmt.Errorf("list directory products: %w", err)
	}
	cards := make([]domain.ProductCard, 0, len(rows))
	for _, row := range rows {
		cards = append(cards, toProductCard(row))
	}
	return cards, nil
}

func consignorListParams(query domain.ConsignorQuery) sqlc.ListDirectoryConsignorsParams {
	filters := query.Filters
	afterID := query.After.ID
	if query.After.IsStart() {
		afterID = nilID
	}
	return sqlc.ListDirectoryConsignorsParams{
		FromStart: query.After.IsStart(), AfterText: query.After.Text, AfterNumber: query.After.Number,
		AfterID: afterID, RowLimit: int64(query.Limit), OriginLongitude: query.Origin.Point.Longitude,
		OriginLatitude: query.Origin.Point.Latitude, ViewerLongitude: query.ViewerLocation.Longitude,
		ViewerLatitude: query.ViewerLocation.Latitude, HasOrigin: query.Origin.IsKnown, ViewerID: query.ViewerID,
		WithinRadius: query.Origin.IsKnown && query.RadiusKm != domain.EveryDistance,
		RadiusMeters: float64(query.RadiusKm * metersPerKm), HasPermit: filters.HasPermit,
		HasHalal: filters.HasHalal, CanSample: filters.CanSample, DeliversToViewer: filters.DeliversToViewer,
		CategoryIds: nonNull(filters.CategoryIDs), HasMinPrice: filters.MinPrice.IsGiven,
		MinPrice: int64(filters.MinPrice.Value), HasMaxPrice: filters.MaxPrice.IsGiven,
		MaxPrice: int64(filters.MaxPrice.Value), HasMinShelfLife: filters.MinShelfLifeDays.IsGiven,
		MinShelfLifeDays: int64(filters.MinShelfLifeDays.Value), Query: filters.Text, SortOrder: string(query.Sort),
	}
}

func distanceOf(meters float64) domain.Distance {
	if meters < 0 {
		return domain.Distance{}
	}
	return domain.KnownDistance(int(meters + roundingHalf))
}

func productListParams(query domain.ConsignorQuery) sqlc.ListDirectoryProductsParams {
	list := consignorListParams(query)
	return sqlc.ListDirectoryProductsParams{
		FromStart: list.FromStart, AfterText: list.AfterText, AfterNumber: list.AfterNumber, AfterID: list.AfterID,
		RowLimit: list.RowLimit, OriginLongitude: list.OriginLongitude, OriginLatitude: list.OriginLatitude,
		ViewerLongitude: list.ViewerLongitude, ViewerLatitude: list.ViewerLatitude, HasOrigin: list.HasOrigin,
		ViewerID: list.ViewerID, WithinRadius: list.WithinRadius, RadiusMeters: list.RadiusMeters,
		CategoryIds: list.CategoryIds, HasMinPrice: list.HasMinPrice, MinPrice: list.MinPrice,
		HasMaxPrice: list.HasMaxPrice, MaxPrice: list.MaxPrice, HasMinShelfLife: list.HasMinShelfLife,
		MinShelfLifeDays: list.MinShelfLifeDays, HasPermit: list.HasPermit, HasHalal: list.HasHalal,
		CanSample: list.CanSample, DeliversToViewer: list.DeliversToViewer, Query: list.Query,
		SortOrder: list.SortOrder, ConsignorID: query.Filters.ConsignorID,
	}
}

func toConsignorCard(row sqlc.ListDirectoryConsignorsRow) domain.ConsignorCard {
	return domain.ConsignorCard{
		ID: row.ID, Slug: row.Slug, Name: row.Name, CityName: row.CityName, District: row.District,
		LogoStorageKey: row.LogoStorageKey, Distance: distanceOf(row.DistanceMeters),
		Categories: categoriesOf(row.CategoryIds, row.CategoryNames), ProductCount: int(row.ProductCount),
		Traits: domain.ConsignorTraits{
			Legality: domain.Legality{
				HasDistributionPermit: row.HasPermit, HasHalalCertificate: row.HasHalal,
				CanProvideSample: row.CanProvideSample,
			},
			DeliveryRadiusKm: int(row.DeliveryRadiusKm), PreferredStoreTypes: row.PreferredStoreTypes,
			Reputation:   domain.ConsignorReputation{ActivePartners: int(row.ActivePartnerCount)},
			LastActiveAt: row.LastActiveAt,
		},
		CreatedAt: row.CreatedAt, IsSelf: row.IsSelf, IsPartner: row.IsPartner,
		Position: domain.Position{Text: row.SortText, Number: row.SortNumber, ID: row.ID},
	}
}

func toProductCard(row sqlc.ListDirectoryProductsRow) domain.ProductCard {
	return domain.ProductCard{
		ID: row.ID, Name: row.Name, SKU: row.Sku, Unit: row.Unit,
		Category:             domain.Category{ID: row.CategoryID, Name: row.CategoryName},
		SuggestedRetailPrice: row.SuggestedRetailPrice, ShelfLifeDays: int(row.ShelfLifeDays),
		PhotoStorageKey: row.PhotoStorageKey,
		Consignor: domain.ProductConsignor{
			ID: row.ConsignorID, Slug: row.ConsignorSlug, Name: row.ConsignorName, CityName: row.CityName,
		},
		Distance: distanceOf(row.DistanceMeters),
		Traits: domain.ConsignorTraits{
			Legality: domain.Legality{
				HasDistributionPermit: row.HasPermit, HasHalalCertificate: row.HasHalal,
				CanProvideSample: row.CanProvideSample,
			},
			DeliveryRadiusKm: int(row.DeliveryRadiusKm), PreferredStoreTypes: row.PreferredStoreTypes,
			Reputation:   domain.ConsignorReputation{ActivePartners: int(row.ActivePartnerCount)},
			LastActiveAt: row.LastActiveAt,
		},
		CreatedAt: row.CreatedAt, IsSelf: row.IsSelf, IsPartner: row.IsPartner,
		Position: domain.Position{Text: row.SortText, Number: row.SortNumber, ID: row.ID},
	}
}
