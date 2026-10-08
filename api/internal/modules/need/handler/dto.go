package handler

import (
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type needRequest struct {
	Title                string  `json:"title"`
	CategoryID           string  `json:"category_id"`
	Description          string  `json:"description"`
	EstQtyPerWeek        *int64  `json:"est_qty_per_week"`
	MaxRetailPrice       *int64  `json:"max_retail_price"`
	PreferredModel       *string `json:"preferred_model"`
	OfferedCommissionBPS *int64  `json:"offered_commission_bps"`
}

func (r needRequest) toForm() (domain.Form, error) {
	title := httpserver.LineInput("title", r.Title)
	description := httpserver.MultilineInput("description", r.Description)
	if err := httpserver.RejectUncleanInputs(title, description); err != nil {
		return domain.Form{}, fmt.Errorf("check need text: %w", err)
	}
	model := ""
	if r.PreferredModel != nil {
		model = *r.PreferredModel
	}
	return domain.Form{
		Title: title.Value, CategoryID: r.CategoryID, Description: description.Value,
		WeeklyQuantity: amountOf(r.EstQtyPerWeek), MaxRetailPrice: amountOf(r.MaxRetailPrice),
		PreferredModel: model, OfferedCommissionBPS: amountOf(r.OfferedCommissionBPS),
	}, nil
}

func amountOf(value *int64) domain.Amount {
	if value == nil {
		return domain.Amount{}
	}
	return domain.Amount{Value: *value, IsGiven: true}
}

type refResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type needResponse struct {
	ID                   string      `json:"id"`
	Store                refResponse `json:"store"`
	Title                string      `json:"title"`
	Category             refResponse `json:"category"`
	Description          string      `json:"description"`
	EstQtyPerWeek        *int64      `json:"est_qty_per_week"`
	MaxRetailPrice       *int64      `json:"max_retail_price"`
	PreferredModel       *string     `json:"preferred_model"`
	OfferedCommissionBPS *int64      `json:"offered_commission_bps"`
	Status               string      `json:"status"`
	ExpiresAt            string      `json:"expires_at"`
	CreatedAt            string      `json:"created_at"`
	ResponseCount        int         `json:"response_count"`
}

type responseItem struct {
	ProposalID    string  `json:"proposal_id"`
	Number        string  `json:"number"`
	Status        string  `json:"status"`
	FromOrgID     string  `json:"from_org_id"`
	AwaitingOrgID *string `json:"awaiting_org_id"`
	ConsignorName string  `json:"consignor_name"`
	CreatedAt     string  `json:"created_at"`
}

type detailResponse struct {
	needResponse
	Responses []responseItem `json:"responses"`
}

func toNeedResponse(need domain.Need) needResponse {
	draft := need.Draft
	return needResponse{
		ID: need.ID, Store: refResponse{ID: need.StoreID, Name: need.StoreName}, Title: draft.Title,
		Category:    refResponse{ID: draft.CategoryID, Name: need.CategoryName},
		Description: draft.Description, EstQtyPerWeek: optionalAmount(draft.WeeklyQuantity),
		MaxRetailPrice: optionalAmount(draft.MaxRetailPrice), PreferredModel: optionalText(draft.PreferredModel),
		OfferedCommissionBPS: optionalAmount(draft.OfferedCommissionBPS), Status: string(need.Status),
		ExpiresAt: need.ExpiresAt.UTC().Format(time.RFC3339), CreatedAt: need.CreatedAt.UTC().Format(time.RFC3339),
		ResponseCount: need.ResponseCount,
	}
}

func toDetailResponse(detail domain.Detail) detailResponse {
	responses := make([]responseItem, 0, len(detail.Responses))
	for _, response := range detail.Responses {
		responses = append(responses, responseItem{
			ProposalID: response.ProposalID, Number: response.Number, Status: response.Status,
			FromOrgID: response.FromOrgID, AwaitingOrgID: optionalText(response.AwaitingOrgID),
			ConsignorName: response.ConsignorName, CreatedAt: response.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return detailResponse{needResponse: toNeedResponse(detail.Need), Responses: responses}
}

func optionalAmount(amount domain.Amount) *int64 {
	if !amount.IsGiven {
		return nil
	}
	return &amount.Value
}

func optionalText(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
