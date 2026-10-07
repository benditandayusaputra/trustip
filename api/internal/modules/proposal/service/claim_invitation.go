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

const organizationParam = "organisasi"

type ViewInvitation struct {
	deps ProposalDeps
}

func NewViewInvitation(deps ProposalDeps) *ViewInvitation {
	return &ViewInvitation{deps: deps}
}

func (uc *ViewInvitation) Execute(
	ctx context.Context, signedIn actor.Actor, code string,
) (domain.InvitationView, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return domain.InvitationView{}, fmt.Errorf("check invitation reader: %w", err)
	}
	view, err := findInvitation(ctx, uc.deps.Proposals, code)
	if err != nil {
		return domain.InvitationView{}, err
	}
	if _, err := view.ClaimBy(signedIn.Membership, uc.deps.Clock.Now()); err != nil {
		return domain.InvitationView{}, fmt.Errorf("check invitation %s: %w", view.ProposalID, err)
	}
	return view, nil
}

func findInvitation(ctx context.Context, reader ProposalReader, code string) (domain.InvitationView, error) {
	if !domain.IsInviteCode(code) {
		return domain.InvitationView{}, apperror.NotFound()
	}
	view, err := reader.GetInvitation(ctx, code)
	if err != nil {
		return domain.InvitationView{}, fmt.Errorf("get invitation: %w", err)
	}
	return view, nil
}

type ClaimInvitation struct {
	deps ProposalDeps
}

func NewClaimInvitation(deps ProposalDeps) *ClaimInvitation {
	return &ClaimInvitation{deps: deps}
}

func (uc *ClaimInvitation) Execute(ctx context.Context, signedIn actor.Actor, code string) (domain.Proposal, error) {
	if err := policy.Require(signedIn, policy.NegotiateProposals); err != nil {
		return domain.Proposal{}, fmt.Errorf("check invitation claimer: %w", err)
	}
	var proposal domain.Proposal
	err := uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		view, err := findInvitation(ctx, uc.deps.Proposals, code)
		if err != nil {
			return err
		}
		claim, err := view.ClaimBy(signedIn.Membership, uc.deps.Clock.Now())
		if err != nil {
			return fmt.Errorf("claim invitation %s: %w", view.ProposalID, err)
		}
		if claim.IsRepeat {
			proposal, err = reloadProposal(ctx, uc.deps.Proposals, signedIn, claim.ProposalID)
			return err
		}
		if err := uc.bind(ctx, claim, view.Sender.Name); err != nil {
			return err
		}
		proposal, err = finishProposalEvent(ctx, uc.deps, proposalEvent{
			signedIn: signedIn, proposalID: claim.ProposalID, action: audit.InvitationClaimed, at: claim.At,
		})
		return err
	})
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("claim invitation: %w", err)
	}
	return proposal, nil
}

func (uc *ClaimInvitation) bind(ctx context.Context, claim domain.Claim, senderName string) error {
	isPartner, err := uc.deps.Proposals.HasActivePartnership(ctx, claim.Pair())
	if err != nil {
		return fmt.Errorf("check partnership: %w", err)
	}
	if isPartner {
		return apperror.New(apperror.CodeAlreadyPartners).WithParam(organizationParam, senderName)
	}
	err = uc.deps.Proposals.ClaimInvitation(ctx, claim)
	switch {
	case errors.Is(err, domain.ErrAlreadyOpen):
		return apperror.New(apperror.CodeProposalAlreadyOpen).WithParam(organizationParam, senderName)
	case errors.Is(err, domain.ErrClaimRaced):
		return apperror.OutsideScope()
	case err != nil:
		return fmt.Errorf("bind claim: %w", err)
	}
	return nil
}

type GetProposal struct {
	reader ProposalReader
}

func NewGetProposal(reader ProposalReader) *GetProposal {
	return &GetProposal{reader: reader}
}

func (uc *GetProposal) Execute(ctx context.Context, signedIn actor.Actor, proposalID string) (domain.Proposal, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return domain.Proposal{}, fmt.Errorf("check proposal reader: %w", err)
	}
	return reloadProposal(ctx, uc.reader, signedIn, proposalID)
}
