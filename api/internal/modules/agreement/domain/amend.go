package domain

import (
	"cmp"
	"slices"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

type ChangeAction string

const (
	AddProduct    ChangeAction = "ADD"
	UpdateProduct ChangeAction = "UPDATE"
	StopProduct   ChangeAction = "STOP"
)

type Change struct {
	ProductID string
	Action    ChangeAction
	Terms     pricing.Terms
}

type AmendmentInput struct {
	PartnershipID  string
	ProposalID     string
	ProposalNumber string
	Terms          terms.General
	Changes        []Change
	ProposedBy     string
	AcceptedBy     string
	AcceptorOrgID  string
	At             time.Time
}

func (a Agreement) Amend(input AmendmentInput, nextID string) Agreement {
	items := map[string]Item{}
	for _, item := range a.Items {
		items[item.ProductID] = item
	}
	for _, change := range input.Changes {
		item := items[change.ProductID]
		item.ProductID = change.ProductID
		item.IsActive = change.Action != StopProduct
		if item.IsActive {
			item.Terms = change.Terms
		}
		items[change.ProductID] = item
	}
	next := make([]Item, 0, len(items))
	for _, item := range items {
		next = append(next, item)
	}
	slices.SortFunc(next, func(left, right Item) int { return cmp.Compare(left.ProductID, right.ProductID) })
	return Agreement{
		ID: nextID, PartnershipID: a.PartnershipID, Version: a.Version + 1, Status: Active, Terms: input.Terms,
		Items: next, SourceProposalID: input.ProposalID, ProposalNumber: input.ProposalNumber,
		ProposedBy: input.ProposedBy, AcceptedBy: input.AcceptedBy, AcceptedAt: input.At,
	}
}

func (a Agreement) SealRequest(actorUserID, actorOrgID string) SealRequest {
	return SealRequest{
		PartnershipID: a.PartnershipID, Event: acceptedEvent, DocumentType: EntityType, DocumentID: a.ID,
		DocumentNumber: a.Number(), Content: a.Content(), ActorUserID: actorUserID, ActorOrgID: actorOrgID,
		At: a.AcceptedAt,
	}
}
