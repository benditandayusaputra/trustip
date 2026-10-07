package domain

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

const (
	EntityType   = "AGREEMENT"
	firstVersion = 1
)

var ErrAlreadyPartners = errors.New("the pair already has an active partnership")

type Status string

const (
	Active     Status = "ACTIVE"
	Superseded Status = "SUPERSEDED"
)

type Item struct {
	ProductID   string
	ProductName string
	SKU         string
	Unit        string
	Terms       pricing.Terms
	IsActive    bool
}

type FoundingInput struct {
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

type FoundingIDs struct {
	PartnershipID string
	AgreementID   string
}

type Partnership struct {
	ID          string
	ConsignorID string
	StoreID     string
	StartedAt   time.Time
}

type Agreement struct {
	ID               string
	PartnershipID    string
	Version          int
	Status           Status
	Terms            terms.General
	Items            []Item
	SourceProposalID string
	ProposalNumber   string
	ProposedBy       string
	AcceptedBy       string
	AcceptedAt       time.Time
}

type Founding struct {
	Partnership Partnership
	Agreement   Agreement
}

func Found(input FoundingInput, ids FoundingIDs) Founding {
	items := make([]Item, 0, len(input.Items))
	for _, item := range input.Items {
		item.IsActive = true
		items = append(items, item)
	}
	return Founding{
		Partnership: Partnership{
			ID: ids.PartnershipID, ConsignorID: input.ConsignorID, StoreID: input.StoreID, StartedAt: input.At,
		},
		Agreement: Agreement{
			ID: ids.AgreementID, PartnershipID: ids.PartnershipID, Version: firstVersion, Status: Active,
			Terms: input.Terms, Items: items, SourceProposalID: input.ProposalID, ProposalNumber: input.ProposalNumber,
			ProposedBy: input.ProposedBy, AcceptedBy: input.AcceptedBy, AcceptedAt: input.At,
		},
	}
}

func (a Agreement) Number() string {
	return fmt.Sprintf("%s/V%d", a.ProposalNumber, a.Version)
}

func (a Agreement) Content() map[string]any {
	items := slices.Clone(a.Items)
	slices.SortFunc(items, func(left, right Item) int { return cmp.Compare(left.ProductID, right.ProductID) })
	contentItems := make([]any, 0, len(items))
	for _, item := range items {
		contentItems = append(contentItems, map[string]any{
			"product_id": item.ProductID, "pricing_model": string(item.Terms.Model),
			"retail_price": item.Terms.RetailPrice.Int64(), "commission_bps": item.Terms.CommissionBPS,
			"deposit_price": item.Terms.DepositPrice.Int64(), "state": item.state(),
		})
	}
	return map[string]any{
		"agreement_id": a.ID, "partnership_id": a.PartnershipID, "version": a.Version,
		"settlement_schedule": string(a.Terms.Schedule), "payment_terms_days": a.Terms.PaymentTermsDays,
		"count_rhythm": string(a.Terms.Rhythm), "notes": a.Terms.Notes, "source_proposal_id": a.SourceProposalID,
		"proposed_by": a.ProposedBy, "accepted_by": a.AcceptedBy, "items": contentItems,
	}
}

func (i Item) state() string {
	if i.IsActive {
		return "ACTIVE"
	}
	return "STOPPED"
}

type SealRequest struct {
	PartnershipID  string
	Event          string
	DocumentType   string
	DocumentID     string
	DocumentNumber string
	Content        map[string]any
	ActorUserID    string
	ActorOrgID     string
	At             time.Time
}

const acceptedEvent = "AGREEMENT_ACCEPTED"

func (f Founding) SealRequest(input FoundingInput) SealRequest {
	agreement := f.Agreement
	return SealRequest{
		PartnershipID: agreement.PartnershipID, Event: acceptedEvent, DocumentType: EntityType,
		DocumentID: agreement.ID, DocumentNumber: agreement.Number(), Content: agreement.Content(),
		ActorUserID: input.AcceptedBy, ActorOrgID: input.AcceptorOrgID, At: input.At,
	}
}
