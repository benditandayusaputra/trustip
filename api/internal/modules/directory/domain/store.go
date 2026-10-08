package domain

import (
	"cmp"
	"slices"
	"time"
)

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
		NeedCategoryIDs: []string{}, PreferredStoreTypes: v.Consignor.PreferredStoreTypes,
		StoreType: card.StoreType, Reputation: card.Reputation, LastActiveAt: card.LastActiveAt, Now: now,
	})
	card.HasMatch = true
	return card
}

func RankByMatch(cards []StoreCard) []StoreCard {
	ranked := slices.Clone(cards)
	slices.SortStableFunc(ranked, func(left, right StoreCard) int {
		return cmp.Or(
			cmp.Compare(right.Match.Score, left.Match.Score),
			cmp.Compare(left.Distance.Meters, right.Distance.Meters),
			cmp.Compare(left.ID, right.ID),
		)
	})
	return ranked
}

func PageAfter(ranked []StoreCard, afterID string, limit int) ([]StoreCard, bool) {
	start := 0
	if afterID != "" {
		start = slices.IndexFunc(ranked, func(card StoreCard) bool { return card.ID == afterID }) + 1
	}
	rest := ranked[start:]
	if len(rest) <= limit {
		return rest, false
	}
	return rest[:limit], true
}

type SearchContext struct {
	Origin            Origin
	Sort              SortOrder
	SelectedRadiusKm  int
	EffectiveRadiusKm int
	CategoryIDs       []string
}

func (c SearchContext) IsExpanded() bool {
	return c.Origin.IsKnown && c.EffectiveRadiusKm != c.SelectedRadiusKm
}

func (s StoreSearch) SortFor(origin Origin) SortOrder {
	if !origin.IsKnown && s.Sort == Nearest {
		return ByCity
	}
	return s.Sort
}
