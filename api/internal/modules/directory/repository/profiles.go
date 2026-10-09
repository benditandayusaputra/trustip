package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Directory) GetStoreProfile(ctx context.Context, slug string) (domain.StoreProfile, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.GetPublicStoreProfile(readCtx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.StoreProfile{}, apperror.NotFound()
	}
	if err != nil {
		return domain.StoreProfile{}, fmt.Errorf("get public store %s: %w", slug, err)
	}
	needs, err := queries.ListPublicStoreNeeds(readCtx, row.ID)
	if err != nil {
		return domain.StoreProfile{}, fmt.Errorf("list open needs of store %s: %w", slug, err)
	}
	profile := toStoreProfile(row)
	profile.OpenNeeds = make([]domain.OpenNeed, 0, len(needs))
	for _, need := range needs {
		profile.OpenNeeds = append(profile.OpenNeeds, domain.OpenNeed{
			ID: need.ID, Title: need.Title, CategoryName: need.CategoryName, ExpiresAt: need.ExpiresAt,
		})
	}
	return profile, nil
}

func toStoreProfile(row sqlc.GetPublicStoreProfileRow) domain.StoreProfile {
	return domain.StoreProfile{
		ID: row.ID, Slug: row.Slug, Name: row.Name, Description: row.Description, Address: row.Address,
		District: row.District, CityName: row.CityName,
		Location: domain.Place{
			Point: domain.Point{Latitude: row.Latitude, Longitude: row.Longitude}, IsKnown: row.HasLocation,
		},
		LogoStorageKey: row.LogoStorageKey, PhotoStorageKeys: nonNull(row.PhotoStorageKeys),
		StoreType: row.StoreType, OpeningHours: row.OpeningHours,
		AcceptedCategories: categoriesOf(row.AcceptedCategoryIds, row.AcceptedCategoryNames),
		Terms: domain.StoreTerms{
			AcceptsFixedDeposit: row.AcceptsFixedDeposit, AcceptsCommission: row.AcceptsCommission,
			Commission: domain.CommissionRange{
				MinBPS: int(row.CommissionMinBps), MaxBPS: int(row.CommissionMaxBps), IsKnown: row.HasCommissionRange,
			},
			RequiresDistributionPermit: row.RequiresDistributionPermit, RequiresHalal: row.RequiresHalal,
			RequiresSample: row.RequiresSample,
		},
		OtherRequirements: row.OtherRequirements,
		Reputation: domain.StoreReputation{
			DueSettlements: int(row.DueSettlements), OnTimeBPS: int(row.OnTimeBps),
			ActivePartners: int(row.ActivePartners),
		},
		JoinedAt: row.CreatedAt,
	}
}

func (r *Directory) GetConsignorProfile(ctx context.Context, slug string) (domain.ConsignorProfile, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.GetPublicConsignorProfile(readCtx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ConsignorProfile{}, apperror.NotFound()
	}
	if err != nil {
		return domain.ConsignorProfile{}, fmt.Errorf("get public consignor %s: %w", slug, err)
	}
	products, err := queries.ListPublicConsignorProducts(readCtx, sqlc.ListPublicConsignorProductsParams{
		ConsignorOrgID: row.ID, RowLimit: domain.PublicProductLimit,
	})
	if err != nil {
		return domain.ConsignorProfile{}, fmt.Errorf("list products of consignor %s: %w", slug, err)
	}
	profile := toConsignorProfile(row)
	profile.Products = make([]domain.PublicProduct, 0, len(products))
	for _, product := range products {
		profile.Products = append(profile.Products, domain.PublicProduct{
			SKU: product.Sku, Name: product.Name, Unit: product.Unit, CategoryName: product.CategoryName,
			SuggestedRetailPrice: product.SuggestedRetailPrice, PhotoStorageKey: product.PhotoStorageKey,
		})
	}
	return profile, nil
}

func toConsignorProfile(row sqlc.GetPublicConsignorProfileRow) domain.ConsignorProfile {
	return domain.ConsignorProfile{
		ID: row.ID, Slug: row.Slug, Name: row.Name, Description: row.Description, District: row.District,
		CityName: row.CityName,
		Location: domain.Place{
			Point: domain.Point{Latitude: row.Latitude, Longitude: row.Longitude}, IsKnown: row.HasLocation,
		},
		LogoStorageKey: row.LogoStorageKey,
		Legality: domain.DeclaredLegality{
			BusinessIDNo: row.BusinessIDNo, DistributionPermitNo: row.DistributionPermitNo,
			HalalCertNo: row.HalalCertNo,
		},
		ActivePartners: int(row.ActivePartners), JoinedAt: row.CreatedAt,
	}
}
