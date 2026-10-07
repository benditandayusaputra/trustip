package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
)

type AmendmentCommand struct {
	PartnershipID string
	Input         domain.RevisionInput
}

type SendAmendment struct {
	deps ProposalDeps
}

func NewSendAmendment(deps ProposalDeps) *SendAmendment {
	return &SendAmendment{deps: deps}
}

func (uc *SendAmendment) Execute(
	ctx context.Context, signedIn actor.Actor, cmd AmendmentCommand,
) (domain.Proposal, error) {
	if err := policy.Require(signedIn, policy.AmendAgreements); err != nil {
		return domain.Proposal{}, fmt.Errorf("check amendment sender: %w", err)
	}
	draft, err := domain.NewRevisionDraft(cmd.Input)
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("validate amendment: %w", err)
	}
	proposalID, proposalErr := uc.deps.IDs.NewID()
	revisionID, revisionErr := uc.deps.IDs.NewID()
	if err := errors.Join(proposalErr, revisionErr); err != nil {
		return domain.Proposal{}, fmt.Errorf("generate amendment ids: %w", err)
	}
	params := domain.AmendmentParams{
		IDs: domain.InvitationIDs{ProposalID: proposalID, RevisionID: revisionID}, At: uc.deps.Clock.Now(),
		Sender: domain.Sender{
			UserID: signedIn.UserID, OrganizationID: signedIn.Membership.OrganizationID,
			OrganizationType: signedIn.Membership.OrganizationType,
		},
	}
	var proposal domain.Proposal
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		opening, err := uc.open(ctx, draft, params, cmd.PartnershipID)
		if err != nil {
			return err
		}
		if err := uc.deps.Announcer.Announce(ctx, opening.ReceivedNotice()); err != nil {
			return fmt.Errorf("announce amendment: %w", err)
		}
		proposal, err = finishProposalEvent(ctx, uc.deps, proposalEvent{
			signedIn: signedIn, proposalID: proposalID, action: audit.ProposalSent, at: params.At,
		})
		return err
	})
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("send amendment for %s: %w", cmd.PartnershipID, err)
	}
	return proposal, nil
}

func (uc *SendAmendment) open(
	ctx context.Context, draft domain.RevisionDraft, params domain.AmendmentParams, partnershipID string,
) (domain.Opening, error) {
	current, err := uc.deps.Current.CurrentAgreement(ctx, domain.PartnershipKey{
		OrganizationID: params.Sender.OrganizationID, PartnershipID: partnershipID,
	})
	if err != nil {
		return domain.Opening{}, fmt.Errorf("read current agreement: %w", err)
	}
	params.Current = current
	params.Sequence, err = uc.deps.Proposals.NextNumber(ctx, domain.SequenceKey{
		OrganizationID: params.Sender.OrganizationID, Prefix: domain.NumberPrefix, Period: period.MonthKey(params.At),
	})
	if err != nil {
		return domain.Opening{}, fmt.Errorf("next proposal number: %w", err)
	}
	opening, err := draft.Amend(params)
	if err != nil {
		return domain.Opening{}, fmt.Errorf("prepare amendment: %w", err)
	}
	err = uc.deps.Proposals.Open(ctx, opening)
	if errors.Is(err, domain.ErrAlreadyOpen) {
		return domain.Opening{}, apperror.New(apperror.CodeProposalAlreadyOpen).
			WithParam(organizationParam, current.CounterpartName(params.Sender.OrganizationID))
	}
	if err != nil {
		return domain.Opening{}, explainCreateFailure(err)
	}
	return opening, nil
}
