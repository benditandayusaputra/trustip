package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
)

type GetRunning struct {
	settlements SettlementReader
}

func NewGetRunning(deps Deps) *GetRunning {
	return &GetRunning{settlements: deps.Settlements}
}

func (uc *GetRunning) Execute(ctx context.Context, signedIn actor.Actor, partnershipID string) (domain.Running, error) {
	if err := policy.Require(signedIn, policy.ViewDocumentValues); err != nil {
		return domain.Running{}, fmt.Errorf("check running sales reader: %w", err)
	}
	running, err := uc.settlements.GetRunning(ctx, domain.RunningKey{
		OrganizationID: signedIn.Membership.OrganizationID, PartnershipID: partnershipID,
	})
	if err != nil {
		return domain.Running{}, fmt.Errorf("get running sales of %s: %w", partnershipID, err)
	}
	return running, nil
}
