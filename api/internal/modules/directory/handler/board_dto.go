package handler

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
)

type needStoreResponse struct {
	ID        string  `json:"id"`
	Slug      string  `json:"slug"`
	Name      string  `json:"name"`
	CityName  string  `json:"city_name"`
	StoreType *string `json:"store_type"`
}

type needCardResponse struct {
	ID                   string            `json:"id"`
	Title                string            `json:"title"`
	Description          string            `json:"description"`
	Category             categoryResponse  `json:"category"`
	EstQtyPerWeek        *int              `json:"est_qty_per_week"`
	MaxRetailPrice       *int64            `json:"max_retail_price"`
	PreferredModel       *string           `json:"preferred_model"`
	OfferedCommissionBPS *int              `json:"offered_commission_bps"`
	ExpiresAt            string            `json:"expires_at"`
	CreatedAt            string            `json:"created_at"`
	Store                needStoreResponse `json:"store"`
	DistanceM            *int              `json:"distance_m"`
	IsOwn                bool              `json:"is_own"`
}

func needCardOf(card domain.NeedCard) needCardResponse {
	response := needCardResponse{
		ID: card.ID, Title: card.Title, Description: card.Description, Category: categoryResponse(card.Category),
		PreferredModel: optionalText(card.PreferredModel), ExpiresAt: card.ExpiresAt.UTC().Format(time.RFC3339),
		CreatedAt: card.CreatedAt.UTC().Format(time.RFC3339), DistanceM: distanceOf(card.Distance), IsOwn: card.IsOwn,
		Store: needStoreResponse{
			ID: card.Store.ID, Slug: card.Store.Slug, Name: card.Store.Name, CityName: card.Store.CityName,
			StoreType: optionalText(card.Store.StoreType),
		},
	}
	if card.WeeklyQuantity > 0 {
		response.EstQtyPerWeek = &card.WeeklyQuantity
	}
	if card.MaxRetailPrice > 0 {
		response.MaxRetailPrice = &card.MaxRetailPrice
	}
	if card.HasOfferedCommission {
		response.OfferedCommissionBPS = &card.OfferedCommissionBPS
	}
	return response
}
