package domain

import (
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

type CounterParams struct {
	RevisionID string
	Counterer  Decider
	Current    CurrentAgreement
	At         time.Time
}

type Counteroffer struct {
	ProposalID     string
	RevisionID     string
	RevisionNo     int
	PreviousNo     int
	AuthorOrgID    string
	AuthorUserID   string
	AwaitingOrgID  string
	ConsignorID    string
	Draft          RevisionDraft
	At             time.Time
	ExpiresAt      time.Time
	CountererName  string
	ProposalNumber string
}

func (p Proposal) Counter(draft RevisionDraft, params CounterParams) (Counteroffer, error) {
	if err := p.checkReceiver(params.Counterer, params.At); err != nil {
		return Counteroffer{}, err
	}
	draft, err := p.checkCounterItems(draft, params.Current)
	if err != nil {
		return Counteroffer{}, err
	}
	return Counteroffer{
		ProposalID: p.ID, RevisionID: params.RevisionID, RevisionNo: p.Revision.No + 1, PreviousNo: p.Revision.No,
		AuthorOrgID: params.Counterer.OrganizationID, AuthorUserID: params.Counterer.UserID,
		AwaitingOrgID: p.otherPartyOf(params.Counterer.OrganizationID), ConsignorID: p.Consignor.OrganizationID,
		Draft: draft, At: params.At, ExpiresAt: params.At.Add(OpenFor),
		CountererName: p.nameOf(params.Counterer.OrganizationID), ProposalNumber: p.Number,
	}, nil
}

func (p Proposal) checkCounterItems(draft RevisionDraft, current CurrentAgreement) (RevisionDraft, error) {
	if p.Kind == New {
		if problems := draft.onlyAdditions(); len(problems) > 0 {
			return RevisionDraft{}, apperror.Validation(problems...)
		}
		return draft, nil
	}
	items, problems := current.resolve(draft.Items)
	if len(items) == 0 && draft.Terms == current.Terms {
		problems = append(problems, apperror.InvalidField("items", apperror.FieldRequired))
	}
	if len(problems) > 0 {
		return RevisionDraft{}, apperror.Validation(problems...)
	}
	draft.Items = items
	return draft, nil
}

func (p Proposal) otherPartyOf(organizationID string) string {
	if p.Consignor.OrganizationID == organizationID {
		return p.Store.OrganizationID
	}
	return p.Consignor.OrganizationID
}

func (c Counteroffer) Notice() Notice {
	return Notice{
		OrganizationID: c.AwaitingOrgID, ActorUserID: c.AuthorUserID, Type: "PROPOSAL_COUNTERED",
		Title: "Ajuan ditawar", Body: fmt.Sprintf("%s menawar ajuan %s.", c.CountererName, c.ProposalNumber),
		Link: proposalLinkRoot + c.ProposalID, OwnersOnly: true,
	}
}
