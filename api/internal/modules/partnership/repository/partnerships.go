package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const nilID = "00000000-0000-0000-0000-000000000000"

type Partnerships struct {
	pool *database.Pool
}

func NewPartnerships(pool *database.Pool) *Partnerships {
	return &Partnerships{pool: pool}
}

func (r *Partnerships) ListPartnerships(ctx context.Context, query domain.ListQuery) ([]domain.Summary, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	afterID := query.After.ID
	if query.After.IsStart() {
		afterID = nilID
	}
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListPartnerships(readCtx, sqlc.ListPartnershipsParams{
		OrganizationID: query.OrganizationID, Status: string(query.Status), FromStart: query.After.IsStart(),
		AfterAt: query.After.At, AfterID: afterID, RowLimit: int64(query.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list partnerships of %s: %w", query.OrganizationID, err)
	}
	summaries := make([]domain.Summary, 0, len(rows))
	for _, row := range rows {
		summary := domain.Summary{
			ID: row.ID, Status: domain.Status(row.Status), StartedAt: row.StartedAt, CounterpartID: row.CounterpartID,
			CounterpartName: row.CounterpartName, CounterpartType: actor.OrganizationType(row.CounterpartType),
		}
		if row.HasEnded {
			summary.EndedAt = row.EndedAt
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func (r *Partnerships) GetPartnership(ctx context.Context, key domain.Key) (domain.Partnership, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetPartnership(readCtx, sqlc.GetPartnershipParams{
		ID: key.PartnershipID, OrganizationID: key.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Partnership{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Partnership{}, fmt.Errorf("get partnership %s: %w", key.PartnershipID, err)
	}
	partnership := domain.Partnership{
		ID: row.ID, Status: domain.Status(row.Status), StartedAt: row.StartedAt,
		Counterpart: domain.Counterpart{
			ID: row.CounterpartID, Type: actor.OrganizationType(row.CounterpartType), Name: row.CounterpartName,
			Slug: row.CounterpartSlug, Description: row.Description, WhatsApp: row.Whatsapp, CityName: row.CityName,
			District: row.District, Address: row.Address, HasLocation: row.HasLocation,
			Exact:   domain.Coordinates{Latitude: row.Latitude, Longitude: row.Longitude},
			Public:  domain.Coordinates{Latitude: row.PublicLatitude, Longitude: row.PublicLongitude},
			LogoKey: row.LogoStorageKey,
		},
	}
	if row.HasEnded {
		partnership.EndedAt = row.EndedAt
	}
	return partnership, nil
}

func (r *Partnerships) ListCatalog(ctx context.Context, key domain.Key) ([]domain.CatalogItem, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListPartnerCatalog(readCtx, sqlc.ListPartnerCatalogParams{
		PartnershipID: key.PartnershipID, OrganizationID: key.OrganizationID,
	})
	if err != nil {
		return nil, fmt.Errorf("list catalog of %s: %w", key.PartnershipID, err)
	}
	items := make([]domain.CatalogItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.CatalogItem{
			ProductID: row.ID, Name: row.Name, SKU: row.Sku, Unit: row.Unit, Description: row.Description,
			NetContent: row.NetContent, CategoryName: row.CategoryName, PhotoKey: row.PhotoStorageKey,
			IsAgreed: row.IsAgreed, IsActive: row.IsActive,
			Terms: pricing.Terms{
				Model: pricing.Model(row.PricingModel), RetailPrice: money.Rupiah(row.RetailPrice),
				CommissionBPS: int(row.CommissionBps), DepositPrice: money.Rupiah(row.DepositPrice),
			},
		})
	}
	return items, nil
}
