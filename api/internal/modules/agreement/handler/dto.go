package handler

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
)

type partyResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type termsResponse struct {
	SettlementSchedule string `json:"settlement_schedule"`
	PaymentTermsDays   int    `json:"payment_terms_days"`
	CountRhythm        string `json:"count_rhythm"`
	Notes              string `json:"notes"`
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
	PricingModel  string        `json:"pricing_model"`
	RetailPrice   int64         `json:"retail_price"`
	CommissionBPS int           `json:"commission_bps"`
	DepositPrice  int64         `json:"deposit_price"`
	IsActive      bool          `json:"is_active"`
	PerUnit       splitResponse `json:"per_unit"`
}

type personResponse struct {
	Name             string `json:"name"`
	OrganizationName string `json:"organization_name"`
}

type trailResponse struct {
	ProposedBy           personResponse `json:"proposed_by"`
	AcceptedBy           personResponse `json:"accepted_by"`
	AcceptedAt           string         `json:"accepted_at"`
	SourceProposalID     string         `json:"source_proposal_id"`
	SourceProposalNumber string         `json:"source_proposal_number"`
}

type sealEventResponse struct {
	EventType     string `json:"event_type"`
	ActorOrgName  string `json:"actor_org_name"`
	ActorUserName string `json:"actor_user_name"`
	OccurredAt    string `json:"occurred_at"`
	Code          string `json:"code"`
}

type sealResponse struct {
	IsValid bool                `json:"is_valid"`
	Events  []sealEventResponse `json:"events"`
}

type agreementResponse struct {
	ID            string         `json:"id"`
	PartnershipID string         `json:"partnership_id"`
	Version       int            `json:"version"`
	Status        string         `json:"status"`
	Number        string         `json:"number"`
	Consignor     partyResponse  `json:"consignor"`
	Store         partyResponse  `json:"store"`
	Terms         termsResponse  `json:"terms"`
	Items         []itemResponse `json:"items"`
	Trail         trailResponse  `json:"trail"`
	Seal          sealResponse   `json:"seal"`
}

func toResponse(sealed domain.SealedView) agreementResponse {
	view := sealed.View
	agreement := view.Agreement
	return agreementResponse{
		ID: agreement.ID, PartnershipID: agreement.PartnershipID, Version: agreement.Version,
		Status: string(agreement.Status), Number: agreement.Number(),
		Consignor: partyResponse{ID: view.Consignor.OrganizationID, Name: view.Consignor.Name},
		Store:     partyResponse{ID: view.Store.OrganizationID, Name: view.Store.Name},
		Terms: termsResponse{
			SettlementSchedule: string(agreement.Terms.Schedule), PaymentTermsDays: agreement.Terms.PaymentTermsDays,
			CountRhythm: string(agreement.Terms.Rhythm), Notes: agreement.Terms.Notes,
		},
		Items: toItems(agreement.Items),
		Trail: trailResponse{
			ProposedBy: personResponse(view.Proposer), AcceptedBy: personResponse(view.Acceptor),
			AcceptedAt: agreement.AcceptedAt.UTC().Format(time.RFC3339), SourceProposalID: agreement.SourceProposalID,
			SourceProposalNumber: agreement.ProposalNumber,
		},
		Seal: toSealResponse(sealed.Seal),
	}
}

func toItems(items []domain.Item) []itemResponse {
	responses := make([]itemResponse, 0, len(items))
	for _, item := range items {
		perUnit := item.Terms.PerUnit()
		responses = append(responses, itemResponse{
			ProductID: item.ProductID, ProductName: item.ProductName, SKU: item.SKU, Unit: item.Unit,
			PricingModel: string(item.Terms.Model), RetailPrice: item.Terms.RetailPrice.Int64(),
			CommissionBPS: item.Terms.CommissionBPS, DepositPrice: item.Terms.DepositPrice.Int64(),
			IsActive: item.IsActive,
			PerUnit: splitResponse{
				Gross: perUnit.Gross.Int64(), StoreShare: perUnit.StoreShare.Int64(),
				ConsignorShare: perUnit.ConsignorShare.Int64(),
			},
		})
	}
	return responses
}

func toSealResponse(verdict domain.SealVerdict) sealResponse {
	events := make([]sealEventResponse, 0, len(verdict.Events))
	for _, event := range verdict.Events {
		events = append(events, sealEventResponse{
			EventType: event.Event, ActorOrgName: event.ActorOrgName, ActorUserName: event.ActorUserName,
			OccurredAt: event.At.UTC().Format(time.RFC3339), Code: event.Code,
		})
	}
	return sealResponse{IsValid: verdict.IsValid, Events: events}
}
