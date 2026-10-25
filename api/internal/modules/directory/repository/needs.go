package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Directory) CountNeedsWithin(ctx context.Context, query domain.NeedQuery) (map[int]int, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).CountDirectoryNeedsByRadius(readCtx,
		sqlc.CountDirectoryNeedsByRadiusParams{
			RadiiMeters: radiusStepsInMeters(), OriginLongitude: query.Origin.Point.Longitude,
			OriginLatitude: query.Origin.Point.Latitude, Now: query.Now, CategoryIds: nonNull(query.CategoryIDs),
			Query: query.Text, ViewerID: query.ViewerID,
		})
	if err != nil {
		return nil, fmt.Errorf("count board needs: %w", err)
	}
	counts := make(map[int]int, len(rows))
	for _, row := range rows {
		counts[int(row.RadiusMeters)/metersPerKm] = int(row.NeedCount)
	}
	return counts, nil
}

func (r *Directory) ListNeeds(ctx context.Context, query domain.NeedQuery) ([]domain.NeedCard, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	afterID := query.After.ID
	if query.After.IsStart() {
		afterID = nilID
	}
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListDirectoryNeeds(readCtx, sqlc.ListDirectoryNeedsParams{
		FromStart: query.After.IsStart(), AfterText: query.After.Text, AfterNumber: query.After.Number,
		AfterID: afterID, RowLimit: int64(query.Limit), OriginLongitude: query.Origin.Point.Longitude,
		OriginLatitude: query.Origin.Point.Latitude, HasOrigin: query.Origin.IsKnown, ViewerID: query.ViewerID,
		Now: query.Now, WithinRadius: query.Origin.IsKnown && query.RadiusKm != domain.EveryDistance,
		RadiusMeters: float64(query.RadiusKm * metersPerKm), CategoryIds: nonNull(query.CategoryIDs),
		Query: query.Text, SortOrder: string(query.Sort),
	})
	if err != nil {
		return nil, fmt.Errorf("list board needs: %w", err)
	}
	cards := make([]domain.NeedCard, 0, len(rows))
	for _, row := range rows {
		cards = append(cards, toNeedCard(row))
	}
	return cards, nil
}

func toNeedCard(row sqlc.ListDirectoryNeedsRow) domain.NeedCard {
	return domain.NeedCard{
		ID: row.ID, Title: row.Title, Description: row.Description,
		Category:       domain.Category{ID: row.CategoryID, Name: row.CategoryName},
		WeeklyQuantity: int(row.EstQtyPerWeek), MaxRetailPrice: row.MaxRetailPrice,
		PreferredModel: row.PreferredModel, OfferedCommissionBPS: int(row.OfferedCommissionBps),
		HasOfferedCommission: row.OfferedCommissionBps >= 0, ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt,
		Store: domain.NeedStore{
			ID: row.StoreID, Slug: row.StoreSlug, Name: row.StoreName, CityName: row.CityName,
			StoreType: row.StoreType, LogoStorageKey: row.StoreLogoStorageKey,
		},
		Distance: distanceOf(row.DistanceMeters), IsOwn: row.IsOwn,
		Position: domain.Position{Text: row.SortText, Number: row.SortNumber, ID: row.ID},
	}
}
