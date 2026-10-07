package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Profiles) UpdateStoreProfile(ctx context.Context, change domain.StoreProfileChange) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	if err := saveStoreProfile(writeCtx, queries, change); err != nil {
		return err
	}
	if err := replaceCategories(writeCtx, queries, change); err != nil {
		return err
	}
	return replacePhotos(writeCtx, queries, change)
}

func saveStoreProfile(ctx context.Context, queries *sqlc.Queries, change domain.StoreProfileChange) error {
	profile := change.Profile
	updated, err := queries.UpdateStoreProfile(ctx, sqlc.UpdateStoreProfileParams{
		OrganizationID:             change.OrganizationID,
		StoreType:                  profile.StoreType,
		OpeningHours:               profile.OpeningHours,
		AcceptsFixedDeposit:        profile.AcceptsFixedDeposit,
		AcceptsCommission:          profile.AcceptsCommission,
		HasCommission:              profile.Commission.Given,
		CommissionMinBps:           int64(profile.Commission.MinBPS),
		CommissionMaxBps:           int64(profile.Commission.MaxBPS),
		RequiresDistributionPermit: profile.Conditions.RequiresDistributionPermit,
		RequiresHalal:              profile.Conditions.RequiresHalal,
		RequiresSample:             profile.Conditions.RequiresSample,
		OtherRequirements:          profile.Conditions.OtherRequirements,
		UpdatedAt:                  change.At,
	})
	if err != nil {
		return fmt.Errorf("update store profile of %s: %w", change.OrganizationID, err)
	}
	if updated == 0 {
		return fmt.Errorf("store profile of %s is missing: %w", change.OrganizationID, pgx.ErrNoRows)
	}
	return nil
}

func replaceCategories(ctx context.Context, queries *sqlc.Queries, change domain.StoreProfileChange) error {
	if err := queries.ClearAcceptedCategories(ctx, change.OrganizationID); err != nil {
		return fmt.Errorf("clear categories of %s: %w", change.OrganizationID, err)
	}
	added, err := queries.AddAcceptedCategories(ctx, sqlc.AddAcceptedCategoriesParams{
		OrganizationID: change.OrganizationID, CategoryIds: change.Profile.CategoryIDs,
	})
	if err != nil {
		return fmt.Errorf("add categories of %s: %w", change.OrganizationID, err)
	}
	if added != int64(len(change.Profile.CategoryIDs)) {
		return domain.ErrUnknownCategory
	}
	return nil
}

func replacePhotos(ctx context.Context, queries *sqlc.Queries, change domain.StoreProfileChange) error {
	if err := queries.ClearOrganizationPhotos(ctx, change.OrganizationID); err != nil {
		return fmt.Errorf("clear photos of %s: %w", change.OrganizationID, err)
	}
	added, err := queries.AddOrganizationPhotos(ctx, sqlc.AddOrganizationPhotosParams{
		OrganizationID: change.OrganizationID, FileIds: change.Profile.PhotoFileIDs,
	})
	if err != nil {
		return fmt.Errorf("add photos of %s: %w", change.OrganizationID, err)
	}
	if added != int64(len(change.Profile.PhotoFileIDs)) {
		return domain.ErrUnknownPhoto
	}
	return nil
}

func storeSettings(ctx context.Context, queries *sqlc.Queries, organizationID string) (domain.StoreSettings, error) {
	row, err := queries.GetStoreProfile(ctx, organizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.StoreSettings{}, nil
	}
	if err != nil {
		return domain.StoreSettings{}, fmt.Errorf("get store profile of %s: %w", organizationID, err)
	}
	photos, err := queries.ListOrganizationPhotos(ctx, organizationID)
	if err != nil {
		return domain.StoreSettings{}, fmt.Errorf("list photos of %s: %w", organizationID, err)
	}
	return toStoreSettings(row, photos), nil
}

func toStoreSettings(row sqlc.GetStoreProfileRow, photos []sqlc.ListOrganizationPhotosRow) domain.StoreSettings {
	refs := make([]domain.FileRef, 0, len(photos))
	photoIDs := make([]string, 0, len(photos))
	for _, photo := range photos {
		refs = append(refs, domain.FileRef{ID: photo.ID, StorageKey: photo.StorageKey})
		photoIDs = append(photoIDs, photo.ID)
	}
	return domain.StoreSettings{Photos: refs, Profile: domain.StoreProfile{
		StoreType:   row.StoreType,
		CategoryIDs: row.CategoryIds,
		Acceptance: domain.Acceptance{
			AcceptsFixedDeposit: row.AcceptsFixedDeposit, AcceptsCommission: row.AcceptsCommission,
		},
		Commission: domain.CommissionRange{
			Given: row.HasCommission, MinBPS: int(row.CommissionMinBps), MaxBPS: int(row.CommissionMaxBps),
		},
		Conditions: domain.StoreConditions{
			RequiresDistributionPermit: row.RequiresDistributionPermit, RequiresHalal: row.RequiresHalal,
			RequiresSample: row.RequiresSample, OtherRequirements: row.OtherRequirements,
		},
		OpeningHours: row.OpeningHours,
		PhotoFileIDs: photoIDs,
	}}
}
