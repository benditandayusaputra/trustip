package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
)

type ListingCommand struct {
	PartnershipID string
	ShowOnPublic  bool
}

type SetShownOnPublic struct {
	deps EndDeps
}

func NewSetShownOnPublic(deps EndDeps) *SetShownOnPublic {
	return &SetShownOnPublic{deps: deps}
}

func (uc *SetShownOnPublic) Execute(
	ctx context.Context, signedIn actor.Actor, cmd ListingCommand,
) (domain.Partnership, error) {
	if err := policy.Require(signedIn, policy.EditProfile); err != nil {
		return domain.Partnership{}, fmt.Errorf("check listing editor: %w", err)
	}
	if signedIn.Membership.OrganizationType != actor.Store {
		return domain.Partnership{}, apperror.New(apperror.CodeForbidden)
	}
	err := uc.deps.Partnerships.SetShownOnPublic(ctx, domain.ListingChange{
		StoreID: signedIn.Membership.OrganizationID, PartnershipID: cmd.PartnershipID,
		ShowOnPublic: cmd.ShowOnPublic, At: uc.deps.Clock.Now(),
	})
	if err != nil {
		return domain.Partnership{}, fmt.Errorf("set public listing of %s: %w", cmd.PartnershipID, err)
	}
	partnership, err := uc.deps.Reader.GetPartnership(ctx, keyOf(signedIn, cmd.PartnershipID))
	if err != nil {
		return domain.Partnership{}, fmt.Errorf("read partnership %s: %w", cmd.PartnershipID, err)
	}
	return partnership, nil
}
