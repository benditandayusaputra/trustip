package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const roundingHalf = 0.5

func (r *Needs) ListInterested(ctx context.Context, need domain.Need) ([]domain.Interested, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListConsignorsNearNeed(readCtx, sqlc.ListConsignorsNearNeedParams{
		StoreOrgID: need.StoreID, CategoryID: need.Draft.CategoryID, RadiusMeters: domain.MatchingRadiusMeters,
	})
	if err != nil {
		return nil, fmt.Errorf("list consignors near need %s: %w", need.ID, err)
	}
	interested := make([]domain.Interested, 0, len(rows))
	for _, row := range rows {
		interested = append(interested, domain.Interested{
			OrganizationID: row.ID, DistanceMeters: int(row.DistanceMeters + roundingHalf),
		})
	}
	return interested, nil
}

func (r *Needs) ClaimExpiring(ctx context.Context, window domain.ReminderWindow) ([]domain.Expiring, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ClaimExpiringNeeds(writeCtx, sqlc.ClaimExpiringNeedsParams{
		Now: window.From, RemindUntil: window.Until,
	})
	if err != nil {
		return nil, fmt.Errorf("claim expiring needs: %w", err)
	}
	expiring := make([]domain.Expiring, 0, len(rows))
	for _, row := range rows {
		expiring = append(expiring, domain.Expiring{NeedID: row.ID, StoreID: row.StoreOrgID, Title: row.Title})
	}
	return expiring, nil
}
