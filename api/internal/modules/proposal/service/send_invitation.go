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

type SendInvitation struct {
	deps ProposalDeps
}

func NewSendInvitation(deps ProposalDeps) *SendInvitation {
	return &SendInvitation{deps: deps}
}

func (uc *SendInvitation) Execute(
	ctx context.Context, signedIn actor.Actor, input domain.RevisionInput,
) (domain.Proposal, error) {
	if err := policy.Require(signedIn, policy.NegotiateProposals); err != nil {
		return domain.Proposal{}, fmt.Errorf("check proposal sender: %w", err)
	}
	draft, err := domain.NewRevisionDraft(input)
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("validate invitation: %w", err)
	}
	ids, err := uc.newIDs()
	if err != nil {
		return domain.Proposal{}, err
	}
	sentAt := uc.deps.Clock.Now()
	var proposal domain.Proposal
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := consumeQuota(ctx, uc.deps.Direct, signedIn.Membership.OrganizationID, sentAt); err != nil {
			return err
		}
		invitation, err := uc.invite(ctx, draft, domain.InvitationParams{IDs: ids, Sender: domain.Sender{
			UserID: signedIn.UserID, OrganizationID: signedIn.Membership.OrganizationID,
			OrganizationType: signedIn.Membership.OrganizationType,
		}, At: sentAt})
		if err != nil {
			return err
		}
		if err := uc.deps.Proposals.Open(ctx, invitation.Opening()); err != nil {
			return explainCreateFailure(err)
		}
		proposal, err = finishProposalEvent(ctx, uc.deps, proposalEvent{
			signedIn: signedIn, proposalID: ids.ProposalID, action: audit.ProposalSent, at: sentAt,
		})
		return err
	})
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("send invitation: %w", err)
	}
	return proposal, nil
}

func (uc *SendInvitation) newIDs() (domain.InvitationIDs, error) {
	proposalID, proposalErr := uc.deps.IDs.NewID()
	revisionID, revisionErr := uc.deps.IDs.NewID()
	code, codeErr := uc.deps.Codes.NewCode()
	if err := errors.Join(proposalErr, revisionErr, codeErr); err != nil {
		return domain.InvitationIDs{}, fmt.Errorf("generate invitation ids: %w", err)
	}
	return domain.InvitationIDs{ProposalID: proposalID, RevisionID: revisionID, Code: code}, nil
}

func (uc *SendInvitation) invite(
	ctx context.Context, draft domain.RevisionDraft, params domain.InvitationParams,
) (domain.Invitation, error) {
	sequence, err := uc.deps.Proposals.NextNumber(ctx, domain.SequenceKey{
		OrganizationID: params.Sender.OrganizationID, Prefix: domain.NumberPrefix, Period: period.MonthKey(params.At),
	})
	if err != nil {
		return domain.Invitation{}, fmt.Errorf("next proposal number: %w", err)
	}
	params.Sequence = sequence
	invitation, err := draft.Invite(params)
	if err != nil {
		return domain.Invitation{}, fmt.Errorf("prepare invitation: %w", err)
	}
	return invitation, nil
}

func explainCreateFailure(err error) error {
	if errors.Is(err, domain.ErrUnknownProduct) {
		return apperror.Validation(apperror.InvalidField("items", apperror.FieldUnknownChoice))
	}
	return fmt.Errorf("create invitation: %w", err)
}
