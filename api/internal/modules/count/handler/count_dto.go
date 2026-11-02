package handler

import (
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
)

type lineResponse struct {
	ID           string        `json:"id"`
	ProductID    string        `json:"product_id"`
	ProductName  string        `json:"product_name"`
	SKU          string        `json:"sku"`
	Unit         string        `json:"unit"`
	RecordedQty  int           `json:"recorded_qty"`
	CountedQty   int           `json:"counted_qty"`
	SoldQty      int           `json:"sold_qty"`
	ReturnedQty  int           `json:"returned_qty"`
	KeptQty      int           `json:"kept_qty"`
	ReturnReason *string       `json:"return_reason"`
	Split        splitResponse `json:"split"`
}

type confirmationResponse struct {
	ByName      string `json:"by_name"`
	At          string `json:"at"`
	IsAutomatic bool   `json:"is_automatic"`
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

type countResponse struct {
	ID               string                `json:"id"`
	Number           string                `json:"number"`
	PartnershipID    string                `json:"partnership_id"`
	Status           string                `json:"status"`
	Occasion         string                `json:"occasion"`
	Note             string                `json:"note"`
	CountedAt        string                `json:"counted_at"`
	ClientRecordedAt string                `json:"client_recorded_at"`
	ReceivedAt       string                `json:"received_at"`
	RecordedOffline  bool                  `json:"recorded_offline"`
	Store            partyResponse         `json:"store"`
	Consignor        partyResponse         `json:"consignor"`
	RecordedByName   string                `json:"recorded_by_name"`
	Confirmation     *confirmationResponse `json:"confirmation"`
	Lines            []lineResponse        `json:"lines"`
	SoldQty          int                   `json:"sold_qty"`
	ReturnedQty      int                   `json:"returned_qty"`
	Split            splitResponse         `json:"split"`
	Settlement       *settlementRef        `json:"settlement"`
	Seal             sealResponse          `json:"seal"`
}

type settlementRef struct {
	ID     string `json:"id"`
	Number string `json:"number"`
}

func toSettlementRef(ref domain.SettlementRef) *settlementRef {
	if ref.ID == "" {
		return nil
	}
	return &settlementRef{ID: ref.ID, Number: ref.Number}
}

func toCountResponse(sealed domain.SealedView) countResponse {
	view := sealed.View
	return countResponse{
		ID: view.ID, Number: view.Number, PartnershipID: view.PartnershipID, Status: string(view.Status),
		Occasion: string(view.Occasion), Note: view.Note, CountedAt: timestamp(view.CountedAt),
		ClientRecordedAt: timestamp(view.ClientRecordedAt), ReceivedAt: timestamp(view.ReceivedAt),
		RecordedOffline: view.RecordedOffline,
		Store:           partyResponse{ID: view.StoreID, Name: view.StoreName},
		Consignor:       partyResponse{ID: view.ConsignorID, Name: view.ConsignorName},
		RecordedByName:  view.RecordedByName, Confirmation: toConfirmation(view), Lines: toLines(view),
		SoldQty: view.SoldQty, ReturnedQty: view.ReturnedQty, Split: toSplit(view.Split),
		Settlement: toSettlementRef(view.Settlement), Seal: toSealResponse(sealed.Seal),
	}
}

func toConfirmation(view domain.View) *confirmationResponse {
	if !view.IsConfirmed {
		return nil
	}
	return &confirmationResponse{
		ByName: view.ConfirmedByName, At: timestamp(view.ConfirmedAt), IsAutomatic: view.AutoConfirmed,
	}
}

func toLines(view domain.View) []lineResponse {
	lines := make([]lineResponse, 0, len(view.Lines))
	for _, line := range view.Lines {
		product := view.Products[line.ProductID]
		splits := make([]pricing.Split, 0, len(line.Sold))
		for _, sale := range line.Sold {
			splits = append(splits, sale.Split)
		}
		lines = append(lines, lineResponse{
			ID: line.ID, ProductID: line.ProductID, ProductName: product.Name, SKU: product.SKU, Unit: product.Unit,
			RecordedQty: line.RecordedQty, CountedQty: line.CountedQty, SoldQty: line.SoldQty,
			ReturnedQty: line.ReturnedQty, KeptQty: line.KeptQty, ReturnReason: reasonOf(line.ReturnReason),
			Split: toSplit(pricing.Summarize(splits)),
		})
	}
	return lines
}

func reasonOf(reason domain.ReturnReason) *string {
	if reason == "" {
		return nil
	}
	text := string(reason)
	return &text
}

func toSealResponse(verdict domain.SealVerdict) sealResponse {
	events := make([]sealEventResponse, 0, len(verdict.Events))
	for _, event := range verdict.Events {
		events = append(events, sealEventResponse{
			EventType: event.Event, ActorOrgName: event.ActorOrgName, ActorUserName: event.ActorUserName,
			OccurredAt: timestamp(event.At), Code: event.Code,
		})
	}
	return sealResponse{IsValid: verdict.IsValid, Events: events}
}

type counterpartResponse struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	LogoURL *string `json:"logo_url"`
}

type summaryResponse struct {
	ID              string              `json:"id"`
	Number          string              `json:"number"`
	PartnershipID   string              `json:"partnership_id"`
	Status          string              `json:"status"`
	CountedAt       string              `json:"counted_at"`
	RecordedOffline bool                `json:"recorded_offline"`
	Counterpart     counterpartResponse `json:"counterpart"`
	SoldQty         int                 `json:"sold_qty"`
	ReturnedQty     int                 `json:"returned_qty"`
	ConsignorShare  int64               `json:"consignor_share"`
}

func toSummaryResponse(summary domain.Summary, filesBaseURL string) summaryResponse {
	counterpart := counterpartResponse{
		ID: summary.CounterpartID, Name: summary.CounterpartName, Type: summary.CounterpartType,
	}
	if summary.CounterpartLogo != "" {
		logoURL := filesBaseURL + "/" + summary.CounterpartLogo
		counterpart.LogoURL = &logoURL
	}
	return summaryResponse{
		ID: summary.ID, Number: summary.Number, PartnershipID: summary.PartnershipID, Status: string(summary.Status),
		CountedAt: timestamp(summary.CountedAt), RecordedOffline: summary.RecordedOffline, Counterpart: counterpart,
		SoldQty: summary.SoldQty, ReturnedQty: summary.ReturnedQty, ConsignorShare: summary.ConsignorShare,
	}
}
