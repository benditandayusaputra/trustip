package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/trust/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

type Reports struct {
	pool *database.Pool
}

func NewReports(pool *database.Pool) *Reports {
	return &Reports{pool: pool}
}

func (r *Reports) SaveReport(ctx context.Context, filing domain.Filing) (domain.Filed, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).SaveListingReport(writeCtx, sqlc.SaveListingReportParams{
		ID: filing.ID, ReporterUserID: filing.ReporterUserID, ReporterOrgID: filing.ReporterOrgID,
		TargetType: string(filing.Draft.Target.Type), TargetID: filing.Draft.Target.ID,
		Reason: string(filing.Draft.Reason), Note: filing.Draft.Note, ReportedAt: filing.ReportedAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Filed{}, domain.ErrTargetNotFound
	}
	if err != nil {
		return domain.Filed{}, fmt.Errorf("save report %s: %w", filing.ID, err)
	}
	return domain.Filed{ID: row.ID, Status: row.Status}, nil
}
