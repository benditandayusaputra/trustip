package handler

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
)

type itemRequest struct {
	ProductID  string  `json:"product_id"`
	Qty        int     `json:"qty"`
	ExpiryDate *string `json:"expiry_date"`
}

type dropRequest struct {
	PartnershipID    string        `json:"partnership_id"`
	AgreementVersion int           `json:"agreement_version"`
	OccurredAt       *string       `json:"occurred_at"`
	Items            []itemRequest `json:"items"`
}

func (r dropRequest) toInput(now time.Time) (domain.RecordInput, error) {
	var problems []apperror.FieldError
	if !isUUID(r.PartnershipID) {
		problems = append(problems, apperror.InvalidField("partnership_id", apperror.FieldUnknownChoice))
	}
	occurredAt := now
	if r.OccurredAt != nil {
		parsed, err := time.Parse(time.RFC3339, *r.OccurredAt)
		occurredAt = parsed
		if err != nil {
			problems = append(problems, apperror.InvalidField("occurred_at", apperror.FieldInvalidDate))
		}
	}
	if len(problems) > 0 {
		return domain.RecordInput{}, apperror.Validation(problems...)
	}
	items := make([]domain.ItemInput, 0, len(r.Items))
	for _, item := range r.Items {
		expiry := ""
		if item.ExpiryDate != nil {
			expiry = *item.ExpiryDate
		}
		items = append(items, domain.ItemInput{ProductID: item.ProductID, Qty: item.Qty, ExpiryDate: expiry})
	}
	return domain.RecordInput{
		PartnershipID: r.PartnershipID, AgreementVersion: r.AgreementVersion, OccurredAt: occurredAt, Items: items,
	}, nil
}

type partyResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type splitResponse struct {
	Gross          int64 `json:"gross"`
	StoreShare     int64 `json:"store_share"`
	ConsignorShare int64 `json:"consignor_share"`
}

type termsResponse struct {
	PricingModel  string `json:"pricing_model"`
	RetailPrice   int64  `json:"retail_price"`
	CommissionBPS int    `json:"commission_bps"`
	DepositPrice  int64  `json:"deposit_price"`
}

type itemResponse struct {
	ID          string        `json:"id"`
	ProductID   string        `json:"product_id"`
	ProductName string        `json:"product_name"`
	SKU         string        `json:"sku"`
	Unit        string        `json:"unit"`
	BatchCode   string        `json:"batch_code"`
	ExpiryDate  *string       `json:"expiry_date"`
	Qty         int           `json:"qty"`
	Terms       termsResponse `json:"terms"`
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

type dropResponse struct {
	ID               string                `json:"id"`
	Number           string                `json:"number"`
	PartnershipID    string                `json:"partnership_id"`
	AgreementVersion int                   `json:"agreement_version"`
	Status           string                `json:"status"`
	OccurredAt       string                `json:"occurred_at"`
	ReceivedAt       string                `json:"received_at"`
	RecordedOffline  bool                  `json:"recorded_offline"`
	Store            partyResponse         `json:"store"`
	Consignor        partyResponse         `json:"consignor"`
	RecordedByName   string                `json:"recorded_by_name"`
	Confirmation     *confirmationResponse `json:"confirmation"`
	Items            []itemResponse        `json:"items"`
	TotalQty         int                   `json:"total_qty"`
	ValueIfSold      splitResponse         `json:"value_if_sold"`
	Seal             sealResponse          `json:"seal"`
}

func toDropResponse(sealed domain.SealedView) dropResponse {
	view := sealed.View
	return dropResponse{
		ID: view.ID, Number: view.Number, PartnershipID: view.PartnershipID, AgreementVersion: view.AgreementVersion,
		Status: string(view.Status), OccurredAt: timestamp(view.OccurredAt), ReceivedAt: timestamp(view.ReceivedAt),
		RecordedOffline: view.RecordedOffline,
		Store:           partyResponse{ID: view.StoreID, Name: view.StoreName},
		Consignor:       partyResponse{ID: view.ConsignorID, Name: view.ConsignorName},
		RecordedByName:  view.RecordedByName, Confirmation: toConfirmation(view),
		Items: toItems(view.ItemViews), TotalQty: view.TotalQty(), ValueIfSold: toSplit(valueIfSold(view.Items)),
		Seal: toSealResponse(sealed.Seal),
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

func toItems(items []domain.ItemView) []itemResponse {
	responses := make([]itemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, itemResponse{
			ID: item.ID, ProductID: item.ProductID, ProductName: item.ProductName, SKU: item.SKU, Unit: item.Unit,
			BatchCode: item.BatchCode, ExpiryDate: expiryDate(item.Item), Qty: item.Qty,
			Terms: termsResponse{
				PricingModel: string(item.Terms.Model), RetailPrice: item.Terms.RetailPrice.Int64(),
				CommissionBPS: item.Terms.CommissionBPS, DepositPrice: item.Terms.DepositPrice.Int64(),
			},
		})
	}
	return responses
}

func expiryDate(item domain.Item) *string {
	date, isGiven := item.Expiry.Date()
	if !isGiven {
		return nil
	}
	formatted := date.Format(time.DateOnly)
	return &formatted
}

func valueIfSold(items []domain.Item) pricing.Split {
	splits := make([]pricing.Split, 0, len(items))
	for _, item := range items {
		splits = append(splits, item.Terms.SplitOf(item.Qty))
	}
	return pricing.Summarize(splits)
}

func toSplit(split pricing.Split) splitResponse {
	return splitResponse{
		Gross: split.Gross.Int64(), StoreShare: split.StoreShare.Int64(), ConsignorShare: split.ConsignorShare.Int64(),
	}
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

func timestamp(at time.Time) string {
	return at.UTC().Format(time.RFC3339)
}
