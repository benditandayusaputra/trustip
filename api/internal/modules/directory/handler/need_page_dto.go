package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const publicNeedPath = "/public/needs/:id"

type NeedPageReader interface {
	Execute(ctx context.Context, needID string) (domain.PublicNeed, error)
}

type needPageStoreResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	StoreType   *string `json:"store_type"`
	District    string  `json:"district"`
	CityName    string  `json:"city_name"`
	ProfileSlug *string `json:"profile_slug"`
}

type publicNeedResponse struct {
	ID                   string                `json:"id"`
	Title                string                `json:"title"`
	Category             string                `json:"category"`
	Description          string                `json:"description"`
	EstQtyPerWeek        *int                  `json:"est_qty_per_week"`
	MaxRetailPrice       *int64                `json:"max_retail_price"`
	PreferredModel       *string               `json:"preferred_model"`
	OfferedCommissionBPS *int                  `json:"offered_commission_bps"`
	Status               string                `json:"status"`
	ExpiresAt            string                `json:"expires_at"`
	Store                needPageStoreResponse `json:"store"`
}

func getNeedPage(reader NeedPageReader) fiber.Handler {
	return func(c fiber.Ctx) error {
		needID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read need id: %w", err)
		}
		need, err := reader.Execute(c.Context(), needID)
		if err != nil {
			return fmt.Errorf("get need page: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: needPageOf(need)})
	}
}

func needPageOf(need domain.PublicNeed) publicNeedResponse {
	response := publicNeedResponse{
		ID: need.ID, Title: need.Title, Category: need.CategoryName, Description: need.Description,
		PreferredModel: optionalText(need.PreferredModel), Status: need.Status,
		ExpiresAt: need.ExpiresAt.UTC().Format(time.RFC3339),
		Store: needPageStoreResponse{
			ID: need.Store.ID, Name: need.Store.Name, StoreType: optionalText(need.Store.StoreType),
			District: need.Store.District,
			CityName: need.Store.CityName, ProfileSlug: optionalText(need.Store.ProfileSlug),
		},
	}
	if need.EstQtyPerWeek > 0 {
		response.EstQtyPerWeek = &need.EstQtyPerWeek
	}
	if need.MaxRetailPrice > 0 {
		response.MaxRetailPrice = &need.MaxRetailPrice
	}
	if need.OfferedCommission.IsGiven {
		response.OfferedCommissionBPS = &need.OfferedCommission.BPS
	}
	return response
}
