package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
)

type OrganizationFounder interface {
	TakenSlugs(ctx context.Context, base domain.Slug) ([]domain.Slug, error)
	CreateOrganization(ctx context.Context, founding domain.Founding) (string, error)
}

type Transactor interface {
	WithinTx(ctx context.Context, work func(context.Context) error) error
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() (string, error)
}

type RegionReader interface {
	ListProvinces(ctx context.Context) ([]domain.Province, error)
	ListCities(ctx context.Context, provinceCode string) ([]domain.City, error)
}

type ProfileStore interface {
	GetSettings(ctx context.Context, organizationID string) (domain.Settings, error)
	UpdateProfile(ctx context.Context, change domain.ProfileChange) error
}
