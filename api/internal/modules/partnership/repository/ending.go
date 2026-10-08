package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Partnerships) LockForEnding(ctx context.Context, key domain.EndKey) (domain.EndState, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.LockPartnershipForEnding(writeCtx, sqlc.LockPartnershipForEndingParams{
		ID: key.PartnershipID, OrganizationID: key.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.EndState{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.EndState{}, fmt.Errorf("lock partnership %s: %w", key.PartnershipID, err)
	}
	rows, err := queries.ListPartnershipBlockers(writeCtx, sqlc.ListPartnershipBlockersParams{
		PartnershipID: row.ID, ConfirmCutoff: key.ConfirmCutoff,
	})
	if err != nil {
		return domain.EndState{}, fmt.Errorf("list blockers of %s: %w", row.ID, err)
	}
	state := domain.EndState{
		PartnershipID: row.ID, Status: domain.Status(row.Status), ConsignorID: row.ConsignorOrgID,
		StoreID: row.StoreOrgID, ConsignorName: row.ConsignorName, StoreName: row.StoreName,
		Blockers: make([]domain.Blocker, 0, len(rows)),
	}
	for _, blocker := range rows {
		state.Blockers = append(state.Blockers, domain.Blocker{
			Kind: domain.BlockerKind(blocker.Kind), DocumentID: blocker.DocumentID,
			DocumentNumber: blocker.DocumentNumber, Qty: int(blocker.Qty), Amount: blocker.Amount,
		})
	}
	return state, nil
}

func (r *Partnerships) SaveEnding(ctx context.Context, ending domain.Ending) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	changed, err := sqlc.New(r.pool.Querier(ctx)).EndPartnership(writeCtx, sqlc.EndPartnershipParams{
		EndedAt: ending.At, EndedByOrgID: ending.EndedByOrgID, EndReason: ending.Reason, ID: ending.PartnershipID,
	})
	if err != nil {
		return fmt.Errorf("end partnership %s: %w", ending.PartnershipID, err)
	}
	if changed == 0 {
		return apperror.New(apperror.CodePartnershipNotActive)
	}
	return nil
}

func (r *Partnerships) SetShownOnPublic(ctx context.Context, change domain.ListingChange) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	changed, err := sqlc.New(r.pool.Querier(ctx)).SetPartnershipShownOnPublic(writeCtx,
		sqlc.SetPartnershipShownOnPublicParams{
			ShowOnPublic: change.ShowOnPublic, UpdatedAt: change.At, ID: change.PartnershipID,
			StoreOrgID: change.StoreID,
		})
	if err != nil {
		return fmt.Errorf("set public listing of %s: %w", change.PartnershipID, err)
	}
	if changed == 0 {
		return apperror.OutsideScope()
	}
	return nil
}
