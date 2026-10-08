package handler

import "github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"

type categoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type storeTermsResponse struct {
	AcceptsFixedDeposit        bool `json:"accepts_fixed_deposit"`
	AcceptsCommission          bool `json:"accepts_commission"`
	CommissionMinBPS           *int `json:"commission_min_bps"`
	CommissionMaxBPS           *int `json:"commission_max_bps"`
	RequiresDistributionPermit bool `json:"requires_distribution_permit"`
	RequiresHalal              bool `json:"requires_halal"`
	RequiresSample             bool `json:"requires_sample"`
}

type storeReputationResponse struct {
	IsNew     bool `json:"is_new"`
	OnTimeBPS *int `json:"on_time_bps"`
}

type reasonResponse struct {
	Code         string  `json:"code"`
	DistanceM    *int    `json:"distance_m,omitempty"`
	CategoryName *string `json:"category,omitempty"`
	OnTimeBPS    *int    `json:"on_time_bps,omitempty"`
}

type matchResponse struct {
	Score   int              `json:"score"`
	Label   *string          `json:"label"`
	Reasons []reasonResponse `json:"reasons"`
}

type storeCardResponse struct {
	ID                 string                  `json:"id"`
	Slug               string                  `json:"slug"`
	Name               string                  `json:"name"`
	CityName           string                  `json:"city_name"`
	District           string                  `json:"district"`
	StoreType          *string                 `json:"store_type"`
	LogoURL            *string                 `json:"logo_url"`
	DistanceM          *int                    `json:"distance_m"`
	AcceptedCategories []categoryResponse      `json:"accepted_categories"`
	Terms              storeTermsResponse      `json:"terms"`
	Reputation         storeReputationResponse `json:"reputation"`
	IsSelf             bool                    `json:"is_self"`
	IsPartner          bool                    `json:"is_partner"`
	Match              *matchResponse          `json:"match"`
}

type originResponse struct {
	Kind    string `json:"kind"`
	IsKnown bool   `json:"is_known"`
}

type searchContextResponse struct {
	Origin            originResponse `json:"origin"`
	Sort              string         `json:"sort"`
	SelectedRadiusKm  *int           `json:"selected_radius_km"`
	EffectiveRadiusKm *int           `json:"effective_radius_km"`
	IsExpanded        bool           `json:"is_expanded"`
	CategoryIDs       []string       `json:"category_ids"`
}

type presenter struct {
	filesBaseURL string
}

func (p presenter) storeCard(card domain.StoreCard) storeCardResponse {
	response := storeCardResponse{
		ID: card.ID, Slug: card.Slug, Name: card.Name, CityName: card.CityName, District: card.District,
		StoreType: optionalText(card.StoreType), AcceptedCategories: categoriesOf(card.AcceptedCategories),
		Terms: termsOf(card.Terms), Reputation: reputationOf(card.Reputation), IsSelf: card.IsSelf,
		IsPartner: card.IsPartner,
	}
	response.LogoURL = p.fileURL(card.LogoStorageKey)
	response.DistanceM = distanceOf(card.Distance)
	response.Match = optionalMatch(card)
	return response
}

func (p presenter) fileURL(storageKey string) *string {
	if storageKey == "" {
		return nil
	}
	url := p.filesBaseURL + "/" + storageKey
	return &url
}

func distanceOf(distance domain.Distance) *int {
	if !distance.IsKnown {
		return nil
	}
	return &distance.Meters
}

type scored interface {
	ScoredMatch() (domain.Match, bool)
}

func optionalMatch(card scored) *matchResponse {
	match, hasMatch := card.ScoredMatch()
	if !hasMatch {
		return nil
	}
	return matchOf(match)
}

func categoriesOf(categories []domain.Category) []categoryResponse {
	responses := make([]categoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, categoryResponse(category))
	}
	return responses
}

func termsOf(terms domain.StoreTerms) storeTermsResponse {
	response := storeTermsResponse{
		AcceptsFixedDeposit: terms.AcceptsFixedDeposit, AcceptsCommission: terms.AcceptsCommission,
		RequiresDistributionPermit: terms.RequiresDistributionPermit, RequiresHalal: terms.RequiresHalal,
		RequiresSample: terms.RequiresSample,
	}
	if terms.Commission.IsKnown {
		response.CommissionMinBPS, response.CommissionMaxBPS = &terms.Commission.MinBPS, &terms.Commission.MaxBPS
	}
	return response
}

func reputationOf(reputation domain.StoreReputation) storeReputationResponse {
	if reputation.IsNew() {
		return storeReputationResponse{IsNew: true}
	}
	return storeReputationResponse{OnTimeBPS: &reputation.OnTimeBPS}
}

func matchOf(match domain.Match) *matchResponse {
	reasons := make([]reasonResponse, 0, len(match.Reasons))
	for _, reason := range match.Reasons {
		reasons = append(reasons, reasonOf(reason))
	}
	return &matchResponse{Score: match.Score, Label: optionalText(string(match.Label)), Reasons: reasons}
}

func reasonOf(reason domain.Reason) reasonResponse {
	response := reasonResponse{Code: string(reason.Code)}
	switch reason.Code {
	case domain.ReasonDistance:
		response.DistanceM = &reason.DistanceMeters
	case domain.ReasonNeedsCategory, domain.ReasonAcceptsCategory:
		response.CategoryName = &reason.CategoryName
	case domain.ReasonOnTimePayment:
		response.OnTimeBPS = &reason.OnTimeBPS
	case domain.ReasonEstablished, domain.ReasonPreferredType, domain.ReasonActiveRecently:
	}
	return response
}

func searchContextOf(searched domain.SearchContext) searchContextResponse {
	return searchContextResponse{
		Origin: originResponse{Kind: string(searched.Origin.Kind), IsKnown: searched.Origin.IsKnown},
		Sort:   string(searched.Sort), SelectedRadiusKm: optionalRadius(searched.SelectedRadiusKm),
		EffectiveRadiusKm: optionalRadius(searched.EffectiveRadiusKm), IsExpanded: searched.IsExpanded(),
		CategoryIDs: append([]string{}, searched.CategoryIDs...),
	}
}

func optionalRadius(radiusKm int) *int {
	if radiusKm == domain.EveryDistance {
		return nil
	}
	return &radiusKm
}

func optionalText(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
