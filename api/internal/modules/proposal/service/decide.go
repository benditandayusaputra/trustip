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

type decideFunc func(ctx context.Context, proposal domain.Proposal, decider domain.Decider) (domain.Decision, error)

type decisionFlow struct {
	deps   ProposalDeps
	action audit.Action
}

func (f decisionFlow) run(
	ctx context.Context, signedIn actor.Actor, proposalID string, decide decideFunc,
) (domain.Proposal, error) {
	decider := domain.Decider{UserID: signedIn.UserID, OrganizationID: signedIn.Membership.OrganizationID}
	var proposal domain.Proposal
	err := f.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		locked, err := f.deps.Proposals.LockProposal(ctx, domain.ProposalKey{
			OrganizationID: decider.OrganizationID, ProposalID: proposalID,
		})
		if err != nil {
			return fmt.Errorf("lock proposal: %w", err)
		}
		decision, err := decide(ctx, locked, decider)
		if err != nil {
			return err
		}
		if err := f.record(ctx, locked, decision); err != nil {
			return err
		}
		proposal, err = finishProposalEvent(ctx, f.deps, proposalEvent{
			signedIn: signedIn, proposalID: proposalID, action: f.action, at: decision.At,
		})
		return err
	})
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("decide proposal %s: %w", proposalID, err)
	}
	return proposal, nil
}

func (f decisionFlow) record(ctx context.Context, proposal domain.Proposal, decision domain.Decision) error {
	err := f.deps.Proposals.SaveDecision(ctx, decision)
	if errors.Is(err, domain.ErrAlreadyDecided) {
		return apperror.New(apperror.CodeProposalDecided)
	}
	if err != nil {
		return fmt.Errorf("save decision: %w", err)
	}
	return f.announce(ctx, proposal, decision)
}

func (f decisionFlow) announce(ctx context.Context, proposal domain.Proposal, decision domain.Decision) error {
	if !decision.IsAnnounced() {
		return nil
	}
	notices := []domain.Notice{proposal.NoticeOf(decision)}
	if decision.AgreementVersion > 0 {
		notices = append(notices, proposal.AgreementNotices(decision, decision.AgreementVersion)...)
	}
	for _, notice := range notices {
		if err := f.deps.Announcer.Announce(ctx, notice); err != nil {
			return fmt.Errorf("announce %s: %w", notice.Type, err)
		}
	}
	return nil
}

type AcceptProposal struct {
	flow decisionFlow
}

func NewAcceptProposal(deps ProposalDeps) *AcceptProposal {
	return &AcceptProposal{flow: decisionFlow{deps: deps, action: audit.ProposalAccepted}}
}

func (uc *AcceptProposal) Execute(
	ctx context.Context, signedIn actor.Actor, proposalID string,
) (domain.Proposal, error) {
	if err := policy.Require(signedIn, policy.NegotiateProposals); err != nil {
		return domain.Proposal{}, fmt.Errorf("check proposal decider: %w", err)
	}
	return uc.flow.run(ctx, signedIn, proposalID, uc.accept)
}

func (uc *AcceptProposal) accept(
	ctx context.Context, proposal domain.Proposal, decider domain.Decider,
) (domain.Decision, error) {
	decision, founding, err := proposal.Accept(decider, uc.flow.deps.Clock.Now())
	if err != nil {
		return domain.Decision{}, fmt.Errorf("accept: %w", err)
	}
	if proposal.Kind == domain.Amendment {
		return uc.amend(ctx, decision, founding)
	}
	partnershipID, err := uc.flow.deps.Keeper.Found(ctx, founding)
	if errors.Is(err, domain.ErrAlreadyPartners) {
		return domain.Decision{}, apperror.New(apperror.CodeAlreadyPartners).
			WithParam(organizationParam, proposal.CounterpartOf(decider.OrganizationID))
	}
	if err != nil {
		return domain.Decision{}, fmt.Errorf("found agreement: %w", err)
	}
	decision.PartnershipID = partnershipID
	return decision, nil
}

func (uc *AcceptProposal) amend(
	ctx context.Context, decision domain.Decision, founding domain.Founding,
) (domain.Decision, error) {
	version, err := uc.flow.deps.Keeper.Amend(ctx, founding)
	if err != nil {
		return domain.Decision{}, fmt.Errorf("amend agreement: %w", err)
	}
	decision.PartnershipID = founding.PartnershipID
	decision.AgreementVersion = version
	return decision, nil
}

type DeclineCommand struct {
	ProposalID string
	Reason     string
}

type DeclineProposal struct {
	flow decisionFlow
}

func NewDeclineProposal(deps ProposalDeps) *DeclineProposal {
	return &DeclineProposal{flow: decisionFlow{deps: deps, action: audit.ProposalDeclined}}
}

func (uc *DeclineProposal) Execute(
	ctx context.Context, signedIn actor.Actor, cmd DeclineCommand,
) (domain.Proposal, error) {
	if err := policy.Require(signedIn, policy.NegotiateProposals); err != nil {
		return domain.Proposal{}, fmt.Errorf("check proposal decider: %w", err)
	}
	decline := func(_ context.Context, proposal domain.Proposal, decider domain.Decider) (domain.Decision, error) {
		return proposal.Decline(decider, cmd.Reason, uc.flow.deps.Clock.Now())
	}
	return uc.flow.run(ctx, signedIn, cmd.ProposalID, decline)
}

type WithdrawProposal struct {
	flow decisionFlow
}

func NewWithdrawProposal(deps ProposalDeps) *WithdrawProposal {
	return &WithdrawProposal{flow: decisionFlow{deps: deps, action: audit.ProposalWithdrawn}}
}

func (uc *WithdrawProposal) Execute(
	ctx context.Context, signedIn actor.Actor, proposalID string,
) (domain.Proposal, error) {
	if err := policy.Require(signedIn, policy.NegotiateProposals); err != nil {
		return domain.Proposal{}, fmt.Errorf("check proposal sender: %w", err)
	}
	withdraw := func(_ context.Context, proposal domain.Proposal, decider domain.Decider) (domain.Decision, error) {
		return proposal.Withdraw(decider, uc.flow.deps.Clock.Now())
	}
	return uc.flow.run(ctx, signedIn, proposalID, withdraw)
}
