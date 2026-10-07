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

const maxSlugAttempts = 3

type CreateDeps struct {
	Founder OrganizationFounder
	Audit   AuditTrail
	Tx      Transactor
	Clock   Clock
	IDs     IDGenerator
}

type CreateOrganization struct {
	deps CreateDeps
}

func NewCreateOrganization(deps CreateDeps) *CreateOrganization {
	return &CreateOrganization{deps: deps}
}

func (uc *CreateOrganization) Execute(
	ctx context.Context, signedIn actor.Actor, form domain.OnboardingForm,
) (domain.Organization, error) {
	if err := policy.RequireNewcomer(signedIn); err != nil {
		return domain.Organization{}, fmt.Errorf("check newcomer %s: %w", signedIn.UserID, err)
	}
	onboarding, err := domain.NewOnboarding(form)
	if err != nil {
		return domain.Organization{}, fmt.Errorf("validate onboarding: %w", err)
	}
	var lastErr error
	for range maxSlugAttempts {
		organization, err := uc.attempt(ctx, signedIn, onboarding)
		if !errors.Is(err, domain.ErrSlugTaken) {
			return organization, explainFoundingFailure(err)
		}
		lastErr = err
	}
	return domain.Organization{}, fmt.Errorf("found organization for %s: %w", signedIn.UserID, lastErr)
}

func (uc *CreateOrganization) attempt(
	ctx context.Context, founder actor.Actor, onboarding domain.Onboarding,
) (domain.Organization, error) {
	ids, err := uc.newIDs(founder.UserID)
	if err != nil {
		return domain.Organization{}, err
	}
	var organization domain.Organization
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		base := domain.SlugFrom(onboarding.Name)
		taken, err := uc.deps.Founder.TakenSlugs(ctx, base)
		if err != nil {
			return fmt.Errorf("list taken slugs for %s: %w", base, err)
		}
		founding := onboarding.Found(ids, domain.NextFreeSlug(base, taken), uc.deps.Clock.Now())
		cityName, err := uc.deps.Founder.CreateOrganization(ctx, founding)
		if err != nil {
			return fmt.Errorf("create organization %s: %w", ids.OrganizationID, err)
		}
		organization = founding.Organization
		organization.CityName = cityName
		return recordOrganizationEvent(ctx, uc.deps.Audit, organizationEvent{
			signedIn: founder, organizationID: ids.OrganizationID, action: audit.OrganizationCreated,
			metadata: map[string]string{"type": string(onboarding.Type)}, at: founding.Organization.CreatedAt,
		})
	})
	if err != nil {
		return domain.Organization{}, fmt.Errorf("found organization %s: %w", ids.OrganizationID, err)
	}
	return organization, nil
}

func (uc *CreateOrganization) newIDs(founderID string) (domain.FoundingIDs, error) {
	organizationID, err := uc.deps.IDs.NewID()
	if err != nil {
		return domain.FoundingIDs{}, fmt.Errorf("generate organization id: %w", err)
	}
	membershipID, err := uc.deps.IDs.NewID()
	if err != nil {
		return domain.FoundingIDs{}, fmt.Errorf("generate membership id: %w", err)
	}
	return domain.FoundingIDs{OrganizationID: organizationID, MembershipID: membershipID, FounderID: founderID}, nil
}

func explainFoundingFailure(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrUnknownCity):
		return apperror.Validation(apperror.InvalidField("city_code", apperror.FieldUnknownChoice))
	case errors.Is(err, domain.ErrAlreadyMember):
		return apperror.New(apperror.CodeOrganizationExists).WithCause(err)
	}
	return err
}
