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

type DirectCommand struct {
	TargetID string
	Input    domain.RevisionInput
}

type SendProposal struct {
	deps ProposalDeps
}

func NewSendProposal(deps ProposalDeps) *SendProposal {
	return &SendProposal{deps: deps}
}

func (uc *SendProposal) Execute(ctx context.Context, signedIn actor.Actor, cmd DirectCommand) (domain.Proposal, error) {
	if err := policy.Require(signedIn, policy.NegotiateProposals); err != nil {
		return domain.Proposal{}, fmt.Errorf("check proposal sender: %w", err)
	}
	draft, err := domain.NewRevisionDraft(cmd.Input)
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("validate proposal: %w", err)
	}
	proposalID, proposalErr := uc.deps.IDs.NewID()
	revisionID, revisionErr := uc.deps.IDs.NewID()
	if err := errors.Join(proposalErr, revisionErr); err != nil {
		return domain.Proposal{}, fmt.Errorf("generate proposal ids: %w", err)
	}
	params := domain.DirectParams{
		IDs: domain.InvitationIDs{ProposalID: proposalID, RevisionID: revisionID}, At: uc.deps.Clock.Now(),
		Sender: domain.Sender{
			UserID: signedIn.UserID, OrganizationID: signedIn.Membership.OrganizationID,
			OrganizationType: signedIn.Membership.OrganizationType,
		},
	}
	var proposal domain.Proposal
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.open(ctx, draft, params, cmd.TargetID); err != nil {
			return err
		}
		proposal, err = finishProposalEvent(ctx, uc.deps, proposalEvent{
			signedIn: signedIn, proposalID: proposalID, action: audit.ProposalSent, at: params.At,
		})
		return err
	})
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("send proposal to %s: %w", cmd.TargetID, err)
	}
	return proposal, nil
}

func (uc *SendProposal) open(
	ctx context.Context, draft domain.RevisionDraft, params domain.DirectParams, targetID string,
) error {
	params, err := uc.withParties(ctx, params, targetID)
	if err != nil {
		return err
	}
	opening, err := draft.Propose(params)
	if err != nil {
		return fmt.Errorf("prepare proposal: %w", err)
	}
	if err := consumeQuota(ctx, uc.deps.Direct, params.Sender.OrganizationID, params.At); err != nil {
		return err
	}
	err = uc.deps.Proposals.Open(ctx, opening)
	if errors.Is(err, domain.ErrAlreadyOpen) {
		return apperror.New(apperror.CodeProposalAlreadyOpen).WithParam(organizationParam, params.Target.Name)
	}
	if err != nil {
		return explainCreateFailure(err)
	}
	if err := uc.deps.Announcer.Announce(ctx, opening.ReceivedNotice()); err != nil {
		return fmt.Errorf("announce proposal: %w", err)
	}
	return nil
}

func (uc *SendProposal) withParties(
	ctx context.Context, params domain.DirectParams, targetID string,
) (domain.DirectParams, error) {
	parties, err := uc.deps.Direct.GetDirectParties(ctx, domain.DirectPair{
		SenderID: params.Sender.OrganizationID, TargetID: targetID,
	})
	if err != nil {
		return domain.DirectParams{}, fmt.Errorf("read proposal parties: %w", err)
	}
	if parties.ArePartners {
		return domain.DirectParams{}, apperror.New(apperror.CodeAlreadyPartners).
			WithParam(organizationParam, parties.Target.Name)
	}
	params.SenderName, params.Target, params.LastDeclinedAt = parties.SenderName, parties.Target, parties.LastDeclinedAt
	params.Sequence, err = uc.deps.Proposals.NextNumber(ctx, domain.SequenceKey{
		OrganizationID: params.Sender.OrganizationID, Prefix: domain.NumberPrefix, Period: period.MonthKey(params.At),
	})
	if err != nil {
		return domain.DirectParams{}, fmt.Errorf("next proposal number: %w", err)
	}
	return params, nil
}
