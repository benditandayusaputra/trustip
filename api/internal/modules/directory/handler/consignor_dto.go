package handler

import "github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"

type legalityResponse struct {
	HasDistributionPermit bool `json:"has_distribution_permit"`
	HasHalalCertificate   bool `json:"has_halal_certificate"`
	CanProvideSample      bool `json:"can_provide_sample"`
}

type consignorReputationResponse struct {
	IsEstablished  bool `json:"is_established"`
	ActivePartners int  `json:"active_partners"`
}

type consignorTraitsResponse struct {
	Legality         legalityResponse            `json:"legality"`
	DeliveryRadiusKm *int                        `json:"delivery_radius_km"`
	Reputation       consignorReputationResponse `json:"reputation"`
}

type consignorCardResponse struct {
	ID           string                  `json:"id"`
	Slug         string                  `json:"slug"`
	Name         string                  `json:"name"`
	CityName     string                  `json:"city_name"`
	District     string                  `json:"district"`
	LogoURL      *string                 `json:"logo_url"`
	DistanceM    *int                    `json:"distance_m"`
	Categories   []categoryResponse      `json:"categories"`
	ProductCount int                     `json:"product_count"`
	Traits       consignorTraitsResponse `json:"traits"`
	IsSelf       bool                    `json:"is_self"`
	IsPartner    bool                    `json:"is_partner"`
	Match        *matchResponse          `json:"match"`
}

type productConsignorResponse struct {
	ID       string `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	CityName string `json:"city_name"`
}

type productCardResponse struct {
	ID                   string                   `json:"id"`
	Name                 string                   `json:"name"`
	SKU                  string                   `json:"sku"`
	Unit                 string                   `json:"unit"`
	Category             categoryResponse         `json:"category"`
	SuggestedRetailPrice int64                    `json:"suggested_retail_price"`
	ShelfLifeDays        *int                     `json:"shelf_life_days"`
	PhotoURL             *string                  `json:"photo_url"`
	Consignor            productConsignorResponse `json:"consignor"`
	DistanceM            *int                     `json:"distance_m"`
	Traits               consignorTraitsResponse  `json:"traits"`
	IsSelf               bool                     `json:"is_self"`
	IsPartner            bool                     `json:"is_partner"`
	Match                *matchResponse           `json:"match"`
}

func (p presenter) consignorCard(card domain.ConsignorCard) consignorCardResponse {
	return consignorCardResponse{
		ID: card.ID, Slug: card.Slug, Name: card.Name, CityName: card.CityName, District: card.District,
		LogoURL: p.fileURL(card.LogoStorageKey), DistanceM: distanceOf(card.Distance),
		Categories: categoriesOf(card.Categories), ProductCount: card.ProductCount, Traits: traitsOf(card.Traits),
		IsSelf: card.IsSelf, IsPartner: card.IsPartner, Match: optionalMatch(card),
	}
}

func (p presenter) productCard(card domain.ProductCard) productCardResponse {
	response := productCardResponse{
		ID: card.ID, Name: card.Name, SKU: card.SKU, Unit: card.Unit, Category: categoryResponse(card.Category),
		SuggestedRetailPrice: card.SuggestedRetailPrice, PhotoURL: p.fileURL(card.PhotoStorageKey),
		Consignor: productConsignorResponse(card.Consignor),
		DistanceM: distanceOf(card.Distance), Traits: traitsOf(card.Traits), IsSelf: card.IsSelf,
		IsPartner: card.IsPartner, Match: optionalMatch(card),
	}
	if card.ShelfLifeDays > 0 {
		response.ShelfLifeDays = &card.ShelfLifeDays
	}
	return response
}

func traitsOf(traits domain.ConsignorTraits) consignorTraitsResponse {
	response := consignorTraitsResponse{
		Legality: legalityResponse(traits.Legality),
		Reputation: consignorReputationResponse{
			IsEstablished: traits.Reputation.IsEstablished(), ActivePartners: traits.Reputation.ActivePartners,
		},
	}
	if traits.DeliveryRadiusKm > 0 {
		response.DeliveryRadiusKm = &traits.DeliveryRadiusKm
	}
	return response
}
