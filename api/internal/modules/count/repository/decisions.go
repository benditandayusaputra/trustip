package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const nilID = "00000000-0000-0000-0000-000000000000"

func (r *Counts) LockCount(ctx context.Context, key domain.CountKey) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	_, err := sqlc.New(r.pool.Querier(ctx)).LockCount(writeCtx, sqlc.LockCountParams{
		ID: key.CountID, OrganizationID: key.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.OutsideScope()
	}
	if err != nil {
		return fmt.Errorf("lock count %s: %w", key.CountID, err)
	}
	return nil
}

func (r *Counts) SaveConfirmation(ctx context.Context, decision domain.Decision) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	changed, err := sqlc.New(r.pool.Querier(ctx)).ConfirmCount(writeCtx, sqlc.ConfirmCountParams{
		ID: decision.CountID, ConfirmedBy: decision.By, ConfirmedAt: decision.At,
	})
	if err != nil {
		return fmt.Errorf("confirm count %s: %w", decision.CountID, err)
	}
	if changed == 0 {
		return domain.ErrNotPending
	}
	return nil
}

func (r *Counts) ListCounts(ctx context.Context, query domain.ListQuery) ([]domain.Summary, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	afterID := query.After.ID
	if query.After.IsStart() {
		afterID = nilID
	}
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListCounts(readCtx, sqlc.ListCountsParams{
		ConfirmCutoff: query.ConfirmCutoff, OrganizationID: query.OrganizationID, PartnershipID: query.PartnershipID,
		FromAt: query.From, ToAt: query.To, FromStart: query.After.IsStart(), AfterAt: query.After.At,
		AfterID: afterID, RowLimit: int64(query.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list counts of %s: %w", query.OrganizationID, err)
	}
	summaries := make([]domain.Summary, 0, len(rows))
	for _, row := range rows {
		summaries = append(summaries, domain.Summary{
			ID: row.ID, Number: row.Number, PartnershipID: row.PartnershipID, Status: domain.Status(row.Status),
			CountedAt: row.CountedAt, RecordedOffline: row.RecordedOffline, CounterpartID: row.CounterpartID,
			CounterpartName: row.CounterpartName, CounterpartType: row.CounterpartType,
			CounterpartLogo: row.CounterpartLogoStorageKey, SoldQty: int(row.SoldQty),
			ReturnedQty: int(row.ReturnedQty), ConsignorShare: row.ConsignorShareAmount,
		})
	}
	return summaries, nil
}

func (r *Counts) ConfirmUnanswered(ctx context.Context, deadline domain.Deadline) (int, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	changed, err := sqlc.New(r.pool.Querier(ctx)).ConfirmUnansweredCounts(writeCtx, sqlc.ConfirmUnansweredCountsParams{
		WindowSeconds: deadline.Window.Seconds(), Now: deadline.At, Cutoff: deadline.Cutoff,
	})
	if err != nil {
		return 0, fmt.Errorf("confirm counts received before %s: %w", deadline.Cutoff, err)
	}
	return int(changed), nil
}

func (r *Counts) ApproveCountsOfSettlement(ctx context.Context, approval domain.SettlementApproval) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	_, err := sqlc.New(r.pool.Querier(ctx)).ApproveCountsOfSettlement(writeCtx, sqlc.ApproveCountsOfSettlementParams{
		ConfirmedBy: approval.By, ConfirmedAt: approval.At, SettlementID: approval.SettlementID,
	})
	if err != nil {
		return fmt.Errorf("approve counts of settlement %s: %w", approval.SettlementID, err)
	}
	return nil
}
