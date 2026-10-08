package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
)

type ProposalDeps struct {
	Proposals ProposalStore
	Direct    DirectProposals
	Audit     AuditTrail
	Tx        Transactor
	Clock     Clock
	IDs       IDGenerator
	Codes     CodeGenerator
	Keeper    AgreementKeeper
	Current   CurrentAgreementReader
	Announcer Announcer
}

type proposalEvent struct {
	signedIn   actor.Actor
	proposalID string
	action     audit.Action
	at         time.Time
}

func finishProposalEvent(ctx context.Context, deps ProposalDeps, event proposalEvent) (domain.Proposal, error) {
	err := deps.Audit.Record(ctx, audit.Entry{
		OrganizationID: event.signedIn.Membership.OrganizationID, ActorID: event.signedIn.UserID,
		Action: event.action, EntityType: audit.ProposalEntity, EntityID: event.proposalID, At: event.at,
	})
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("audit %s: %w", event.action, err)
	}
	return reloadProposal(ctx, deps.Proposals, event.signedIn, event.proposalID)
}

func reloadProposal(
	ctx context.Context, reader ProposalReader, signedIn actor.Actor, proposalID string,
) (domain.Proposal, error) {
	proposal, err := reader.GetProposal(ctx, domain.ProposalKey{
		OrganizationID: signedIn.Membership.OrganizationID, ProposalID: proposalID,
	})
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("load proposal %s: %w", proposalID, err)
	}
	return proposal, nil
}

func consumeQuota(ctx context.Context, direct DirectProposals, organizationID string, at time.Time) error {
	err := direct.ConsumeQuota(ctx, domain.QuotaKeyOf(organizationID, at))
	if errors.Is(err, domain.ErrQuotaExhausted) {
		return apperror.New(apperror.CodeProposalLimitReached)
	}
	if err != nil {
		return fmt.Errorf("consume proposal quota: %w", err)
	}
	return nil
}
