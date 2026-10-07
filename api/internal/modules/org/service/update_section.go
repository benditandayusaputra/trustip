package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
)

type SectionCommand[F any] struct {
	OrganizationID string
	Form           F
}

type (
	UpdateStoreProfileCommand     = SectionCommand[domain.StoreProfileForm]
	UpdateConsignorProfileCommand = SectionCommand[domain.ConsignorProfileForm]
	UpdatePaymentInfoCommand      = SectionCommand[domain.PaymentInfoForm]
)

type sectionWrite func(ctx context.Context) error

type sectionPlan[F any] func(cmd SectionCommand[F], at time.Time) (sectionWrite, error)

type UpdateSection[F any] struct {
	deps         SectionDeps
	organization actor.OrganizationType
	plan         sectionPlan[F]
}

func (uc *UpdateSection[F]) Execute(
	ctx context.Context, signedIn actor.Actor, cmd SectionCommand[F],
) (domain.Settings, error) {
	if err := policy.Require(signedIn, policy.EditProfile); err != nil {
		return domain.Settings{}, fmt.Errorf("check section editor: %w", err)
	}
	target := sectionTarget{signedIn: signedIn, organizationID: cmd.OrganizationID, organization: uc.organization}
	if !target.isReachable() {
		return domain.Settings{}, apperror.NotFound()
	}
	write, err := uc.plan(cmd, uc.deps.Clock.Now())
	if err != nil {
		return domain.Settings{}, err
	}
	return saveSection(ctx, uc.deps, target, write)
}

func NewUpdateStoreProfile(deps SectionDeps) *UpdateSection[domain.StoreProfileForm] {
	plan := func(cmd UpdateStoreProfileCommand, at time.Time) (sectionWrite, error) {
		profile, err := domain.NewStoreProfile(cmd.Form)
		if err != nil {
			return nil, fmt.Errorf("validate store profile: %w", err)
		}
		change := domain.StoreProfileChange{OrganizationID: cmd.OrganizationID, Profile: profile, At: at}
		return func(ctx context.Context) error { return deps.Sections.UpdateStoreProfile(ctx, change) }, nil
	}
	return &UpdateSection[domain.StoreProfileForm]{deps: deps, organization: actor.Store, plan: plan}
}

func NewUpdateConsignorProfile(deps SectionDeps) *UpdateSection[domain.ConsignorProfileForm] {
	plan := func(cmd UpdateConsignorProfileCommand, at time.Time) (sectionWrite, error) {
		profile, err := domain.NewConsignorProfile(cmd.Form)
		if err != nil {
			return nil, fmt.Errorf("validate consignor profile: %w", err)
		}
		change := domain.ConsignorProfileChange{OrganizationID: cmd.OrganizationID, Profile: profile, At: at}
		return func(ctx context.Context) error { return deps.Sections.UpdateConsignorProfile(ctx, change) }, nil
	}
	return &UpdateSection[domain.ConsignorProfileForm]{deps: deps, organization: actor.Consignor, plan: plan}
}

func NewUpdatePaymentInfo(deps SectionDeps) *UpdateSection[domain.PaymentInfoForm] {
	plan := func(cmd UpdatePaymentInfoCommand, at time.Time) (sectionWrite, error) {
		payment, err := domain.NewPaymentInfo(cmd.Form)
		if err != nil {
			return nil, fmt.Errorf("validate payment info: %w", err)
		}
		change := domain.PaymentInfoChange{OrganizationID: cmd.OrganizationID, Payment: payment, At: at}
		return func(ctx context.Context) error { return deps.Sections.UpdatePaymentInfo(ctx, change) }, nil
	}
	return &UpdateSection[domain.PaymentInfoForm]{deps: deps, organization: actor.Consignor, plan: plan}
}
