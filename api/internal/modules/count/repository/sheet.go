package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const activeStatus = "ACTIVE"

type Counts struct {
	pool *database.Pool
}

func NewCounts(pool *database.Pool) *Counts {
	return &Counts{pool: pool}
}

func (r *Counts) LockSheet(ctx context.Context, key domain.SheetKey) (domain.Sheet, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	_, err := sqlc.New(r.pool.Querier(ctx)).LockPartnershipRow(writeCtx, sqlc.LockPartnershipRowParams{
		PartnershipID: key.PartnershipID, StoreOrgID: key.StoreID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Sheet{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Sheet{}, fmt.Errorf("lock partnership %s: %w", key.PartnershipID, err)
	}
	return r.GetSheet(ctx, key)
}

func (r *Counts) GetSheet(ctx context.Context, key domain.SheetKey) (domain.Sheet, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.GetCountPartnership(readCtx, sqlc.GetCountPartnershipParams{
		PartnershipID: key.PartnershipID, StoreOrgID: key.StoreID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Sheet{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Sheet{}, fmt.Errorf("get partnership %s: %w", key.PartnershipID, err)
	}
	sheet := domain.Sheet{
		PartnershipID: row.ID, IsActive: row.Status == activeStatus && row.AgreementVersion > 0,
		AgreementVersion: int(row.AgreementVersion), StoreID: row.StoreOrgID, StoreName: row.StoreName,
		ConsignorID: row.ConsignorOrgID, ConsignorName: row.ConsignorName,
	}
	if sheet.LastDocument, err = lastDocument(readCtx, queries, key.PartnershipID); err != nil {
		return domain.Sheet{}, err
	}
	if sheet.Settled, err = settled(readCtx, queries, key.PartnershipID); err != nil {
		return domain.Sheet{}, err
	}
	sheet.Products, err = stockedProducts(readCtx, queries, key.PartnershipID)
	return sheet, err
}

func lastDocument(ctx context.Context, queries *sqlc.Queries, partnershipID string) (domain.LastDocument, error) {
	row, err := queries.GetLastDocumentAt(ctx, partnershipID)
	if err != nil {
		return domain.LastDocument{}, fmt.Errorf("find last document of %s: %w", partnershipID, err)
	}
	return domain.LastDocument{At: row.LastAt, Given: row.HasDocument}, nil
}

func settled(ctx context.Context, queries *sqlc.Queries, partnershipID string) (domain.Settled, error) {
	row, err := queries.GetSettledUntil(ctx, partnershipID)
	if err != nil {
		return domain.Settled{}, fmt.Errorf("find settled period of %s: %w", partnershipID, err)
	}
	return domain.Settled{Until: row.PeriodEnd, Number: row.Number, Given: row.HasSettlement}, nil
}

func stockedProducts(
	ctx context.Context, queries *sqlc.Queries, partnershipID string,
) ([]domain.StockedProduct, error) {
	rows, err := queries.ListCountBatches(ctx, partnershipID)
	if err != nil {
		return nil, fmt.Errorf("list batches of %s: %w", partnershipID, err)
	}
	products := []domain.StockedProduct{}
	for _, row := range rows {
		last := len(products) - 1
		if last < 0 || products[last].ProductID != row.ProductID {
			products = append(products, domain.StockedProduct{
				ProductID: row.ProductID, Name: row.ProductName, SKU: row.Sku, Unit: row.Unit,
			})
			last++
		}
		products[last].Batches = append(products[last].Batches, pricedBatchOf(row))
	}
	return products, nil
}

func pricedBatchOf(row sqlc.ListCountBatchesRow) domain.PricedBatch {
	expiry := stock.NoExpiry()
	if date, err := time.Parse(time.DateOnly, row.ExpiryDate); err == nil {
		expiry = stock.ExpiresOn(date)
	}
	return domain.PricedBatch{
		Stock: stock.Batch{ID: row.DropItemID, Expiry: expiry, ReceivedAt: row.ReceivedAt, Balance: int(row.Balance)},
		Terms: pricing.Terms{
			Model: pricing.Model(row.PricingModel), RetailPrice: money.Rupiah(row.RetailPrice),
			CommissionBPS: int(row.CommissionBps), DepositPrice: money.Rupiah(row.DepositPrice),
		},
	}
}
