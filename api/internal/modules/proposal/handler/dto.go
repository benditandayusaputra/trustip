package handler

import (
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type termsBody struct {
	SettlementSchedule string `json:"settlement_schedule"`
	PaymentTermsDays   int    `json:"payment_terms_days"`
	CountRhythm        string `json:"count_rhythm"`
	Notes              string `json:"notes"`
}

type itemRequest struct {
	ProductID     string `json:"product_id"`
	Action        string `json:"action"`
	PricingModel  string `json:"pricing_model"`
	RetailPrice   int64  `json:"retail_price"`
	CommissionBPS int    `json:"commission_bps"`
	DepositPrice  int64  `json:"deposit_price"`
}

type invitationRequest struct {
	Terms   termsBody     `json:"terms"`
	Items   []itemRequest `json:"items"`
	Message string        `json:"message"`
}

func (r invitationRequest) toInput() (domain.RevisionInput, error) {
	notes := httpserver.MultilineInput("terms.notes", r.Terms.Notes)
	message := httpserver.MultilineInput("message", r.Message)
	if err := httpserver.RejectUncleanInputs(notes, message); err != nil {
		return domain.RevisionInput{}, fmt.Errorf("check invitation text: %w", err)
	}
	items := make([]domain.ItemInput, 0, len(r.Items))
	for _, item := range r.Items {
		items = append(items, domain.ItemInput{
			ProductID: item.ProductID, Action: item.Action,
			Terms: pricing.TermsInput{
				Model: item.PricingModel, RetailPrice: item.RetailPrice, CommissionBPS: item.CommissionBPS,
				DepositPrice: item.DepositPrice,
			},
		})
	}
	return domain.RevisionInput{
		Terms: terms.Input{
			SettlementSchedule: r.Terms.SettlementSchedule, PaymentTermsDays: r.Terms.PaymentTermsDays,
			CountRhythm: r.Terms.CountRhythm, Notes: notes.Value,
		},
		Items: items, Message: message.Value,
	}, nil
}

type invitationResponse struct {
	SenderName string `json:"sender_name"`
	SenderType string `json:"sender_type"`
	ItemCount  int    `json:"item_count"`
	ExpiresAt  string `json:"expires_at"`
}

type splitResponse struct {
	Gross          int64 `json:"gross"`
	StoreShare     int64 `json:"store_share"`
	ConsignorShare int64 `json:"consignor_share"`
}

type itemResponse struct {
	ProductID     string        `json:"product_id"`
	ProductName   string        `json:"product_name"`
	SKU           string        `json:"sku"`
	Unit          string        `json:"unit"`
	Action        string        `json:"action"`
	PricingModel  string        `json:"pricing_model"`
	RetailPrice   int64         `json:"retail_price"`
	CommissionBPS int           `json:"commission_bps"`
	DepositPrice  int64         `json:"deposit_price"`
	PerUnit       splitResponse `json:"per_unit"`
}

type partyResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type revisionResponse struct {
	RevisionNo  int            `json:"revision_no"`
	AuthorOrgID string         `json:"author_org_id"`
	Terms       termsBody      `json:"terms"`
	Message     string         `json:"message"`
	Items       []itemResponse `json:"items"`
	CreatedAt   string         `json:"created_at"`
}

type proposalResponse struct {
	ID             string           `json:"id"`
	Number         string           `json:"number"`
	Kind           string           `json:"kind"`
	Status         string           `json:"status"`
	Consignor      *partyResponse   `json:"consignor"`
	Store          *partyResponse   `json:"store"`
	FromOrgID      string           `json:"from_org_id"`
	AwaitingOrgID  *string          `json:"awaiting_org_id"`
	InviteCode     *string          `json:"invite_code"`
	PartnershipID  *string          `json:"partnership_id"`
	DecisionReason string           `json:"decision_reason"`
	ExpiresAt      string           `json:"expires_at"`
	CreatedAt      string           `json:"created_at"`
	Revision       revisionResponse `json:"revision"`
}

func toProposalResponse(proposal domain.Proposal) proposalResponse {
	return proposalResponse{
		ID: proposal.ID, Number: proposal.Number, Kind: string(proposal.Kind), Status: string(proposal.Status),
		Consignor: partyOf(proposal.Consignor), Store: partyOf(proposal.Store), FromOrgID: proposal.FromOrgID,
		AwaitingOrgID: optional(proposal.AwaitingOrgID), InviteCode: optional(proposal.InviteCode),
		PartnershipID: optional(proposal.PartnershipID), DecisionReason: proposal.DecisionReason,
		ExpiresAt: proposal.ExpiresAt.UTC().Format(time.RFC3339),
		CreatedAt: proposal.CreatedAt.UTC().Format(time.RFC3339),
		Revision:  toRevisionResponse(proposal.Revision),
	}
}

func toRevisionResponse(revision domain.Revision) revisionResponse {
	items := make([]itemResponse, 0, len(revision.Items))
	for _, item := range revision.Items {
		perUnit := item.Terms.PerUnit()
		items = append(items, itemResponse{
			ProductID: item.ProductID, ProductName: item.ProductName, SKU: item.SKU, Unit: item.Unit,
			Action: string(item.Action), PricingModel: string(item.Terms.Model),
			RetailPrice: item.Terms.RetailPrice.Int64(), CommissionBPS: item.Terms.CommissionBPS,
			DepositPrice: item.Terms.DepositPrice.Int64(),
			PerUnit: splitResponse{
				Gross: perUnit.Gross.Int64(), StoreShare: perUnit.StoreShare.Int64(),
				ConsignorShare: perUnit.ConsignorShare.Int64(),
			},
		})
	}
	terms := revision.Terms
	return revisionResponse{
		RevisionNo: revision.No, AuthorOrgID: revision.AuthorOrgID, Message: revision.Message, Items: items,
		Terms: termsBody{
			SettlementSchedule: string(terms.Schedule), PaymentTermsDays: terms.PaymentTermsDays,
			CountRhythm: string(terms.Rhythm), Notes: terms.Notes,
		},
		CreatedAt: revision.At.UTC().Format(time.RFC3339),
	}
}

func partyOf(party domain.Party) *partyResponse {
	if party.OrganizationID == "" {
		return nil
	}
	return &partyResponse{ID: party.OrganizationID, Name: party.Name}
}

func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
