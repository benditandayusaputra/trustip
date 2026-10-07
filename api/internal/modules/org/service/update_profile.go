package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
)

type UpdateProfileCommand struct {
	OrganizationID string
	Form           domain.ProfileUpdateForm
}

type ProfileDeps struct {
	Profiles ProfileStore
	Audit    AuditTrail
	Tx       Transactor
	Clock    Clock
}

type UpdateProfile struct {
	deps ProfileDeps
}

func NewUpdateProfile(deps ProfileDeps) *UpdateProfile {
	return &UpdateProfile{deps: deps}
}

func (uc *UpdateProfile) Execute(
	ctx context.Context, signedIn actor.Actor, cmd UpdateProfileCommand,
) (domain.Settings, error) {
	if err := policy.Require(signedIn, policy.EditProfile); err != nil {
		return domain.Settings{}, fmt.Errorf("check profile editor: %w", err)
	}
	if !isOwnOrganization(signedIn, cmd.OrganizationID) {
		return domain.Settings{}, apperror.OutsideScope()
	}
	update, err := domain.NewProfileUpdate(cmd.Form, signedIn.Membership.OrganizationType)
	if err != nil {
		return domain.Settings{}, fmt.Errorf("validate profile: %w", err)
	}
	change := domain.ProfileChange{OrganizationID: cmd.OrganizationID, Update: update, At: uc.deps.Clock.Now()}
	var settings domain.Settings
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.deps.Profiles.UpdateProfile(ctx, change); err != nil {
			return explainProfileFailure(err)
		}
		if err := recordOrganizationEvent(ctx, uc.deps.Audit, organizationEvent{
			signedIn: signedIn, organizationID: cmd.OrganizationID, action: audit.ProfileUpdated, at: change.At,
		}); err != nil {
			return err
		}
		settings, err = uc.deps.Profiles.GetSettings(ctx, cmd.OrganizationID)
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
