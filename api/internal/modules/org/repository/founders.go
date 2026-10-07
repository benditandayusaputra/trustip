package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const (
	slugUniqueConstraint       = "organizations_slug_key"
	membershipUniqueConstraint = "memberships_user_id_key"
)

type Founders struct {
	pool *database.Pool
}

func NewFounders(pool *database.Pool) *Founders {
	return &Founders{pool: pool}
}

func (r *Founders) TakenSlugs(ctx context.Context, base domain.Slug) ([]domain.Slug, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListTakenSlugs(readCtx, string(base))
	if err != nil {
		return nil, fmt.Errorf("list slugs like %s: %w", base, err)
	}
	slugs := make([]domain.Slug, 0, len(rows))
	for _, slug := range rows {
		slugs = append(slugs, domain.Slug(slug))
	}
	return slugs, nil
}

func (r *Founders) CreateOrganization(ctx context.Context, founding domain.Founding) (string, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	cityName, err := insertOrganization(writeCtx, queries, founding.Organization)
	if err != nil {
		return "", err
	}
	if err := insertProfile(writeCtx, queries, founding.Organization); err != nil {
		return "", err
	}
	if err := insertOwner(writeCtx, queries, founding); err != nil {
		return "", err
	}
	return cityName, nil
}

func insertOrganization(ctx context.Context, queries *sqlc.Queries, organization domain.Organization) (string, error) {
	cityName, err := queries.CreateOrganization(ctx, sqlc.CreateOrganizationParams{
		ID:          organization.ID,
		Type:        string(organization.Type),
		Name:        organization.Name,
		Slug:        string(organization.Slug),
		Description: organization.Description,
		Whatsapp:    organization.WhatsApp.String(),
		CityCode:    organization.CityCode,
		District:    organization.District,
		Address:     organization.Address,
		CreatedAt:   organization.CreatedAt,
	})
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return "", domain.ErrUnknownCity
	case database.IsUniqueViolation(err, slugUniqueConstraint):
		return "", domain.ErrSlugTaken
	case err != nil:
		return "", fmt.Errorf("insert organization %s: %w", organization.ID, err)
	}
	return cityName, nil
}

func insertProfile(ctx context.Context, queries *sqlc.Queries, organization domain.Organization) error {
	var err error
	switch organization.Type {
	case actor.Store:
		err = queries.CreateStoreProfile(ctx, sqlc.CreateStoreProfileParams{
			OrganizationID: organization.ID, CreatedAt: organization.CreatedAt,
		})
	case actor.Consignor:
		err = queries.CreateConsignorProfile(ctx, sqlc.CreateConsignorProfileParams{
			OrganizationID: organization.ID, CreatedAt: organization.CreatedAt,
		})
	}
	if err != nil {
		return fmt.Errorf("insert profile of %s: %w", organization.ID, err)
	}
	return nil
}

func insertOwner(ctx context.Context, queries *sqlc.Queries, founding domain.Founding) error {
	err := queries.CreateMembership(ctx, sqlc.CreateMembershipParams{
		ID:             founding.Owner.ID,
		OrganizationID: founding.Organization.ID,
		UserID:         founding.Owner.UserID,
		Role:           string(founding.Owner.Role),
		CreatedAt:      founding.Organization.CreatedAt,
	})
	if database.IsUniqueViolation(err, membershipUniqueConstraint) {
		return domain.ErrAlreadyMember
	}
	if err != nil {
		return fmt.Errorf("insert owner of %s: %w", founding.Organization.ID, err)
	}
	return nil
}
