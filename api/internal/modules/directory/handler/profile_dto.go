package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	publicStorePath     = "/public/stores/:slug"
	publicConsignorPath = "/public/consignors/:slug"
)

type StoreProfileReader interface {
	Execute(ctx context.Context, slug string) (domain.StoreProfile, error)
}

type ConsignorProfileReader interface {
	Execute(ctx context.Context, slug string) (domain.ConsignorProfile, error)
}

type ProfileReaders struct {
	Store     StoreProfileReader
	Consignor ConsignorProfileReader
}

func PublicRoutes() []httpserver.PublicRoute {
	return []httpserver.PublicRoute{
		{Method: fiber.MethodGet, Path: httpserver.APIPrefix + publicStorePath},
		{Method: fiber.MethodGet, Path: httpserver.APIPrefix + publicConsignorPath},
	}
}

type placeResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type openNeedResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Category  string `json:"category"`
	ExpiresAt string `json:"expires_at"`
}

type publicStoreResponse struct {
	ID                 string                  `json:"id"`
	Slug               string                  `json:"slug"`
	Name               string                  `json:"name"`
	StoreType          *string                 `json:"store_type"`
	Description        string                  `json:"description"`
	Address            string                  `json:"address"`
	District           string                  `json:"district"`
	CityName           string                  `json:"city_name"`
	Location           *placeResponse          `json:"location"`
	OpeningHours       string                  `json:"opening_hours"`
	LogoURL            *string                 `json:"logo_url"`
	PhotoURLs          []string                `json:"photo_urls"`
	AcceptedCategories []categoryResponse      `json:"accepted_categories"`
	Terms              storeTermsResponse      `json:"terms"`
	OtherRequirements  string                  `json:"other_requirements"`
	Reputation         storeReputationResponse `json:"reputation"`
	JoinedOn           string                  `json:"joined_on"`
	OpenNeeds          []openNeedResponse      `json:"open_needs"`
}

type declaredLegalityResponse struct {
	BusinessIDNo         *string `json:"business_id_no"`
	DistributionPermitNo *string `json:"distribution_permit_no"`
	HalalCertNo          *string `json:"halal_cert_no"`
}

type publicProductCardResponse struct {
	SKU                  string  `json:"sku"`
	Name                 string  `json:"name"`
	Unit                 string  `json:"unit"`
	Category             string  `json:"category"`
	SuggestedRetailPrice int64   `json:"suggested_retail_price"`
	PhotoURL             *string `json:"photo_url"`
}

type publicConsignorResponse struct {
	ID             string                      `json:"id"`
	Slug           string                      `json:"slug"`
	Name           string                      `json:"name"`
	Description    string                      `json:"description"`
	District       string                      `json:"district"`
	CityName       string                      `json:"city_name"`
	Location       *placeResponse              `json:"location"`
	LogoURL        *string                     `json:"logo_url"`
	Legality       declaredLegalityResponse    `json:"legality"`
	ActivePartners int                         `json:"active_partners"`
	JoinedOn       string                      `json:"joined_on"`
	Products       []publicProductCardResponse `json:"products"`
}

func mountProfiles(api fiber.Router, profiles ProfileReaders, present presenter) {
	api.Get(publicStorePath, getPublic(profiles.Store.Execute, present.publicStore))
	api.Get(publicConsignorPath, getPublic(profiles.Consignor.Execute, present.publicConsignor))
}

func getPublic[T, R any](read func(ctx context.Context, slug string) (T, error), present func(T) R) fiber.Handler {
	return func(c fiber.Ctx) error {
		found, err := read(c.Context(), c.Params("slug"))
		if err != nil {
			return fmt.Errorf("read public page: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: present(found)})
	}
}

func (p presenter) publicStore(profile domain.StoreProfile) publicStoreResponse {
	return publicStoreResponse{
		ID: profile.ID, Slug: profile.Slug, Name: profile.Name, StoreType: optionalText(profile.StoreType),
		Description: profile.Description, Address: profile.Address, District: profile.District,
		CityName: profile.CityName, Location: placeOf(profile.Location), OpeningHours: profile.OpeningHours,
		LogoURL: p.fileURL(profile.LogoStorageKey), PhotoURLs: p.fileURLs(profile.PhotoStorageKeys),
		AcceptedCategories: categoriesOf(profile.AcceptedCategories), Terms: termsOf(profile.Terms),
		OtherRequirements: profile.OtherRequirements, Reputation: reputationOf(profile.Reputation),
		JoinedOn: businessDate(profile.JoinedAt), OpenNeeds: openNeedsOf(profile.OpenNeeds),
	}
}

func (p presenter) publicConsignor(profile domain.ConsignorProfile) publicConsignorResponse {
	products := make([]publicProductCardResponse, 0, len(profile.Products))
	for _, product := range profile.Products {
		products = append(products, publicProductCardResponse{
			SKU: product.SKU, Name: product.Name, Unit: product.Unit, Category: product.CategoryName,
			SuggestedRetailPrice: product.SuggestedRetailPrice, PhotoURL: p.fileURL(product.PhotoStorageKey),
		})
	}
	return publicConsignorResponse{
		ID: profile.ID, Slug: profile.Slug, Name: profile.Name, Description: profile.Description,
		District: profile.District, CityName: profile.CityName, Location: placeOf(profile.Location),
		LogoURL: p.fileURL(profile.LogoStorageKey),
		Legality: declaredLegalityResponse{
			BusinessIDNo:         optionalText(profile.Legality.BusinessIDNo),
			DistributionPermitNo: optionalText(profile.Legality.DistributionPermitNo),
			HalalCertNo:          optionalText(profile.Legality.HalalCertNo),
		},
		ActivePartners: profile.ActivePartners, JoinedOn: businessDate(profile.JoinedAt), Products: products,
	}
}

func (p presenter) fileURLs(storageKeys []string) []string {
	urls := make([]string, 0, len(storageKeys))
	for _, key := range storageKeys {
		if url := p.fileURL(key); url != nil {
			urls = append(urls, *url)
		}
	}
	return urls
}

func placeOf(place domain.Place) *placeResponse {
	if !place.IsKnown {
		return nil
	}
	return &placeResponse{Latitude: place.Point.Latitude, Longitude: place.Point.Longitude}
}

func openNeedsOf(needs []domain.OpenNeed) []openNeedResponse {
	responses := make([]openNeedResponse, 0, len(needs))
	for _, need := range needs {
		responses = append(responses, openNeedResponse{
			ID: need.ID, Title: need.Title, Category: need.CategoryName,
			ExpiresAt: need.ExpiresAt.UTC().Format(time.RFC3339),
		})
	}
	return responses
}
