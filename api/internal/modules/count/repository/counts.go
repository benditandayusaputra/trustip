package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Counts) NextNumber(ctx context.Context, key domain.SequenceKey) (int, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	next, err := sqlc.New(r.pool.Querier(ctx)).NextDocumentNumber(writeCtx, sqlc.NextDocumentNumberParams{
		OrganizationID: key.OrganizationID, Prefix: key.Prefix, Period: key.Period,
	})
	if err != nil {
		return 0, fmt.Errorf("next %s number of %s: %w", key.Prefix, key.OrganizationID, err)
	}
	return int(next), nil
}

func (r *Counts) FindByKey(ctx context.Context, key domain.IdempotencyKey) (domain.Stored, bool, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).FindCountByIdempotencyKey(readCtx,
		sqlc.FindCountByIdempotencyKeyParams{StoreOrgID: key.StoreID, IdempotencyKey: key.Key})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Stored{}, false, nil
	}
	if err != nil {
		return domain.Stored{}, false, fmt.Errorf("find count by idempotency key: %w", err)
	}
	return domain.Stored{CountID: row.ID, Fingerprint: row.RequestFingerprint}, true, nil
}

func (r *Counts) CreateCount(ctx context.Context, count domain.Count) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	inserted, err := queries.CreateCount(writeCtx, countParams(count))
	if err != nil {
		return fmt.Errorf("create count %s: %w", count.ID, err)
	}
	if inserted == 0 {
		return domain.ErrDuplicateKey
	}
	steps := []func() error{
		func() error { return queries.AddCountLines(writeCtx, lineParams(count)) },
		func() error { return queries.AddSaleAllocations(writeCtx, saleParams(count)) },
		func() error { return queries.AddReturnAllocations(writeCtx, returnParams(count)) },
		func() error { return queries.AddStockMovements(writeCtx, movementParams(count)) },
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return fmt.Errorf("save count %s: %w", count.ID, err)
		}
	}
	return nil
}

func countParams(count domain.Count) sqlc.CreateCountParams {
	return sqlc.CreateCountParams{
		ID: count.ID, Number: count.Number, PartnershipID: count.PartnershipID, StoreOrgID: count.StoreID,
		Occasion: string(count.Occasion), CountedAt: count.CountedAt, Note: count.Note,
		SoldQty: int64(count.SoldQty), ReturnedQty: int64(count.ReturnedQty), GrossAmount: count.Split.Gross.Int64(),
		StoreShareAmount: count.Split.StoreShare.Int64(), ConsignorShareAmount: count.Split.ConsignorShare.Int64(),
		RecordedBy: count.RecordedBy, IdempotencyKey: count.IdempotencyKey, RequestFingerprint: count.Fingerprint,
		ClientRecordedAt: count.ClientRecordedAt, ReceivedAt: count.ReceivedAt, RecordedOffline: count.RecordedOffline,
	}
}

func lineParams(count domain.Count) sqlc.AddCountLinesParams {
	params := sqlc.AddCountLinesParams{CountID: count.ID}
	for _, line := range count.Lines {
		params.Ids = append(params.Ids, line.ID)
		params.LineNumbers = append(params.LineNumbers, int64(line.Ordinal))
		params.ProductIds = append(params.ProductIds, line.ProductID)
		params.Recorded = append(params.Recorded, int64(line.RecordedQty))
		params.Counted = append(params.Counted, int64(line.CountedQty))
		params.Sold = append(params.Sold, int64(line.SoldQty))
		params.Returned = append(params.Returned, int64(line.ReturnedQty))
		params.Kept = append(params.Kept, int64(line.KeptQty))
		params.Reasons = append(params.Reasons, string(line.ReturnReason))
	}
	return params
}

func saleParams(count domain.Count) sqlc.AddSaleAllocationsParams {
	params := sqlc.AddSaleAllocationsParams{PartnershipID: count.PartnershipID, OccurredAt: count.CountedAt}
	for _, line := range count.Lines {
		for _, sale := range line.Sold {
			params.Ids = append(params.Ids, sale.ID)
			params.CountLineIds = append(params.CountLineIds, line.ID)
			params.DropItemIds = append(params.DropItemIds, sale.BatchID)
			params.ProductIds = append(params.ProductIds, line.ProductID)
			params.Quantities = append(params.Quantities, int64(sale.Qty))
			params.GrossAmounts = append(params.GrossAmounts, sale.Split.Gross.Int64())
			params.StoreShares = append(params.StoreShares, sale.Split.StoreShare.Int64())
			params.ConsignorShares = append(params.ConsignorShares, sale.Split.ConsignorShare.Int64())
		}
	}
	return params
}

func returnParams(count domain.Count) sqlc.AddReturnAllocationsParams {
	params := sqlc.AddReturnAllocationsParams{}
	for _, line := range count.Lines {
		for _, taken := range line.Taken {
			params.Ids = append(params.Ids, taken.ID)
			params.CountLineIds = append(params.CountLineIds, line.ID)
			params.DropItemIds = append(params.DropItemIds, taken.BatchID)
			params.Quantities = append(params.Quantities, int64(taken.Qty))
		}
	}
	return params
}

func movementParams(count domain.Count) sqlc.AddStockMovementsParams {
	params := sqlc.AddStockMovementsParams{
		PartnershipID: count.PartnershipID, OccurredAt: count.CountedAt, SourceType: domain.EntityType,
		SourceID: count.ID, SourceNumber: count.Number, CreatedBy: count.RecordedBy,
	}
	for _, movement := range count.Movements() {
		params.Ids = append(params.Ids, movement.ID)
		params.ProductIds = append(params.ProductIds, movement.ProductID)
		params.DropItemIds = append(params.DropItemIds, movement.DropItemID)
		params.Types = append(params.Types, movement.Type)
		params.Quantities = append(params.Quantities, int64(movement.Qty))
	}
	return params
}
