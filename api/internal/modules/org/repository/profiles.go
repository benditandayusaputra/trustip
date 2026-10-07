package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/contact"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const logoPurpose = "LOGO"

type Profiles struct {
	pool *database.Pool
}

func NewProfiles(pool *database.Pool) *Profiles {
	return &Profiles{pool: pool}
}

func (r *Profiles) GetSettings(ctx context.Context, organizationID string) (domain.Settings, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.GetOrganizationSettings(readCtx, organizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Settings{}, apperror.NotFound()
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("get settings of %s: %w", organizationID, err)
	}
	settings, err := toSettings(row)
	if err != nil {
		return domain.Settings{}, err
	}
	settings.Store, err = storeSettings(readCtx, queries, organizationID)
	return settings, err
}

func toSettings(row sqlc.GetOrganizationSettingsRow) (domain.Settings, error) {
	whatsApp, err := contact.ParseWhatsAppNumber(row.Whatsapp)
	if err != nil {
		return domain.Settings{}, fmt.Errorf("read stored whatsapp of %s: %w", row.ID, err)
	}
	return domain.Settings{
		ID:   row.ID,
		Type: actor.OrganizationType(row.Type),
		Slug: domain.Slug(row.Slug),
		Profile: domain.Profile{
			Name: row.Name, Description: row.Description, WhatsApp: whatsApp, CityCode: row.CityCode,
			District: row.District, Address: row.Address, HasLocation: row.HasLocation,
			Location: domain.Location{
				Exact:  domain.Coordinates{Latitude: row.Latitude, Longitude: row.Longitude},
				Public: domain.Coordinates{Latitude: row.PublicLatitude, Longitude: row.PublicLongitude},
			},
		},
		CityName: row.CityName,
		Logo:     domain.FileRef{ID: row.LogoFileID, StorageKey: row.LogoStorageKey},
	}, nil
}

func (r *Profiles) UpdateProfile(ctx context.Context, change domain.ProfileChange) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	if err := checkLogo(writeCtx, queries, change); err != nil {
		return err
	}
	update := change.Update
	updated, err := queries.UpdateOrganizationProfile(writeCtx, sqlc.UpdateOrganizationProfileParams{
		ID: change.OrganizationID, Name: update.Name, Description: update.Description,
		Whatsapp: update.WhatsApp.String(), CityCode: update.CityCode, District: update.District,
		Address: update.Address, HasLocation: update.HasLocation,
		Latitude: update.Location.Exact.Latitude, Longitude: update.Location.Exact.Longitude,
		PublicLatitude: update.Location.Public.Latitude, PublicLongitude: update.Location.Public.Longitude,
		LogoFileID: update.LogoFileID, UpdatedAt: change.At,
	})
	if err != nil {
		return fmt.Errorf("update profile of %s: %w", change.OrganizationID, err)
	}
	if updated == 0 {
		return domain.ErrUnknownCity
	}
	return nil
}

func checkLogo(ctx context.Context, queries *sqlc.Queries, change domain.ProfileChange) error {
	if change.Update.LogoFileID == "" {
		return nil
	}
	found, err := queries.CountOrganizationFiles(ctx, sqlc.CountOrganizationFilesParams{
		ID: change.Update.LogoFileID, OrganizationID: change.OrganizationID, Purpose: logoPurpose,
	})
	if err != nil {
		return fmt.Errorf("check logo of %s: %w", change.OrganizationID, err)
	}
	if found == 0 {
		return domain.ErrUnknownLogo
	}
	return nil
}
