package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
)

const (
	maxRetailPrice   = 100_000_000
	minShelfLifeDays = 1
	maxShelfLifeDays = 730
	locationField    = "location"
)

type ConsignorSearchForm struct {
	Search           SearchForm
	MinPrice         string
	MaxPrice         string
	MinShelfLifeDays string
	HasPermit        bool
	HasHalal         bool
	CanSample        bool
	DeliversToViewer bool
	ConsignorID      string
}

type ConsignorFilters struct {
	CategoryIDs      []string
	Text             string
	MinPrice         Bound
	MaxPrice         Bound
	MinShelfLifeDays Bound
	HasPermit        bool
	HasHalal         bool
	CanSample        bool
	DeliversToViewer bool
	ConsignorID      string
}

type ConsignorSearch struct {
	Search
	Filters ConsignorFilters
}

func NewConsignorSearch(form ConsignorSearchForm) (ConsignorSearch, error) {
	search, problems := readSearch(form.Search)
	prices := field.Length{Max: maxRetailPrice}
	minPrice, minProblem := readBound("min_price", form.MinPrice, prices)
	maxPrice, maxProblem := readBound("max_price", form.MaxPrice, prices)
	shelfLife, shelfLifeProblem := readBound("min_shelf_life_days", form.MinShelfLifeDays,
		field.Length{Min: minShelfLifeDays, Max: maxShelfLifeDays})
	problems = append(problems, field.Collect(minProblem, maxProblem, shelfLifeProblem,
		findConsignorProblem(form.ConsignorID))...)
	if len(problems) > 0 {
		return ConsignorSearch{}, apperror.Validation(problems...)
	}
	return ConsignorSearch{Search: search, Filters: ConsignorFilters{
		CategoryIDs: search.CategoryIDs, Text: search.Text, MinPrice: minPrice, MaxPrice: maxPrice,
		MinShelfLifeDays: shelfLife, HasPermit: form.HasPermit, HasHalal: form.HasHalal, CanSample: form.CanSample,
		DeliversToViewer: form.DeliversToViewer, ConsignorID: form.ConsignorID,
	}}, nil
}

func findConsignorProblem(consignorID string) field.Problem {
	if consignorID == "" || idFormat.MatchString(consignorID) {
		return field.Problem{}
	}
	return field.Invalid("consignor_id", apperror.FieldUnknownChoice)
}

func (s ConsignorSearch) CheckViewer(viewer Viewer) error {
	if s.Filters.DeliversToViewer && !viewer.HasLocation {
		return apperror.Validation(apperror.InvalidField(locationField, apperror.FieldRequired))
	}
	return nil
}

type ConsignorQuery struct {
	Listing
	Filters        ConsignorFilters
	ViewerLocation Point
}

type Legality struct {
	HasDistributionPermit bool
	HasHalalCertificate   bool
	CanProvideSample      bool
}

type ConsignorTraits struct {
	Legality            Legality
	DeliveryRadiusKm    int
	PreferredStoreTypes []string
	Reputation          ConsignorReputation
	LastActiveAt        time.Time
}

type ConsignorCard struct {
	ID             string
	Slug           string
	Name           string
	CityName       string
	District       string
	LogoStorageKey string
	Distance       Distance
	Categories     []Category
	ProductCount   int
	Traits         ConsignorTraits
	CreatedAt      time.Time
	IsSelf         bool
	IsPartner      bool
	Match          Match
	HasMatch       bool
	Position       Position
}

func (c ConsignorCard) rankKey() rankKey {
	return rankKey{score: c.Match.Score, meters: c.Distance.Meters, id: c.ID}
}

func (c ConsignorCard) SortPosition() Position {
	return c.Position
}

type ProductConsignor struct {
	ID       string
	Slug     string
	Name     string
	CityName string
}

type ProductCard struct {
	ID                   string
	Name                 string
	SKU                  string
	Unit                 string
	Category             Category
	SuggestedRetailPrice int64
	ShelfLifeDays        int
	PhotoStorageKey      string
	Consignor            ProductConsignor
	Distance             Distance
	Traits               ConsignorTraits
	CreatedAt            time.Time
	IsSelf               bool
	IsPartner            bool
	Match                Match
	HasMatch             bool
	Position             Position
}

func (c ProductCard) rankKey() rankKey {
	return rankKey{score: c.Match.Score, meters: c.Distance.Meters, id: c.ID}
}

func (c ProductCard) SortPosition() Position {
	return c.Position
}

func (v Viewer) MatchConsignor(card ConsignorCard, now time.Time) ConsignorCard {
	if !v.IsStore() {
		return card
	}
	card.Match, card.HasMatch = v.scoreConsignor(consignorSide{card.Distance, card.Traits, card.Categories}, now), true
	return card
}

func (v Viewer) MatchProduct(card ProductCard, now time.Time) ProductCard {
	if !v.IsStore() {
		return card
	}
	side := consignorSide{distance: card.Distance, traits: card.Traits, categories: []Category{card.Category}}
	card.Match, card.HasMatch = v.scoreConsignor(side, now), true
	return card
}

type consignorSide struct {
	distance   Distance
	traits     ConsignorTraits
	categories []Category
}

func (v Viewer) scoreConsignor(side consignorSide, now time.Time) Match {
	return ScoreMatch(MatchFacts{
		Distance: side.distance, DeliveryRadiusKm: side.traits.DeliveryRadiusKm, ConsignorCategories: side.categories,
		AcceptedCategoryIDs: v.Store.AcceptedCategoryIDs, NeedCategoryIDs: v.Store.NeedCategoryIDs,
		PreferredStoreTypes: side.traits.PreferredStoreTypes, StoreType: v.Store.StoreType,
		Reputation: side.traits.Reputation, LastActiveAt: side.traits.LastActiveAt, Now: now,
	})
}
