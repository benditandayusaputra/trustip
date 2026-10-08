package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
	"github.com/benditandayusaputra/trustip/api/internal/modules/stock/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

type Stocks struct {
	pool *database.Pool
}

func NewStocks(pool *database.Pool) *Stocks {
	return &Stocks{pool: pool}
}

func (r *Stocks) ListPartnershipStock(ctx context.Context, key domain.PartnershipKey) ([]domain.Product, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	if err := findPartnership(readCtx, queries, key); err != nil {
		return nil, err
	}
	rows, err := queries.ListPartnershipBatches(readCtx, sqlc.ListPartnershipBatchesParams{
		PartnershipID: key.PartnershipID, OrganizationID: key.OrganizationID,
	})
	if err != nil {
		return nil, fmt.Errorf("list batches of %s: %w", key.PartnershipID, err)
	}
	grouped := productGroups{}
	for _, row := range rows {
		grouped.add(row.ProductID, productHeader{
			name: row.ProductName, sku: row.Sku, unit: row.Unit, shelfLife: row.ShelfLifeDays,
		}, domain.Batch{
			ID: row.DropItemID, Code: row.BatchCode, Expiry: expiryOf(row.ExpiryDate), ReceivedAt: row.ReceivedAt,
			Balance: int(row.Balance),
		})
	}
	return grouped.products, nil
}

func findPartnership(ctx context.Context, queries *sqlc.Queries, key domain.PartnershipKey) error {
	_, err := queries.FindStockPartnership(ctx, sqlc.FindStockPartnershipParams{
		PartnershipID: key.PartnershipID, OrganizationID: key.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.OutsideScope()
	}
	if err != nil {
		return fmt.Errorf("find partnership %s: %w", key.PartnershipID, err)
	}
	return nil
}

func (r *Stocks) ListOrganizationStock(ctx context.Context, organizationID string) ([]domain.Product, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListOrganizationBatches(readCtx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list batches of %s: %w", organizationID, err)
	}
	grouped := productGroups{}
	for _, row := range rows {
		grouped.add(row.PartnershipID+"/"+row.ProductID, productHeader{
			id: row.ProductID, name: row.ProductName, sku: row.Sku, unit: row.Unit, shelfLife: row.ShelfLifeDays,
			counterpart: domain.Counterpart{
				PartnershipID: row.PartnershipID, ID: row.CounterpartID, Name: row.CounterpartName,
				Type: row.CounterpartType,
			},
		}, domain.Batch{
			ID: row.DropItemID, Expiry: expiryOf(row.ExpiryDate), ReceivedAt: row.ReceivedAt, Balance: int(row.Balance),
		})
	}
	return grouped.products, nil
}

type productHeader struct {
	id          string
	name        string
	sku         string
	unit        string
	shelfLife   int64
	counterpart domain.Counterpart
}

type productGroups struct {
	products []domain.Product
	keys     []string
}

func (g *productGroups) add(key string, header productHeader, batch domain.Batch) {
	last := len(g.keys) - 1
	if last < 0 || g.keys[last] != key {
		g.keys = append(g.keys, key)
		g.products = append(g.products, header.product(key))
		last++
	}
	g.products[last].Batches = append(g.products[last].Batches, batch)
}

func (h productHeader) product(key string) domain.Product {
	id := h.id
	if id == "" {
		id = key
	}
	shelfLife := stock.UnknownShelfLife()
	if h.shelfLife > 0 {
		shelfLife = stock.ShelfLifeOf(int(h.shelfLife))
	}
	return domain.Product{
		ID: id, Name: h.name, SKU: h.sku, Unit: h.unit, ShelfLife: shelfLife, Counterpart: h.counterpart,
	}
}

func expiryOf(text string) stock.Expiry {
	date, err := time.Parse(time.DateOnly, text)
	if err != nil {
		return stock.NoExpiry()
	}
	return stock.ExpiresOn(date)
}
