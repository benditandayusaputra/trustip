package handler

import "github.com/benditandayusaputra/trustip/api/internal/modules/insight/domain"

type dropSuggestionResponse struct {
	ProductID     string `json:"product_id"`
	SuggestedQty  int    `json:"suggested_qty"`
	HasEnoughData bool   `json:"has_enough_data"`
}

func toDropSuggestions(suggestions []domain.DropSuggestion) []dropSuggestionResponse {
	responses := make([]dropSuggestionResponse, 0, len(suggestions))
	for _, suggestion := range suggestions {
		responses = append(responses, dropSuggestionResponse{
			ProductID: suggestion.ProductID, SuggestedQty: suggestion.Qty, HasEnoughData: suggestion.HasEnoughData,
		})
	}
	return responses
}
