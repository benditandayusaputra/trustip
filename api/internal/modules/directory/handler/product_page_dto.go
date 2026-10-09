package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const publicProductPath = "/public/products/:slug/:sku"

type ProductPageReader interface {
	Execute(ctx context.Context, form domain.ProductPageForm) (domain.ProductPage, error)
}

type productMakerResponse struct {
	Name        string                   `json:"name"`
	District    string                   `json:"district"`
	CityName    string                   `json:"city_name"`
	ProfileSlug *string                  `json:"profile_slug"`
	Legality    declaredLegalityResponse `json:"legality"`
}

type availableStoreResponse struct {
	Name        string         `json:"name"`
	Address     string         `json:"address"`
	District    string         `json:"district"`
	CityName    string         `json:"city_name"`
	ProfileSlug *string        `json:"profile_slug"`
	Location    *placeResponse `json:"location"`
	DistanceM   *int           `json:"distance_m"`
}

type publicProductResponse struct {
	ID                   string                   `json:"id"`
	SKU                  string                   `json:"sku"`
	Name                 string                   `json:"name"`
	Unit                 string                   `json:"unit"`
	Category             string                   `json:"category"`
	Description          string                   `json:"description"`
	Composition          string                   `json:"composition"`
	NetContent           string                   `json:"net_content"`
	SuggestedRetailPrice int64                    `json:"suggested_retail_price"`
	PhotoURL             *string                  `json:"photo_url"`
	Maker                productMakerResponse     `json:"maker"`
	Stores               []availableStoreResponse `json:"stores"`
}

func getProductPage(reader ProductPageReader, present presenter) fiber.Handler {
	return func(c fiber.Ctx) error {
		page, err := reader.Execute(c.Context(), domain.ProductPageForm{
			ConsignorSlug: c.Params("slug"), SKU: c.Params("sku"), Latitude: c.Query("lat"),
			Longitude: c.Query("lng"),
		})
		if err != nil {
			return fmt.Errorf("get product page: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: present.productPage(page)})
	}
}

func (p presenter) productPage(page domain.ProductPage) publicProductResponse {
	product := page.Product
	stores := make([]availableStoreResponse, 0, len(page.Stores))
	for _, store := range page.Stores {
		stores = append(stores, availableStoreResponse{
			Name: store.Name, Address: store.Address, District: store.District, CityName: store.CityName,
			ProfileSlug: optionalText(store.ProfileSlug), Location: placeOf(store.Location),
			DistanceM: distanceOf(store.Distance),
		})
	}
	return publicProductResponse{
		ID: product.ID, SKU: product.SKU, Name: product.Name, Unit: product.Unit, Category: product.CategoryName,
		Description: product.Description, Composition: product.Composition, NetContent: product.NetContent,
		SuggestedRetailPrice: product.SuggestedRetailPrice, PhotoURL: p.fileURL(product.PhotoStorageKey),
		Maker: productMakerResponse{
			Name: page.Maker.Name, District: page.Maker.District, CityName: page.Maker.CityName,
			ProfileSlug: optionalText(page.Maker.ProfileSlug),
			Legality:    legalityOf(page.Maker.Legality),
		},
		Stores: stores,
	}
}
