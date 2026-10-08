package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/trust/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

type Reputation struct {
	pool *database.Pool
}

func NewReputation(pool *database.Pool) *Reputation {
	return &Reputation{pool: pool}
}

func (r *Reputation) ListFacts(ctx context.Context, window domain.Window) ([]domain.Facts, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListReputationFacts(readCtx, sqlc.ListReputationFactsParams{
		Since: window.Since, Today: window.Today,
	})
	if err != nil {
		return nil, fmt.Errorf("list reputation facts since %s: %w", window.Since, err)
	}
	facts := make([]domain.Facts, 0, len(rows))
	for _, row := range rows {
		facts = append(facts, domain.Facts{
			OrganizationID: row.OrganizationID, ActivePartners: int(row.ActivePartners),
			DueSettlements: int(row.DueSettlements), OnTimeSettlements: int(row.OnTimeSettlements),
			Products: int(row.Products),
		})
	}
	return facts, nil
}

func (r *Reputation) SaveStats(ctx context.Context, stats []domain.Stats, computedAt time.Time) (int, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	params := sqlc.SaveOrgStatsParams{ComputedAt: computedAt}
	for _, stat := range stats {
		params.OrganizationIds = append(params.OrganizationIds, stat.OrganizationID)
		params.ActivePartnerCounts = append(params.ActivePartnerCounts, int64(stat.ActivePartners))
		params.OnTimePaymentBps = append(params.OnTimePaymentBps, int64(stat.OnTimeBPS))
		params.SettlementsCounted = append(params.SettlementsCounted, int64(stat.SettlementsCounted))
		params.ProductCounts = append(params.ProductCounts, int64(stat.Products))
	}
	saved, err := sqlc.New(r.pool.Querier(ctx)).SaveOrgStats(writeCtx, params)
	if err != nil {
		return 0, fmt.Errorf("save reputation of %d organizations: %w", len(stats), err)
	}
	return int(saved), nil
}
