package domain

import (
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

type CurrentItem struct {
	IsActive bool
	Terms    pricing.Terms
}

type CurrentAgreement struct {
	PartnershipID string
	IsActive      bool
	Consignor     Party
	Store         Party
	Terms         terms.General
	Items         map[string]CurrentItem
}

func (c CurrentAgreement) nameOf(organizationID string) string {
	if c.Consignor.OrganizationID == organizationID {
		return c.Consignor.Name
	}
	return c.Store.Name
}

func (c CurrentAgreement) counterpartOf(organizationID string) Party {
	if c.Consignor.OrganizationID == organizationID {
		return c.Store
	}
	return c.Consignor
}

type AmendmentParams struct {
	IDs      InvitationIDs
	Sender   Sender
	Current  CurrentAgreement
	Sequence int
	At       time.Time
}

func (d RevisionDraft) Amend(params AmendmentParams) (Opening, error) {
	current := params.Current
	if !current.IsActive {
		return Opening{}, apperror.New(apperror.CodePartnershipNotActive)
	}
	items, problems := current.resolve(d.Items)
	if len(items) == 0 && d.Terms == current.Terms {
		problems = append(problems, apperror.InvalidField("items", apperror.FieldRequired))
	}
	if len(problems) > 0 {
		return Opening{}, apperror.Validation(problems...)
	}
	d.Items = items
	return Opening{
		Kind: Amendment, IDs: params.IDs, Number: Number(params.At, params.Sequence), Sender: params.Sender,
		ConsignorID: current.Consignor.OrganizationID, StoreID: current.Store.OrganizationID,
		AwaitingID:    current.counterpartOf(params.Sender.OrganizationID).OrganizationID,
		PartnershipID: current.PartnershipID, Draft: d, At: params.At, ExpiresAt: params.At.Add(OpenFor),
		SenderName: current.nameOf(params.Sender.OrganizationID),
	}, nil
}

func (c CurrentAgreement) resolve(changes []Item) ([]Item, []apperror.FieldError) {
	resolved := make([]Item, 0, len(changes))
	var problems []apperror.FieldError
	for index, change := range changes {
		existing, found := c.Items[change.ProductID]
		isAgreed := found && existing.IsActive
		if isAgreed == (change.Action == Add) {
			problems = append(problems, apperror.InvalidField(fmt.Sprintf("items[%d].product_id", index),
				apperror.FieldUnknownChoice))
		}
		if change.Action == Stop {
			change.Terms = existing.Terms
		}
		resolved = append(resolved, change)
	}
	return resolved, problems
}

func (c CurrentAgreement) CounterpartName(organizationID string) string {
	return c.counterpartOf(organizationID).Name
}

func (o Opening) ReceivedNotice() Notice {
	return Notice{
		OrganizationID: o.AwaitingID, ActorUserID: o.Sender.UserID, Type: "PROPOSAL_RECEIVED",
		Title: "Ajuan perubahan",
		Body:  fmt.Sprintf("%s mengajukan perubahan Kesepakatan (%s).", o.SenderName, o.Number),
		Link:  proposalLinkRoot + o.IDs.ProposalID, OwnersOnly: true,
	}
}
