package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const (
	metersPerKm          = 1000
	nilID                = "00000000-0000-0000-0000-000000000000"
	ratedFromSettlements = 3
	roundingHalf         = 0.5
)

type Directory struct {
	pool *database.Pool
}

func NewDirectory(pool *database.Pool) *Directory {
	return &Directory{pool: pool}
}

func (r *Directory) GetViewer(ctx context.Context, key domain.ViewerKey) (domain.Viewer, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetDirectoryViewer(readCtx, sqlc.GetDirectoryViewerParams{
		OrganizationID: key.OrganizationID, CityCode: key.CityCode,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Viewer{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Viewer{}, fmt.Errorf("get directory viewer %s: %w", key.OrganizationID, err)
	}
	return domain.Viewer{
		OrganizationID: key.OrganizationID, Type: actor.OrganizationType(row.Type),
		Location: domain.Point{Latitude: row.Latitude, Longitude: row.Longitude}, HasLocation: row.HasLocation,
		City: domain.Point{Latitude: row.CityLatitude, Longitude: row.CityLongitude}, HasCity: row.HasCity,
		Consignor: domain.ConsignorFacts{
			DeliveryRadiusKm: int(row.DeliveryRadiusKm), PreferredStoreTypes: row.PreferredStoreTypes,
			Categories: categoriesOf(row.CategoryIds, row.CategoryNames),
		},
		Store: domain.StoreFacts{
			StoreType: row.StoreType, AcceptedCategoryIDs: row.AcceptedCategoryIds, NeedCategoryIDs: []string{},
		},
	}, nil
}

func (r *Directory) CountStoresWithin(ctx context.Context, query domain.StoreQuery) (map[int]int, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	filters := query.Filters
	rows, err := sqlc.New(r.pool.Querier(ctx)).CountDirectoryStoresByRadius(readCtx,
		sqlc.CountDirectoryStoresByRadiusParams{
			RadiiMeters: radiusStepsInMeters(), OriginLongitude: query.Origin.Point.Longitude,
			OriginLatitude: query.Origin.Point.Latitude, StoreType: filters.StoreType,
			CategoryIds:  nonNull(filters.CategoryIDs),
			PricingModel: filters.PricingModel, HasMaxCommission: filters.HasMaxCommission,
			MaxCommissionBps: int64(filters.MaxCommissionBPS), WithoutPermit: filters.WithoutPermit,
			WithoutHalal: filters.WithoutHalal, Query: filters.Text,
		})
	if err != nil {
		return nil, fmt.Errorf("count directory stores: %w", err)
	}
	counts := make(map[int]int, len(rows))
	for _, row := range rows {
		counts[int(row.RadiusMeters)/metersPerKm] = int(row.StoreCount)
	}
	return counts, nil
}

func radiusStepsInMeters() []float64 {
	radii := make([]float64, 0, len(domain.RadiusSteps()))
	for _, step := range domain.RadiusSteps() {
		radii = append(radii, float64(step*metersPerKm))
	}
	return radii
}

func (r *Directory) ListStores(ctx context.Context, query domain.StoreQuery) ([]domain.StoreCard, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListDirectoryStores(readCtx, storeListParams(query))
	if err != nil {
		return nil, fmt.Errorf("list directory stores: %w", err)
	}
	cards := make([]domain.StoreCard, 0, len(rows))
	for _, row := range rows {
		cards = append(cards, toStoreCard(row))
	}
	return cards, nil
}

func storeListParams(query domain.StoreQuery) sqlc.ListDirectoryStoresParams {
	filters := query.Filters
	afterID := query.After.ID
	if query.After.IsStart() {
		afterID = nilID
	}
	return sqlc.ListDirectoryStoresParams{
		FromStart: query.After.IsStart(), AfterText: query.After.Text, AfterNumber: query.After.Number,
		AfterID: afterID, RowLimit: int64(query.Limit), OriginLongitude: query.Origin.Point.Longitude,
		OriginLatitude: query.Origin.Point.Latitude, HasOrigin: query.Origin.IsKnown, ViewerID: query.ViewerID,
		WithinRadius: query.Origin.IsKnown && query.RadiusKm != domain.EveryDistance,
		RadiusMeters: float64(query.RadiusKm * metersPerKm), StoreType: filters.StoreType,
		CategoryIds: nonNull(filters.CategoryIDs), PricingModel: filters.PricingModel,
		HasMaxCommission: filters.HasMaxCommission, MaxCommissionBps: int64(filters.MaxCommissionBPS),
		WithoutPermit: filters.WithoutPermit, WithoutHalal: filters.WithoutHalal, Query: filters.Text,
		SortOrder: string(query.Sort), RatedFrom: ratedFromSettlements,
	}
}

func toStoreCard(row sqlc.ListDirectoryStoresRow) domain.StoreCard {
	card := domain.StoreCard{
		ID: row.ID, Slug: row.Slug, Name: row.Name, CityName: row.CityName, District: row.District,
		StoreType: row.StoreType, LogoStorageKey: row.LogoStorageKey,
		AcceptedCategories: categoriesOf(row.AcceptedCategoryIds, row.AcceptedCategoryNames),
		Terms: domain.StoreTerms{
			AcceptsFixedDeposit: row.AcceptsFixedDeposit, AcceptsCommission: row.AcceptsCommission,
			Commission: domain.CommissionRange{
				MinBPS: int(row.CommissionMinBps), MaxBPS: int(row.CommissionMaxBps),
				IsKnown: row.CommissionMinBps >= 0,
			},
			RequiresDistributionPermit: row.RequiresDistributionPermit, RequiresHalal: row.RequiresHalal,
			RequiresSample: row.RequiresSample,
		},
		Reputation: domain.StoreReputation{
			DueSettlements: int(row.SettlementsCounted), OnTimeBPS: int(row.OnTimePaymentBps),
		},
		LastActiveAt: row.LastActiveAt, CreatedAt: row.CreatedAt, IsSelf: row.IsSelf, IsPartner: row.IsPartner,
		Position: domain.Position{Text: row.SortText, Number: row.SortNumber, ID: row.ID},
	}
	card.Distance = distanceOf(row.DistanceMeters)
	return card
}

func categoriesOf(ids, names []string) []domain.Category {
	categories := make([]domain.Category, 0, len(ids))
	for index, id := range ids {
		if index < len(names) {
			categories = append(categories, domain.Category{ID: id, Name: names[index]})
		}
	}
	return categories
}

func nonNull(ids []string) []string {
	if ids == nil {
		return []string{}
	}
	return ids
}
