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

type UpdateProfileCommand struct {
	OrganizationID string
	Form           domain.ProfileUpdateForm
}

type UpdateProfile struct {
	profiles ProfileStore
	tx       Transactor
	clock    Clock
}

func NewUpdateProfile(profiles ProfileStore, tx Transactor, clock Clock) *UpdateProfile {
	return &UpdateProfile{profiles: profiles, tx: tx, clock: clock}
}

func (uc *UpdateProfile) Execute(
	ctx context.Context, signedIn actor.Actor, cmd UpdateProfileCommand,
) (domain.Settings, error) {
	if err := policy.Require(signedIn, policy.EditProfile); err != nil {
		return domain.Settings{}, fmt.Errorf("check profile editor: %w", err)
	}
	if !isOwnOrganization(signedIn, cmd.OrganizationID) {
		return domain.Settings{}, apperror.NotFound()
	}
	update, err := domain.NewProfileUpdate(cmd.Form, signedIn.Membership.OrganizationType)
	if err != nil {
		return domain.Settings{}, fmt.Errorf("validate profile: %w", err)
	}
	change := domain.ProfileChange{OrganizationID: cmd.OrganizationID, Update: update, At: uc.clock.Now()}
	var settings domain.Settings
	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.profiles.UpdateProfile(ctx, change); err != nil {
			return explainProfileFailure(err)
		}
		settings, err = uc.profiles.GetSettings(ctx, cmd.OrganizationID)
		if err != nil {
			return fmt.Errorf("reload settings: %w", err)
		}
		return nil
	})
	if err != nil {
		return domain.Settings{}, fmt.Errorf("update profile of %s: %w", cmd.OrganizationID, err)
	}
	return settings, nil
}

func explainProfileFailure(err error) error {
	switch {
	case errors.Is(err, domain.ErrUnknownCity):
		return apperror.Validation(apperror.InvalidField("city_code", apperror.FieldUnknownChoice))
	case errors.Is(err, domain.ErrUnknownLogo):
		return apperror.Validation(apperror.InvalidField("logo_file_id", apperror.FieldUnknownChoice))
	}
	return fmt.Errorf("save profile: %w", err)
}
