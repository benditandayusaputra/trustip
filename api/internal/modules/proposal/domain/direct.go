package domain

import (
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

const (
	DailyNewProposalLimit = 10
	DeclineCooldown       = 7 * 24 * time.Hour
	targetField           = "target_organization_id"
	needField             = "need_id"
)

type Target struct {
	OrganizationID string
	Name           string
	Type           actor.OrganizationType
}

type NeedRef struct {
	ID        string
	StoreID   string
	IsOpen    bool
	ExpiresAt time.Time
}

func (n NeedRef) isGiven() bool {
	return n.ID != ""
}

type DirectParams struct {
	IDs            InvitationIDs
	Sender         Sender
	SenderName     string
	Target         Target
	Need           NeedRef
	LastDeclinedAt time.Time
	Sequence       int
	At             time.Time
}

func (d RevisionDraft) Propose(params DirectParams) (Opening, error) {
	if err := params.checkTarget(); err != nil {
		return Opening{}, err
	}
	if err := params.checkNeed(); err != nil {
		return Opening{}, err
	}
	if problems := d.onlyAdditions(); len(problems) > 0 {
		return Opening{}, apperror.Validation(problems...)
	}
	opening := Opening{
		Kind: New, IDs: params.IDs, Number: Number(params.At, params.Sequence), Sender: params.Sender,
		AwaitingID: params.Target.OrganizationID, SenderName: params.SenderName, Draft: d, At: params.At,
		ExpiresAt: params.At.Add(OpenFor), NeedID: params.Need.ID,
	}
	opening.ConsignorID, opening.StoreID = params.Target.OrganizationID, params.Sender.OrganizationID
	if params.Sender.OrganizationType == actor.Consignor {
		opening.ConsignorID, opening.StoreID = params.Sender.OrganizationID, params.Target.OrganizationID
	}
	return opening, nil
}

func (p DirectParams) checkTarget() error {
	isOpposite := p.Target.Type != p.Sender.OrganizationType
	if !isOpposite || p.Target.OrganizationID == p.Sender.OrganizationID {
		return apperror.Validation(apperror.InvalidField(targetField, apperror.FieldUnknownChoice))
	}
	if p.LastDeclinedAt.IsZero() {
		return nil
	}
	reopensAt := p.LastDeclinedAt.Add(DeclineCooldown)
	if p.At.Before(reopensAt) {
		return apperror.New(apperror.CodeProposalCooldown).WithParam(organizationParam, p.Target.Name).
			WithParam("tanggal", period.DisplayDate(reopensAt))
	}
	return nil
}

func (p DirectParams) checkNeed() error {
	if !p.Need.isGiven() {
		return nil
	}
	isForTarget := p.Need.StoreID == p.Target.OrganizationID && p.Sender.OrganizationType == actor.Consignor
	if !isForTarget {
		return apperror.Validation(apperror.InvalidField(needField, apperror.FieldUnknownChoice))
	}
	if !p.Need.IsOpen || !p.At.Before(p.Need.ExpiresAt) {
		return apperror.New(apperror.CodeNeedNotOpen)
	}
	return nil
}

type receivedText struct {
	title string
	body  func(o Opening) string
}

func receivedTexts() map[Kind]receivedText {
	return map[Kind]receivedText{
		New: {"Ajuan kerja sama", func(o Opening) string {
			return fmt.Sprintf("%s mengajukan kerja sama.", o.SenderName)
		}},
		Amendment: {"Ajuan perubahan", func(o Opening) string {
			return fmt.Sprintf("%s mengajukan perubahan Kesepakatan (%s).", o.SenderName, o.Number)
		}},
	}
}

func (o Opening) ReceivedNotice() Notice {
	text := receivedTexts()[o.Kind]
	return Notice{
		OrganizationID: o.AwaitingID, ActorUserID: o.Sender.UserID, Type: "PROPOSAL_RECEIVED", Title: text.title,
		Body: text.body(o), Link: proposalLinkRoot + o.IDs.ProposalID, OwnersOnly: true,
	}
}

type QuotaKey struct {
	OrganizationID string
	Day            string
}

func QuotaKeyOf(organizationID string, at time.Time) QuotaKey {
	return QuotaKey{OrganizationID: organizationID, Day: period.BusinessDate(at)}
}

type DirectPair struct {
	SenderID string
	TargetID string
	NeedID   string
}

type DirectParties struct {
	SenderName     string
	Target         Target
	Need           NeedRef
	LastDeclinedAt time.Time
	ArePartners    bool
}
