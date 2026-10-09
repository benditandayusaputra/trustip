package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

type StoreProfiles interface {
	GetStoreProfile(ctx context.Context, slug string) (domain.StoreProfile, error)
}

type GetStoreProfile struct {
	profiles StoreProfiles
}

func NewGetStoreProfile(profiles StoreProfiles) *GetStoreProfile {
	return &GetStoreProfile{profiles: profiles}
}

func (uc *GetStoreProfile) Execute(ctx context.Context, slug string) (domain.StoreProfile, error) {
	if !domain.IsSlug(slug) {
		return domain.StoreProfile{}, apperror.NotFound()
	}
	profile, err := uc.profiles.GetStoreProfile(ctx, slug)
	if err != nil {
		return domain.StoreProfile{}, fmt.Errorf("get store profile %s: %w", slug, err)
	}
	return profile, nil
}
