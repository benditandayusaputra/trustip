package domain

import "github.com/benditandayusaputra/trustip/api/internal/kernel/actor"

const fullPercent = 100

type Requirement string

const (
	RequireName             Requirement = "NAME"
	RequireCity             Requirement = "CITY"
	RequireLocation         Requirement = "LOCATION"
	RequireStoreType        Requirement = "STORE_TYPE"
	RequireAcceptedCategory Requirement = "ACCEPTED_CATEGORY"
	RequireActiveProduct    Requirement = "ACTIVE_PRODUCT"
	RecommendDescription    Requirement = "DESCRIPTION"
	RecommendLogo           Requirement = "LOGO"
	RecommendStorePhoto     Requirement = "STORE_PHOTO"
	RecommendPaymentInfo    Requirement = "PAYMENT_INFO"
)

type ListingFacts struct {
	Type               actor.OrganizationType
	IsListed           bool
	HasName            bool
	HasCity            bool
	HasLocation        bool
	HasDescription     bool
	HasLogo            bool
	HasStoreType       bool
	AcceptedCategories int
	Photos             int
	HasPaymentInfo     bool
	ActiveProducts     int
}

type CompletenessItem struct {
	Requirement Requirement
	IsRequired  bool
	IsDone      bool
}

type Completeness struct {
	Items []CompletenessItem
}

func CompletenessOf(facts ListingFacts) Completeness {
	items := []CompletenessItem{
		{RequireName, true, facts.HasName},
		{RequireCity, true, facts.HasCity},
		{RequireLocation, true, facts.HasLocation},
	}
	items = append(items, typeItems(facts)...)
	return Completeness{Items: append(items,
		CompletenessItem{RecommendDescription, false, facts.HasDescription},
		CompletenessItem{RecommendLogo, false, facts.HasLogo},
	)}
}

func typeItems(facts ListingFacts) []CompletenessItem {
	if facts.Type == actor.Store {
		return []CompletenessItem{
			{RequireStoreType, true, facts.HasStoreType},
			{RequireAcceptedCategory, true, facts.AcceptedCategories > 0},
			{RecommendStorePhoto, false, facts.Photos > 0},
		}
	}
	return []CompletenessItem{
		{RequireActiveProduct, true, facts.ActiveProducts > 0},
		{RecommendPaymentInfo, false, facts.HasPaymentInfo},
	}
}

func (c Completeness) MissingRequired() []Requirement {
	missing := []Requirement{}
	for _, item := range c.Items {
		if item.IsRequired && !item.IsDone {
			missing = append(missing, item.Requirement)
		}
	}
	return missing
}

func (c Completeness) CanBeListed() bool {
	return len(c.MissingRequired()) == 0
}

func (c Completeness) Percent() int {
	if len(c.Items) == 0 {
		return 0
	}
	done := 0
	for _, item := range c.Items {
		if item.IsDone {
			done++
		}
	}
	return done * fullPercent / len(c.Items)
}
