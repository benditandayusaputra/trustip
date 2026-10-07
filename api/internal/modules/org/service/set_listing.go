package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
)

type SetListingCommand struct {
	OrganizationID string
	IsListed       bool
}

type ListingDeps struct {
	Profiles ProfileStore
	Listings ListingStore
	Audit    AuditTrail
	Tx       Transactor
	Clock    Clock
}

type SetListing struct {
	deps ListingDeps
}

func NewSetListing(deps ListingDeps) *SetListing {
	return &SetListing{deps: deps}
}

func (uc *SetListing) Execute(
	ctx context.Context, signedIn actor.Actor, cmd SetListingCommand,
) (domain.Settings, error) {
	if err := policy.Require(signedIn, policy.EditProfile); err != nil {
		return domain.Settings{}, fmt.Errorf("check listing editor: %w", err)
	}
	if !isOwnOrganization(signedIn, cmd.OrganizationID) {
		return domain.Settings{}, apperror.OutsideScope()
	}
	change := domain.ListingChange{OrganizationID: cmd.OrganizationID, IsListed: cmd.IsListed, At: uc.deps.Clock.Now()}
	var settings domain.Settings
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.checkReady(ctx, change); err != nil {
			return err
		}
		if err := uc.deps.Listings.SetListed(ctx, change); err != nil {
			return fmt.Errorf("set listed: %w", err)
		}
		if err := recordOrganizationEvent(ctx, uc.deps.Audit, organizationEvent{
			signedIn: signedIn, organizationID: cmd.OrganizationID, action: audit.ListingChanged,
			metadata: map[string]string{"is_listed": strconv.FormatBool(cmd.IsListed)}, at: change.At,
		}); err != nil {
			return err
		}
		reloaded, err := uc.deps.Profiles.GetSettings(ctx, cmd.OrganizationID)
		if err != nil {
			return fmt.Errorf("reload settings: %w", err)
		}
		settings = reloaded
		return nil
	})
	if err != nil {
		return domain.Settings{}, fmt.Errorf("set listing of %s: %w", cmd.OrganizationID, err)
	}
	return settings, nil
}

func (uc *SetListing) checkReady(ctx context.Context, change domain.ListingChange) error {
	if !change.IsListed {
		return nil
	}
	facts, err := uc.deps.Listings.GetListingFacts(ctx, change.OrganizationID)
	if err != nil {
		return fmt.Errorf("get listing facts: %w", err)
	}
	missing := domain.CompletenessOf(facts).MissingRequired()
	if len(missing) == 0 {
		return nil
	}
	codes := make([]string, 0, len(missing))
	for _, requirement := range missing {
		codes = append(codes, string(requirement))
	}
	return apperror.ProfileIncomplete(codes)
}
