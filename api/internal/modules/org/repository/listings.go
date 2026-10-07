package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Profiles) GetListingFacts(ctx context.Context, organizationID string) (domain.ListingFacts, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	return listingFacts(readCtx, sqlc.New(r.pool.Querier(ctx)), organizationID)
}

func listingFacts(ctx context.Context, queries *sqlc.Queries, organizationID string) (domain.ListingFacts, error) {
	row, err := queries.GetListingFacts(ctx, organizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ListingFacts{}, apperror.NotFound()
	}
	if err != nil {
		return domain.ListingFacts{}, fmt.Errorf("get listing facts of %s: %w", organizationID, err)
	}
	return domain.ListingFacts{
		Type: actor.OrganizationType(row.Type), IsListed: row.IsListed, HasName: row.HasName,
		HasCity: row.HasCity, HasLocation: row.HasLocation, HasDescription: row.HasDescription,
		HasLogo: row.HasLogo, HasStoreType: row.HasStoreType, AcceptedCategories: int(row.AcceptedCategories),
		Photos: int(row.Photos), HasPaymentInfo: row.HasPaymentInfo, ActiveProducts: int(row.ActiveProducts),
	}, nil
}

func (r *Profiles) SetListed(ctx context.Context, change domain.ListingChange) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	err := sqlc.New(r.pool.Querier(ctx)).SetOrganizationListed(writeCtx, sqlc.SetOrganizationListedParams{
		ID: change.OrganizationID, IsListed: change.IsListed, UpdatedAt: change.At,
	})
	if err != nil {
		return fmt.Errorf("set listed of %s: %w", change.OrganizationID, err)
	}
	return nil
}
