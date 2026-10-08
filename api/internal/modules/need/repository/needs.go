package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const (
	openLimitConstraint = "needs_open_limit"
	nilID               = "00000000-0000-0000-0000-000000000000"
)

type Needs struct {
	pool *database.Pool
}

func NewNeeds(pool *database.Pool) *Needs {
	return &Needs{pool: pool}
}

func (r *Needs) CreateNeed(ctx context.Context, posting domain.Posting) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	draft := posting.Draft
	created, err := sqlc.New(r.pool.Querier(ctx)).CreateNeed(writeCtx, sqlc.CreateNeedParams{
		ID: posting.ID, StoreOrgID: posting.StoreID, Title: draft.Title, CategoryID: draft.CategoryID,
		Description: draft.Description, EstQtyPerWeek: int8Of(draft.WeeklyQuantity),
		MaxRetailPrice: int8Of(draft.MaxRetailPrice), PreferredModel: draft.PreferredModel,
		OfferedCommissionBps: int8Of(draft.OfferedCommissionBPS), ExpiresAt: posting.ExpiresAt,
		CreatedBy: posting.CreatedBy, CreatedAt: posting.At,
	})
	if database.IsCheckViolation(err, openLimitConstraint) {
		return domain.ErrOpenLimit
	}
	if err != nil {
		return fmt.Errorf("create need %s: %w", posting.ID, err)
	}
	if created == 0 {
		return domain.ErrUnknownCategory
	}
	return nil
}

func (r *Needs) GetNeed(ctx context.Context, key domain.NeedKey) (domain.Need, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetNeed(readCtx, sqlc.GetNeedParams{
		ID: key.NeedID, StoreOrgID: key.StoreID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Need{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Need{}, fmt.Errorf("get need %s: %w", key.NeedID, err)
	}
	return toNeed(row), nil
}

func (r *Needs) LockNeed(ctx context.Context, key domain.NeedKey) (domain.Need, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).LockNeed(writeCtx, sqlc.LockNeedParams{
		ID: key.NeedID, StoreOrgID: key.StoreID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Need{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Need{}, fmt.Errorf("lock need %s: %w", key.NeedID, err)
	}
	return toNeed(sqlc.GetNeedRow(row)), nil
}

func (r *Needs) ListNeeds(ctx context.Context, query domain.ListQuery) ([]domain.Need, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	afterID := query.After.ID
	if query.After.IsStart() {
		afterID = nilID
	}
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListStoreNeeds(readCtx, sqlc.ListStoreNeedsParams{
		StoreOrgID: query.StoreID, Status: string(query.Status), FromStart: query.After.IsStart(),
		AfterAt: query.After.At, AfterID: afterID, RowLimit: int64(query.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list needs of %s: %w", query.StoreID, err)
	}
	needs := make([]domain.Need, 0, len(rows))
	for _, row := range rows {
		needs = append(needs, toNeed(sqlc.GetNeedRow(row)))
	}
	return needs, nil
}

func (r *Needs) ListResponses(ctx context.Context, key domain.NeedKey) ([]domain.Response, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListNeedResponses(readCtx, sqlc.ListNeedResponsesParams{
		NeedID: key.NeedID, StoreOrgID: key.StoreID,
	})
	if err != nil {
		return nil, fmt.Errorf("list responses of %s: %w", key.NeedID, err)
	}
	responses := make([]domain.Response, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, domain.Response{
			ProposalID: row.ID, Number: row.Number, Status: row.Status, FromOrgID: row.FromOrgID,
			AwaitingOrgID: row.AwaitingOrgID, ConsignorName: row.ConsignorName, CreatedAt: row.CreatedAt,
		})
	}
	return responses, nil
}

func (r *Needs) SaveChange(ctx context.Context, change domain.Change) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	saved, err := sqlc.New(r.pool.Querier(ctx)).SaveNeedChange(writeCtx, sqlc.SaveNeedChangeParams{
		ID: change.NeedID, StoreOrgID: change.StoreID, Status: string(change.Status), ExpiresAt: change.ExpiresAt,
		ChangedAt: change.At, PreviousStatus: string(change.Previous),
	})
	if database.IsCheckViolation(err, openLimitConstraint) {
		return domain.ErrOpenLimit
	}
	if err != nil {
		return fmt.Errorf("save change of need %s: %w", change.NeedID, err)
	}
	if saved == 0 {
		return domain.ErrChangedMeanwhile
	}
	return nil
}

func (r *Needs) ExpireOpen(ctx context.Context, now time.Time) (int, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	expired, err := sqlc.New(r.pool.Querier(ctx)).ExpireOpenNeeds(writeCtx, now)
	if err != nil {
		return 0, fmt.Errorf("expire open needs: %w", err)
	}
	return int(expired), nil
}

func toNeed(row sqlc.GetNeedRow) domain.Need {
	return domain.Need{
		ID: row.ID, StoreID: row.StoreOrgID, StoreName: row.StoreName, CategoryName: row.CategoryName,
		Draft: domain.Draft{
			Title: row.Title, CategoryID: row.CategoryID, Description: row.Description,
			WeeklyQuantity: amountOf4(row.EstQtyPerWeek), MaxRetailPrice: amountOf8(row.MaxRetailPrice),
			PreferredModel: row.PreferredModel, OfferedCommissionBPS: amountOf4(row.OfferedCommissionBps),
		},
		Status: domain.Status(row.Status), ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt,
		ResponseCount: int(row.ResponseCount),
	}
}

func int8Of(amount domain.Amount) pgtype.Int8 {
	return pgtype.Int8{Int64: amount.Value, Valid: amount.IsGiven}
}

func amountOf4(value pgtype.Int4) domain.Amount {
	return domain.Amount{Value: int64(value.Int32), IsGiven: value.Valid}
}

func amountOf8(value pgtype.Int8) domain.Amount {
	return domain.Amount{Value: value.Int64, IsGiven: value.Valid}
}
