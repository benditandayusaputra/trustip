package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const activeStatus = "ACTIVE"

type Drops struct {
	pool *database.Pool
}

func NewDrops(pool *database.Pool) *Drops {
	return &Drops{pool: pool}
}

func (r *Drops) NextNumber(ctx context.Context, key domain.SequenceKey) (int, error) {
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

func (r *Drops) GetSetting(ctx context.Context, key domain.SettingKey) (domain.Setting, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.GetDropPartnership(readCtx, sqlc.GetDropPartnershipParams{
		PartnershipID: key.PartnershipID, StoreOrgID: key.StoreID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Setting{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Setting{}, fmt.Errorf("get partnership %s: %w", key.PartnershipID, err)
	}
	setting := domain.Setting{
		PartnershipID: row.ID, IsActive: row.Status == activeStatus && row.AgreementID != "",
		AgreementID: row.AgreementID, AgreementVersion: int(row.AgreementVersion), StoreID: row.StoreOrgID,
		StoreName: row.StoreName, ConsignorID: row.ConsignorOrgID, Products: map[string]domain.AgreedProduct{},
	}
	if !setting.IsActive {
		return setting, nil
	}
	return withProducts(readCtx, queries, setting)
}

func withProducts(ctx context.Context, queries *sqlc.Queries, setting domain.Setting) (domain.Setting, error) {
	rows, err := queries.ListDropProducts(ctx, setting.AgreementID)
	if err != nil {
		return domain.Setting{}, fmt.Errorf("list products of agreement %s: %w", setting.AgreementID, err)
	}
	for _, row := range rows {
		shelfLife := stock.UnknownShelfLife()
		if row.ShelfLifeDays.Valid {
			shelfLife = stock.ShelfLifeOf(int(row.ShelfLifeDays.Int32))
		}
		setting.Products[row.ProductID] = domain.AgreedProduct{
			ProductID: row.ProductID, ShelfLife: shelfLife,
			Terms: termsOf(termsColumns{
				model: row.PricingModel, retail: row.RetailPrice,
				commission: row.CommissionBps, deposit: row.DepositPrice,
			}),
		}
	}
	return setting, nil
}

type termsColumns struct {
	model      string
	retail     int64
	commission int32
	deposit    int64
}

func termsOf(columns termsColumns) pricing.Terms {
	return pricing.Terms{
		Model: pricing.Model(columns.model), RetailPrice: moneyOf(columns.retail),
		CommissionBPS: int(columns.commission), DepositPrice: moneyOf(columns.deposit),
	}
}

func (r *Drops) FindByKey(ctx context.Context, key domain.IdempotencyKey) (domain.Stored, bool, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).FindDropByIdempotencyKey(readCtx, sqlc.FindDropByIdempotencyKeyParams{
		StoreOrgID: key.StoreID, IdempotencyKey: key.Key,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Stored{}, false, nil
	}
	if err != nil {
		return domain.Stored{}, false, fmt.Errorf("find drop by idempotency key: %w", err)
	}
	return domain.Stored{DropID: row.ID, Fingerprint: row.RequestFingerprint}, true, nil
}

func (r *Drops) CreateDrop(ctx context.Context, drop domain.Drop) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	inserted, err := queries.CreateDrop(writeCtx, sqlc.CreateDropParams{
		ID: drop.ID, Number: drop.Number, PartnershipID: drop.PartnershipID, AgreementID: drop.AgreementID,
		StoreOrgID: drop.StoreID, OccurredAt: drop.OccurredAt, RecordedBy: drop.RecordedBy,
		IdempotencyKey: drop.IdempotencyKey, RequestFingerprint: drop.Fingerprint, ReceivedAt: drop.ReceivedAt,
	})
	if err != nil {
		return fmt.Errorf("create drop %s: %w", drop.ID, err)
	}
	if inserted == 0 {
		return domain.ErrDuplicateKey
	}
	if err := queries.AddDropItems(writeCtx, itemParams(drop)); err != nil {
		return fmt.Errorf("add items to drop %s: %w", drop.ID, err)
	}
	if err := queries.AddStockMovements(writeCtx, movementParams(drop)); err != nil {
		return fmt.Errorf("receive stock of drop %s: %w", drop.ID, err)
	}
	return nil
}

func itemParams(drop domain.Drop) sqlc.AddDropItemsParams {
	params := sqlc.AddDropItemsParams{DropID: drop.ID}
	for index, item := range drop.Items {
		params.Ids = append(params.Ids, item.ID)
		params.LineNumbers = append(params.LineNumbers, int64(index+1))
		params.ProductIds = append(params.ProductIds, item.ProductID)
		params.BatchCodes = append(params.BatchCodes, item.BatchCode)
		params.ExpiryDates = append(params.ExpiryDates, expiryText(item.Expiry))
		params.Quantities = append(params.Quantities, int64(item.Qty))
		params.PricingModels = append(params.PricingModels, string(item.Terms.Model))
		params.RetailPrices = append(params.RetailPrices, item.Terms.RetailPrice.Int64())
		params.CommissionBps = append(params.CommissionBps, int64(item.Terms.CommissionBPS))
		params.DepositPrices = append(params.DepositPrices, item.Terms.DepositPrice.Int64())
	}
	return params
}

func movementParams(drop domain.Drop) sqlc.AddStockMovementsParams {
	params := sqlc.AddStockMovementsParams{
		PartnershipID: drop.PartnershipID, OccurredAt: drop.OccurredAt, SourceType: domain.EntityType,
		SourceID: drop.ID, SourceNumber: drop.Number, CreatedBy: drop.RecordedBy,
	}
	for _, movement := range drop.Movements() {
		params.Ids = append(params.Ids, movement.ID)
		params.ProductIds = append(params.ProductIds, movement.ProductID)
		params.DropItemIds = append(params.DropItemIds, movement.DropItemID)
		params.Types = append(params.Types, movement.Type)
		params.Quantities = append(params.Quantities, int64(movement.Qty))
	}
	return params
}

func expiryText(expiry stock.Expiry) string {
	date, isGiven := expiry.Date()
	if !isGiven {
		return ""
	}
	return date.Format(time.DateOnly)
}
