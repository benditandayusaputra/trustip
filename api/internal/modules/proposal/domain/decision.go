package domain

import (
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

const (
	MaxReasonLength   = 500
	unknownInvitee    = "mitra yang diundang"
	organizationParam = "organisasi"
	proposalLinkRoot  = "/app/ajuan/"
)

type Decider struct {
	UserID         string
	OrganizationID string
}

type Decision struct {
	ProposalID       string
	Status           Status
	DecidedBy        string
	DeciderOrgID     string
	Reason           string
	RevisionNo       int
	PartnershipID    string
	AgreementVersion int
	At               time.Time
}

type Founding struct {
	Kind           Kind
	PartnershipID  string
	ProposalID     string
	ProposalNumber string
	ConsignorID    string
	StoreID        string
	Terms          terms.General
	Items          []Item
	ProposedBy     string
	AcceptedBy     string
	AcceptorOrgID  string
	At             time.Time
}

func (p Proposal) Accept(decider Decider, decidedAt time.Time) (Decision, Founding, error) {
	if err := p.checkReceiver(decider, decidedAt); err != nil {
		return Decision{}, Founding{}, err
	}
	founding := Founding{
		Kind: p.Kind, PartnershipID: p.PartnershipID,
		ProposalID: p.ID, ProposalNumber: p.Number, ConsignorID: p.Consignor.OrganizationID,
		StoreID: p.Store.OrganizationID, Terms: p.Revision.Terms, Items: p.Revision.Items,
		ProposedBy: p.Revision.AuthorUserID, AcceptedBy: decider.UserID, AcceptorOrgID: decider.OrganizationID,
		At: decidedAt,
	}
	return p.decide(decider, Accepted, decidedAt), founding, nil
}

func (p Proposal) Decline(decider Decider, reason string, decidedAt time.Time) (Decision, error) {
	if err := p.checkReceiver(decider, decidedAt); err != nil {
		return Decision{}, err
	}
	if problems := field.Collect(findReasonProblem(reason)); len(problems) > 0 {
		return Decision{}, apperror.Validation(problems...)
	}
	decision := p.decide(decider, Declined, decidedAt)
	decision.Reason = reason
	return decision, nil
}

func (p Proposal) Withdraw(decider Decider, decidedAt time.Time) (Decision, error) {
	if err := p.checkOpen(decidedAt); err != nil {
		return Decision{}, err
	}
	if decider.OrganizationID != p.Revision.AuthorOrgID {
		return Decision{}, apperror.New(apperror.CodeForbidden)
	}
	return p.decide(decider, Withdrawn, decidedAt), nil
}

func (p Proposal) decide(decider Decider, status Status, decidedAt time.Time) Decision {
	return Decision{
		ProposalID: p.ID, Status: status, DecidedBy: decider.UserID, DeciderOrgID: decider.OrganizationID,
		RevisionNo: p.Revision.No, At: decidedAt,
	}
}

func (p Proposal) checkReceiver(decider Decider, decidedAt time.Time) error {
	if err := p.checkOpen(decidedAt); err != nil {
		return err
	}
	if decider.OrganizationID != p.AwaitingOrgID {
		return apperror.New(apperror.CodeNotYourTurn).WithParam(organizationParam, p.awaitingName())
	}
	return nil
}

func (p Proposal) checkOpen(decidedAt time.Time) error {
	if p.Status != Open {
		return apperror.New(apperror.CodeProposalDecided)
	}
	if !decidedAt.Before(p.ExpiresAt) {
		return apperror.New(apperror.CodeProposalExpired)
	}
	return nil
}

func (p Proposal) awaitingName() string {
	if name := p.nameOf(p.AwaitingOrgID); name != "" {
		return name
	}
	return unknownInvitee
}

func (p Proposal) nameOf(organizationID string) string {
	for _, party := range []Party{p.Consignor, p.Store} {
		if organizationID != "" && party.OrganizationID == organizationID {
			return party.Name
		}
	}
	return ""
}

func findReasonProblem(reason string) field.Problem {
	if reason == "" {
		return field.Invalid("reason", apperror.FieldRequired)
	}
	return field.CheckLength("reason", reason, field.Length{Max: MaxReasonLength})
}

type Notice struct {
	OrganizationID string
	ActorUserID    string
	Type           string
	Title          string
	Body           string
	Link           string
	OwnersOnly     bool
}

func noticeTexts() map[Status][2]string {
	return map[Status][2]string{
		Accepted: {"Ajuan diterima", "%s menerima ajuan %s."},
		Declined: {"Ajuan ditolak", "%s menolak ajuan %s."},
	}
}

func (p Proposal) NoticeOf(decision Decision) Notice {
	texts := noticeTexts()[decision.Status]
	return Notice{
		OrganizationID: p.Revision.AuthorOrgID, ActorUserID: decision.DecidedBy,
		Type: "PROPOSAL_" + string(decision.Status), Title: texts[0],
		Body: fmt.Sprintf(texts[1], p.nameOf(decision.DeciderOrgID), p.Number),
		Link: proposalLinkRoot + p.ID, OwnersOnly: true,
	}
}

func (d Decision) IsAnnounced() bool {
	_, ok := noticeTexts()[d.Status]
	return ok
}

func (p Proposal) CounterpartOf(organizationID string) string {
	if p.Consignor.OrganizationID == organizationID {
		return p.Store.Name
	}
	return p.Consignor.Name
}

func (p Proposal) AgreementNotices(decision Decision, version int) []Notice {
	notices := make([]Notice, 0, len([]Party{p.Consignor, p.Store}))
	for _, party := range []Party{p.Consignor, p.Store} {
		notices = append(notices, Notice{
			OrganizationID: party.OrganizationID, ActorUserID: decision.DecidedBy, Type: "AGREEMENT_UPDATED",
			Title: "Kesepakatan diperbarui",
			Body:  fmt.Sprintf("Kesepakatan dengan %s kini versi %d.", p.CounterpartOf(party.OrganizationID), version),
			Link:  "/app/mitra/" + p.PartnershipID + "/kesepakatan", OwnersOnly: true,
		})
	}
	return notices
}
