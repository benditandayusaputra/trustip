package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const nilID = "00000000-0000-0000-0000-000000000000"

func (r *Drops) ListDrops(ctx context.Context, query domain.PageQuery) ([]domain.Summary, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	afterID := query.After.ID
	if query.After.IsStart() {
		afterID = nilID
	}
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListDrops(readCtx, sqlc.ListDropsParams{
		ConfirmCutoff: query.ConfirmCutoff, OrganizationID: query.OrganizationID, PartnershipID: query.PartnershipID,
		FromAt: query.From, ToAt: query.To, Status: string(query.Status), FromStart: query.After.IsStart(),
		AfterAt: query.After.At, AfterID: afterID, RowLimit: int64(query.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list drops of %s: %w", query.OrganizationID, err)
	}
	summaries := make([]domain.Summary, 0, len(rows))
	for _, row := range rows {
		summaries = append(summaries, domain.Summary{
			ID: row.ID, Number: row.Number, PartnershipID: row.PartnershipID, Status: domain.Status(row.Status),
			OccurredAt: row.OccurredAt, RecordedOffline: row.RecordedOffline, CounterpartID: row.CounterpartID,
			CounterpartName: row.CounterpartName, CounterpartType: row.CounterpartType,
			TotalQty: int(row.TotalQty), LineCount: int(row.LineCount),
		})
	}
	return summaries, nil
}
