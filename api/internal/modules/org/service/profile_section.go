package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
)

type SectionDeps struct {
	Profiles ProfileStore
	Sections ProfileSectionStore
	Audit    AuditTrail
	Tx       Transactor
	Clock    Clock
}

type sectionTarget struct {
	signedIn       actor.Actor
	organizationID string
	organization   actor.OrganizationType
	action         audit.Action
}

func (t sectionTarget) isReachable() bool {
	return isOwnOrganization(t.signedIn, t.organizationID) && t.signedIn.Membership.OrganizationType == t.organization
}

func saveSection(
	ctx context.Context, deps SectionDeps, target sectionTarget, write func(context.Context) error,
) (domain.Settings, error) {
	var settings domain.Settings
	err := deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := write(ctx); err != nil {
			return explainSectionFailure(err)
		}
		if err := recordOrganizationEvent(ctx, deps.Audit, organizationEvent{
			signedIn: target.signedIn, organizationID: target.organizationID, action: target.action,
			at: deps.Clock.Now(),
		}); err != nil {
			return err
		}
		reloaded, err := deps.Profiles.GetSettings(ctx, target.organizationID)
		if err != nil {
			return fmt.Errorf("reload settings: %w", err)
		}
		settings = reloaded
		return nil
	})
	if err != nil {
		return domain.Settings{}, fmt.Errorf("save section of %s: %w", target.organizationID, err)
	}
	return settings, nil
}

func sectionFields() map[error]string {
	return map[error]string{
		domain.ErrUnknownCategory: "accepted_category_ids",
		domain.ErrUnknownPhoto:    "photo_file_ids",
		domain.ErrUnknownQR:       "payment_qr_file_id",
	}
}

func explainSectionFailure(err error) error {
	for cause, field := range sectionFields() {
		if errors.Is(err, cause) {
			return apperror.Validation(apperror.InvalidField(field, apperror.FieldUnknownChoice))
		}
	}
	return fmt.Errorf("write section: %w", err)
}
