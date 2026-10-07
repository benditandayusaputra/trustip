package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
)

type GetSettings struct {
	profiles ProfileStore
}

func NewGetSettings(profiles ProfileStore) *GetSettings {
	return &GetSettings{profiles: profiles}
}

func (uc *GetSettings) Execute(
	ctx context.Context, signedIn actor.Actor, organizationID string,
) (domain.Settings, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return domain.Settings{}, fmt.Errorf("check settings reader: %w", err)
	}
	if !isOwnOrganization(signedIn, organizationID) {
		return domain.Settings{}, apperror.NotFound()
	}
	settings, err := uc.profiles.GetSettings(ctx, organizationID)
	if err != nil {
		return domain.Settings{}, fmt.Errorf("get settings of %s: %w", organizationID, err)
	}
	return settings, nil
}

func isOwnOrganization(signedIn actor.Actor, organizationID string) bool {
	return signedIn.Membership.OrganizationID == organizationID
}
