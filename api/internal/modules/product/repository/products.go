package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const (
	photoPurpose  = "PRODUCT_PHOTO"
	skuConstraint = "products_sku_key"
	nilID         = "00000000-0000-0000-0000-000000000000"
)

type Products struct {
	pool *database.Pool
}

func NewProducts(pool *database.Pool) *Products {
	return &Products{pool: pool}
}

func (r *Products) NextSKUNumber(ctx context.Context, series domain.SKUSeries) (int, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	if err := queries.LockConsignorForSKU(writeCtx, series.ConsignorID); err != nil {
		return 0, fmt.Errorf("lock consignor %s: %w", series.ConsignorID, err)
	}
	next, err := queries.NextSKUNumber(writeCtx, sqlc.NextSKUNumberParams{
		ConsignorOrgID: series.ConsignorID, Prefix: series.Prefix,
	})
	if err != nil {
		return 0, fmt.Errorf("next sku number of %s: %w", series.ConsignorID, err)
	}
	return int(next), nil
}

func (r *Products) SaveProduct(ctx context.Context, change domain.Change) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	if err := checkOwnPhoto(writeCtx, queries, change.ConsignorID, change.Details.PhotoFileID); err != nil {
		return err
	}
	details := change.Details
	saved, err := queries.SaveProduct(writeCtx, sqlc.SaveProductParams{
		ID: change.ProductID, ConsignorOrgID: change.ConsignorID, Sku: change.SKU, Name: details.Name,
		CategoryID: details.CategoryID, Unit: details.Unit, Description: details.Description,
		Composition: details.Composition, NetContent: details.NetContent, PhotoFileID: details.PhotoFileID,
		SuggestedRetailPrice: details.Prices.Retail.Int64(), SuggestedDepositPrice: details.Prices.Deposit.Int64(),
		ShelfLifeDays: shelfLifeParam(details.ShelfLife), ChangedAt: change.At,
	})
	return explainWrite(err, saved, change.ProductID)
}

func explainWrite(err error, affected int64, productID string) error {
	if database.IsUniqueViolation(err, skuConstraint) {
		return domain.ErrSKUTaken
	}
	if err != nil {
		return fmt.Errorf("write product %s: %w", productID, err)
	}
	if affected == 0 {
		return domain.ErrUnknownCategory
	}
	return nil
}

func checkOwnPhoto(ctx context.Context, queries *sqlc.Queries, consignorID, photoFileID string) error {
	if photoFileID == "" {
		return nil
	}
	found, err := queries.CountOrganizationFiles(ctx, sqlc.CountOrganizationFilesParams{
		ID: photoFileID, OrganizationID: consignorID, Purpose: photoPurpose,
	})
	if err != nil {
		return fmt.Errorf("check photo %s: %w", photoFileID, err)
	}
	if found == 0 {
		return domain.ErrUnknownPhoto
	}
	return nil
}

func shelfLifeParam(shelfLife domain.ShelfLife) pgtype.Int8 {
	return pgtype.Int8{Int64: int64(shelfLife.Days), Valid: shelfLife.Given}
}

func (r *Products) SetArchived(ctx context.Context, change domain.StatusChange) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	updated, err := sqlc.New(r.pool.Querier(ctx)).SetProductArchived(writeCtx, sqlc.SetProductArchivedParams{
		ID: change.ProductID, ConsignorOrgID: change.ConsignorID, IsArchived: change.IsArchived, UpdatedAt: change.At,
	})
	if err != nil {
		return fmt.Errorf("set archived of %s: %w", change.ProductID, err)
	}
	if updated == 0 {
		return apperror.NotFound()
	}
	return nil
}

func (r *Products) GetProduct(ctx context.Context, key domain.ProductKey) (domain.Product, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetProduct(readCtx, sqlc.GetProductParams{
		ID: key.ProductID, ConsignorOrgID: key.ConsignorID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Product{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Product{}, fmt.Errorf("get product %s: %w", key.ProductID, err)
	}
	return toProduct(sqlc.ListProductsRow(row)), nil
}

func (r *Products) ListProducts(ctx context.Context, query domain.PageQuery) ([]domain.Product, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	afterID := query.After.ID
	if query.After.IsStart() {
		afterID = nilID
	}
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListProducts(readCtx, sqlc.ListProductsParams{
		ConsignorOrgID: query.ConsignorID, IsArchived: query.Search.Status == domain.Archived,
		CategoryID: query.Search.CategoryID, Query: query.Search.Text, FromStart: query.After.IsStart(),
		AfterAt: query.After.At, AfterID: afterID, RowLimit: int64(query.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list products of %s: %w", query.ConsignorID, err)
	}
	products := make([]domain.Product, 0, len(rows))
	for _, row := range rows {
		products = append(products, toProduct(row))
	}
	return products, nil
}

func toProduct(row sqlc.ListProductsRow) domain.Product {
	return domain.Product{
		ID: row.ID, ConsignorID: row.ConsignorOrgID, SKU: row.Sku,
		Details: domain.Details{
			Name: row.Name, CategoryID: row.CategoryID, Unit: row.Unit, Description: row.Description,
			Composition: row.Composition, NetContent: row.NetContent, PhotoFileID: row.PhotoFileID,
			Prices: domain.Prices{
				Retail: money.Rupiah(row.SuggestedRetailPrice), Deposit: money.Rupiah(row.SuggestedDepositPrice),
			},
			ShelfLife: domain.ShelfLife{Given: row.ShelfLifeDays.Valid, Days: int(row.ShelfLifeDays.Int32)},
		},
		CategoryName: row.CategoryName,
		Photo:        domain.FileRef{ID: row.PhotoFileID, StorageKey: row.PhotoStorageKey},
		IsArchived:   row.IsArchived,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
