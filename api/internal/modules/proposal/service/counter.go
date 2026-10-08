package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
)

type CounterCommand struct {
	ProposalID string
	Input      domain.RevisionInput
}

type CounterProposal struct {
	deps ProposalDeps
}

func NewCounterProposal(deps ProposalDeps) *CounterProposal {
	return &CounterProposal{deps: deps}
}

func (uc *CounterProposal) Execute(
	ctx context.Context, signedIn actor.Actor, cmd CounterCommand,
) (domain.Proposal, error) {
	if err := policy.Require(signedIn, policy.NegotiateProposals); err != nil {
		return domain.Proposal{}, fmt.Errorf("check counter author: %w", err)
	}
	draft, err := domain.NewRevisionDraft(cmd.Input)
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("validate counter: %w", err)
	}
	revisionID, err := uc.deps.IDs.NewID()
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("generate revision id: %w", err)
	}
	params := domain.CounterParams{
		RevisionID: revisionID, At: uc.deps.Clock.Now(),
		Counterer: domain.Decider{UserID: signedIn.UserID, OrganizationID: signedIn.Membership.OrganizationID},
	}
	var proposal domain.Proposal
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.counter(ctx, draft, params, cmd.ProposalID); err != nil {
			return err
		}
		proposal, err = finishProposalEvent(ctx, uc.deps, proposalEvent{
			signedIn: signedIn, proposalID: cmd.ProposalID, action: audit.ProposalCountered, at: params.At,
		})
		return err
	})
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("counter proposal %s: %w", cmd.ProposalID, err)
	}
	return proposal, nil
}

func (uc *CounterProposal) counter(
	ctx context.Context, draft domain.RevisionDraft, params domain.CounterParams, proposalID string,
) error {
	key := domain.ProposalKey{OrganizationID: params.Counterer.OrganizationID, ProposalID: proposalID}
	locked, err := uc.deps.Proposals.LockProposal(ctx, key)
	if err != nil {
		return fmt.Errorf("lock proposal: %w", err)
	}
	if params.Current, err = uc.currentAgreementOf(ctx, locked, params.Counterer); err != nil {
		return err
	}
	offer, err := locked.Counter(draft, params)
	if err != nil {
		return fmt.Errorf("prepare counter: %w", err)
	}
	err = uc.deps.Bargains.SaveCounter(ctx, offer)
	if errors.Is(err, domain.ErrAlreadyDecided) {
		return apperror.New(apperror.CodeProposalDecided)
	}
	if err != nil {
		return explainCreateFailure(err)
	}
	if err := uc.deps.Announcer.Announce(ctx, offer.Notice()); err != nil {
		return fmt.Errorf("announce counter: %w", err)
	}
	return nil
}

func (uc *CounterProposal) currentAgreementOf(
	ctx context.Context, proposal domain.Proposal, counterer domain.Decider,
) (domain.CurrentAgreement, error) {
	if proposal.Kind != domain.Amendment {
		return domain.CurrentAgreement{}, nil
	}
	current, err := uc.deps.Current.CurrentAgreement(ctx, domain.PartnershipKey{
		OrganizationID: counterer.OrganizationID, PartnershipID: proposal.PartnershipID,
	})
	if err != nil {
		return domain.CurrentAgreement{}, fmt.Errorf("read current agreement: %w", err)
	}
	return current, nil
}

type ListRevisions struct {
	deps ProposalDeps
}

func NewListRevisions(deps ProposalDeps) *ListRevisions {
	return &ListRevisions{deps: deps}
}

func (uc *ListRevisions) Execute(
	ctx context.Context, signedIn actor.Actor, proposalID string,
) ([]domain.RevisionRecord, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return nil, fmt.Errorf("check revision reader: %w", err)
	}
	proposal, err := reloadProposal(ctx, uc.deps.Proposals, signedIn, proposalID)
	if err != nil {
		return nil, err
	}
	revisions, err := uc.deps.Bargains.ListRevisions(ctx, domain.ProposalKey{
		OrganizationID: signedIn.Membership.OrganizationID, ProposalID: proposalID,
	})
	if err != nil {
		return nil, fmt.Errorf("list revisions of %s: %w", proposalID, err)
	}
	return proposal.History(revisions), nil
}
