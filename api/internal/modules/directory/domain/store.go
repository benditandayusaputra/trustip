package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/storetype"
)

const maxCommissionBPS = 5000

func pricingModelChoices() []string {
	return []string{"FIXED_DEPOSIT", "COMMISSION"}
}

type StoreSearchForm struct {
	Search                   SearchForm
	StoreType                string
	PricingModel             string
	MaxCommissionBPS         string
	WithoutPermitRequirement bool
	WithoutHalalRequirement  bool
	HasOpenNeed              bool
}

type StoreFilters struct {
	StoreType        string
	CategoryIDs      []string
	PricingModel     string
	HasMaxCommission bool
	MaxCommissionBPS int
	WithoutPermit    bool
	WithoutHalal     bool
	HasOpenNeed      bool
	Text             string
}

type StoreSearch struct {
	Search
	Filters StoreFilters
}

func NewStoreSearch(form StoreSearchForm) (StoreSearch, error) {
	search, problems := readSearch(form.Search)
	commission, commissionProblem := readBound("max_commission_bps", form.MaxCommissionBPS,
		field.Length{Max: maxCommissionBPS})
	problems = append(problems, field.Collect(
		commissionProblem,
		field.CheckChoice("store_type", form.StoreType, storetype.All()),
		field.CheckChoice("pricing_model", form.PricingModel, pricingModelChoices()),
	)...)
	if len(problems) > 0 {
		return StoreSearch{}, apperror.Validation(problems...)
	}
	return StoreSearch{Search: search, Filters: StoreFilters{
		StoreType: form.StoreType, CategoryIDs: search.CategoryIDs, PricingModel: form.PricingModel,
		HasMaxCommission: commission.IsGiven, MaxCommissionBPS: commission.Value,
		WithoutPermit: form.WithoutPermitRequirement, WithoutHalal: form.WithoutHalalRequirement,
		HasOpenNeed: form.HasOpenNeed, Text: search.Text,
	}}, nil
}

type StoreQuery struct {
	Listing
	Filters StoreFilters
}

type CommissionRange struct {
	MinBPS  int
	MaxBPS  int
	IsKnown bool
}

type StoreTerms struct {
	AcceptsFixedDeposit        bool
	AcceptsCommission          bool
	Commission                 CommissionRange
	RequiresDistributionPermit bool
	RequiresHalal              bool
	RequiresSample             bool
}

type StoreCard struct {
	ID                 string
	Slug               string
	Name               string
	CityName           string
	District           string
	StoreType          string
	LogoStorageKey     string
	Distance           Distance
	AcceptedCategories []Category
	NeedCategoryIDs    []string
	Terms              StoreTerms
	Reputation         StoreReputation
	LastActiveAt       time.Time
	CreatedAt          time.Time
	IsSelf             bool
	IsPartner          bool
	Match              Match
	HasMatch           bool
	Position           Position
}

func (c StoreCard) rankKey() rankKey {
	return rankKey{score: c.Match.Score, meters: c.Distance.Meters, id: c.ID}
}

func (c StoreCard) SortPosition() Position {
	return c.Position
}

func (c StoreCard) acceptedCategoryIDs() []string {
	ids := make([]string, 0, len(c.AcceptedCategories))
	for _, category := range c.AcceptedCategories {
		ids = append(ids, category.ID)
	}
	return ids
}

func (v Viewer) MatchStore(card StoreCard, now time.Time) StoreCard {
	if !v.IsConsignor() {
		return card
	}
	card.Match = ScoreMatch(MatchFacts{
		Distance: card.Distance, DeliveryRadiusKm: v.Consignor.DeliveryRadiusKm,
		ConsignorCategories: v.Consignor.Categories, AcceptedCategoryIDs: card.acceptedCategoryIDs(),
		NeedCategoryIDs: card.NeedCategoryIDs, PreferredStoreTypes: v.Consignor.PreferredStoreTypes,
		StoreType: card.StoreType, Reputation: card.Reputation, LastActiveAt: card.LastActiveAt, Now: now,
	})
	card.HasMatch = true
	return card
}
