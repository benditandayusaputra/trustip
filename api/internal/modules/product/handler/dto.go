package handler

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type productRequest struct {
	SKU                   string  `json:"sku"`
	Name                  string  `json:"name"`
	CategoryID            string  `json:"category_id"`
	Unit                  string  `json:"unit"`
	Description           string  `json:"description"`
	Composition           string  `json:"composition"`
	NetContent            string  `json:"net_content"`
	PhotoFileID           *string `json:"photo_file_id"`
	SuggestedRetailPrice  int64   `json:"suggested_retail_price"`
	SuggestedDepositPrice int64   `json:"suggested_deposit_price"`
	ShelfLifeDays         *int    `json:"shelf_life_days"`
}

type imageRef struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type productResponse struct {
	ID                    string    `json:"id"`
	SKU                   string    `json:"sku"`
	Name                  string    `json:"name"`
	CategoryID            string    `json:"category_id"`
	CategoryName          string    `json:"category_name"`
	Unit                  string    `json:"unit"`
	Description           string    `json:"description"`
	Composition           string    `json:"composition"`
	NetContent            string    `json:"net_content"`
	Photo                 *imageRef `json:"photo"`
	SuggestedRetailPrice  int64     `json:"suggested_retail_price"`
	SuggestedDepositPrice int64     `json:"suggested_deposit_price"`
	ShelfLifeDays         *int      `json:"shelf_life_days"`
	Status                string    `json:"status"`
	CreatedAt             string    `json:"created_at"`
	UpdatedAt             string    `json:"updated_at"`
}

func readForm(c fiber.Ctx) (domain.ProductForm, error) {
	var request productRequest
	if err := httpserver.DecodeJSON(c, &request); err != nil {
		return domain.ProductForm{}, fmt.Errorf("read product: %w", err)
	}
	return request.toForm()
}

func (r productRequest) toForm() (domain.ProductForm, error) {
	inputs := []httpserver.TextInput{
		httpserver.LineInput("sku", r.SKU),
		httpserver.LineInput("name", r.Name),
		httpserver.LineInput("unit", r.Unit),
		httpserver.MultilineInput("description", r.Description),
		httpserver.MultilineInput("composition", r.Composition),
		httpserver.LineInput("net_content", r.NetContent),
	}
	if err := httpserver.RejectUncleanInputs(inputs...); err != nil {
		return domain.ProductForm{}, fmt.Errorf("check product text: %w", err)
	}
	form := domain.ProductForm{
		SKU: inputs[0].Value, Name: inputs[1].Value, CategoryID: r.CategoryID, Unit: inputs[2].Value,
		Description: inputs[3].Value, Composition: inputs[4].Value, NetContent: inputs[5].Value,
		Prices: domain.SuggestedPrices{Retail: r.SuggestedRetailPrice, Deposit: r.SuggestedDepositPrice},
	}
	if r.PhotoFileID != nil {
		form.PhotoFileID = *r.PhotoFileID
	}
	if r.ShelfLifeDays != nil {
		form.ShelfLife = domain.ShelfLife{Given: true, Days: *r.ShelfLifeDays}
	}
	return form, nil
}

type presenter struct {
	filesBaseURL string
}

func (p presenter) product(product domain.Product) productResponse {
	details := product.Details
	response := productResponse{
		ID: product.ID, SKU: product.SKU, Name: details.Name, CategoryID: details.CategoryID,
		CategoryName: product.CategoryName, Unit: details.Unit, Description: details.Description,
		Composition: details.Composition, NetContent: details.NetContent,
		SuggestedRetailPrice: details.Prices.Retail.Int64(), SuggestedDepositPrice: details.Prices.Deposit.Int64(),
		Status: string(product.Status()), CreatedAt: product.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: product.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if product.Photo.ID != "" {
		response.Photo = &imageRef{ID: product.Photo.ID, URL: p.filesBaseURL + "/" + product.Photo.StorageKey}
	}
	if details.ShelfLife.Given {
		days := details.ShelfLife.Days
		response.ShelfLifeDays = &days
	}
	return response
}
