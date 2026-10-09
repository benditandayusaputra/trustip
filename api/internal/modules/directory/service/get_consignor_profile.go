package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

type ConsignorProfiles interface {
	GetConsignorProfile(ctx context.Context, slug string) (domain.ConsignorProfile, error)
}

type GetConsignorProfile struct {
	profiles ConsignorProfiles
}

func NewGetConsignorProfile(profiles ConsignorProfiles) *GetConsignorProfile {
	return &GetConsignorProfile{profiles: profiles}
}

func (uc *GetConsignorProfile) Execute(ctx context.Context, slug string) (domain.ConsignorProfile, error) {
	if !domain.IsSlug(slug) {
		return domain.ConsignorProfile{}, apperror.NotFound()
	}
	profile, err := uc.profiles.GetConsignorProfile(ctx, slug)
	if err != nil {
		return domain.ConsignorProfile{}, fmt.Errorf("get consignor profile %s: %w", slug, err)
	}
	return profile, nil
}
