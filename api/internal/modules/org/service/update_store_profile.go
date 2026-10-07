package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
)

type StoreProfileDeps struct {
	Profiles ProfileStore
	Stores   StoreProfileStore
	Tx       Transactor
	Clock    Clock
}

type UpdateStoreProfileCommand struct {
	OrganizationID string
	Form           domain.StoreProfileForm
}

type UpdateStoreProfile struct {
	deps StoreProfileDeps
}

func NewUpdateStoreProfile(deps StoreProfileDeps) *UpdateStoreProfile {
	return &UpdateStoreProfile{deps: deps}
}

func (uc *UpdateStoreProfile) Execute(
	ctx context.Context, signedIn actor.Actor, cmd UpdateStoreProfileCommand,
) (domain.Settings, error) {
	if err := policy.Require(signedIn, policy.EditProfile); err != nil {
		return domain.Settings{}, fmt.Errorf("check store profile editor: %w", err)
	}
	isStore := signedIn.Membership.OrganizationType == actor.Store
	if !isOwnOrganization(signedIn, cmd.OrganizationID) || !isStore {
		return domain.Settings{}, apperror.NotFound()
	}
	profile, err := domain.NewStoreProfile(cmd.Form)
	if err != nil {
		return domain.Settings{}, fmt.Errorf("validate store profile: %w", err)
	}
	change := domain.StoreProfileChange{OrganizationID: cmd.OrganizationID, Profile: profile, At: uc.deps.Clock.Now()}
	var settings domain.Settings
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.deps.Stores.UpdateStoreProfile(ctx, change); err != nil {
			return explainStoreProfileFailure(err)
		}
		settings, err = uc.deps.Profiles.GetSettings(ctx, cmd.OrganizationID)
		if err != nil {
			return fmt.Errorf("reload settings: %w", err)
		}
		return nil
	})
	if err != nil {
		return domain.Settings{}, fmt.Errorf("update store profile of %s: %w", cmd.OrganizationID, err)
	}
	return settings, nil
}

func explainStoreProfileFailure(err error) error {
	switch {
	case errors.Is(err, domain.ErrUnknownCategory):
		return apperror.Validation(apperror.InvalidField("accepted_category_ids", apperror.FieldUnknownChoice))
	case errors.Is(err, domain.ErrUnknownPhoto):
		return apperror.Validation(apperror.InvalidField("photo_file_ids", apperror.FieldUnknownChoice))
	}
	return fmt.Errorf("save store profile: %w", err)
}
