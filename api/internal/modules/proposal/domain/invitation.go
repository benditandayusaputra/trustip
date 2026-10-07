package domain

import (
	"fmt"
	"regexp"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

const (
	OpenFor        = 14 * 24 * time.Hour
	NumberPrefix   = "AJ"
	inviteAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	InviteLength   = 10
)

var inviteFormat = regexp.MustCompile(`^[A-HJ-NP-Z2-9]{10}$`)

func InviteAlphabet() string {
	return inviteAlphabet
}

func IsInviteCode(code string) bool {
	return inviteFormat.MatchString(code)
}

type Kind string

const (
	New       Kind = "NEW"
	Amendment Kind = "AMENDMENT"
)

type Status string

const (
	Open      Status = "OPEN"
	Accepted  Status = "ACCEPTED"
	Declined  Status = "DECLINED"
	Withdrawn Status = "WITHDRAWN"
	Expired   Status = "EXPIRED"
)

type Sender struct {
	UserID           string
	OrganizationID   string
	OrganizationType actor.OrganizationType
}

type InvitationIDs struct {
	ProposalID string
	RevisionID string
	Code       string
}

type InvitationParams struct {
	IDs      InvitationIDs
	Sender   Sender
	Sequence int
	At       time.Time
}

type Invitation struct {
	IDs       InvitationIDs
	Number    string
	Sender    Sender
	Draft     RevisionDraft
	At        time.Time
	ExpiresAt time.Time
}

func (d RevisionDraft) Invite(params InvitationParams) (Invitation, error) {
	if params.Sender.OrganizationType == actor.Store && len(d.Items) > 0 {
		return Invitation{}, apperror.Validation(field.Collect(field.Invalid("items", apperror.FieldUnknownChoice))...)
	}
	if problems := d.onlyAdditions(); len(problems) > 0 {
		return Invitation{}, apperror.Validation(problems...)
	}
	return Invitation{
		IDs: params.IDs, Number: Number(params.At, params.Sequence), Sender: params.Sender, Draft: d,
		At: params.At, ExpiresAt: params.At.Add(OpenFor),
	}, nil
}

func Number(at time.Time, sequence int) string {
	return fmt.Sprintf("%s-%s-%04d", NumberPrefix, period.MonthKey(at), sequence)
}

func (i Invitation) ConsignorID() string {
	return i.partyOf(actor.Consignor)
}

func (i Invitation) StoreID() string {
	return i.partyOf(actor.Store)
}

func (i Invitation) partyOf(organizationType actor.OrganizationType) string {
	if i.Sender.OrganizationType == organizationType {
		return i.Sender.OrganizationID
	}
	return ""
}

type SenderView struct {
	OrganizationID string
	Name           string
	Type           actor.OrganizationType
}

type InvitationView struct {
	ProposalID  string
	Status      Status
	ExpiresAt   time.Time
	Sender      SenderView
	ConsignorID string
	StoreID     string
	ItemCount   int
}

type Claim struct {
	ProposalID  string
	ConsignorID string
	StoreID     string
	ClaimerID   string
	IsRepeat    bool
	At          time.Time
}

func (v InvitationView) ClaimBy(claimer actor.Membership, claimedAt time.Time) (Claim, error) {
	claim := Claim{
		ProposalID: v.ProposalID, ConsignorID: v.ConsignorID, StoreID: v.StoreID, ClaimerID: claimer.OrganizationID,
		At: claimedAt,
	}
	if v.isParty(claimer.OrganizationID) {
		claim.IsRepeat = true
		return claim, nil
	}
	if err := v.checkClaimable(claimer, claimedAt); err != nil {
		return Claim{}, err
	}
	if claimer.OrganizationType == actor.Store {
		claim.StoreID = claimer.OrganizationID
		return claim, nil
	}
	claim.ConsignorID = claimer.OrganizationID
	return claim, nil
}

func (v InvitationView) checkClaimable(claimer actor.Membership, claimedAt time.Time) error {
	if v.Status != Open || v.isClaimed() {
		return apperror.OutsideScope()
	}
	if !claimedAt.Before(v.ExpiresAt) {
		return apperror.New(apperror.CodeProposalExpired)
	}
	if claimer.OrganizationType == v.Sender.Type {
		wanted := typeLabel(oppositeOf(v.Sender.Type))
		return apperror.New(apperror.CodeInviteWrongOrgType).WithParam("jenis", wanted)
	}
	return nil
}

func (v InvitationView) isParty(organizationID string) bool {
	return organizationID == v.ConsignorID || organizationID == v.StoreID
}

func (v InvitationView) isClaimed() bool {
	return v.ConsignorID != "" && v.StoreID != ""
}

func oppositeOf(organizationType actor.OrganizationType) actor.OrganizationType {
	if organizationType == actor.Store {
		return actor.Consignor
	}
	return actor.Store
}

func typeLabel(organizationType actor.OrganizationType) string {
	labels := map[actor.OrganizationType]string{actor.Consignor: "penitip", actor.Store: "toko"}
	return labels[organizationType]
}
