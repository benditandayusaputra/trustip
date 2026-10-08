package domain

import (
	"regexp"
	"slices"
	"strconv"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/storetype"
)

const (
	DefaultRadiusKm           = 10
	EveryDistance             = 0
	MinResultsBeforeExpanding = 5
	MaxSearchLength           = 80
	MaxCategoryFilters        = 12
	MatchCandidateLimit       = 200
	maxCommissionBPS          = 5000
	everyRadiusChoice         = "all"
	anyCategoryChoice         = "any"
)

var idFormat = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func RadiusSteps() []int {
	return []int{2, 5, 10, 25}
}

type SortOrder string

const (
	Nearest      SortOrder = "NEAREST"
	BestMatch    SortOrder = "BEST_MATCH"
	ByReputation SortOrder = "REPUTATION"
	Newest       SortOrder = "NEWEST"
	ByCity       SortOrder = "CITY"
)

func sortChoices() []string {
	return []string{string(Nearest), string(BestMatch), string(ByReputation), string(Newest)}
}

func pricingModelChoices() []string {
	return []string{"FIXED_DEPOSIT", "COMMISSION"}
}

type StoreSearchForm struct {
	Origin                   OriginForm
	Radius                   string
	Sort                     string
	StoreType                string
	CategoryIDs              []string
	PricingModel             string
	MaxCommissionBPS         string
	WithoutPermitRequirement bool
	WithoutHalalRequirement  bool
	Text                     string
}

type StoreFilters struct {
	StoreType        string
	CategoryIDs      []string
	PricingModel     string
	HasMaxCommission bool
	MaxCommissionBPS int
	WithoutPermit    bool
	WithoutHalal     bool
	Text             string
}

type StoreSearch struct {
	Origin            OriginRequest
	RadiusKm          int
	Sort              SortOrder
	Filters           StoreFilters
	UsesOwnCategories bool
}

func NewStoreSearch(form StoreSearchForm) (StoreSearch, error) {
	origin, originProblems := readOrigin(form.Origin)
	radius, radiusProblem := readRadius(form.Radius)
	commission, commissionProblem := readMaxCommission(form.MaxCommissionBPS)
	categories, categoryProblems := readCategories(form.CategoryIDs)
	problems := append(originProblems, field.Collect(
		radiusProblem, commissionProblem,
		field.CheckChoice("sort", form.Sort, sortChoices()),
		field.CheckChoice("store_type", form.StoreType, storetype.All()),
		field.CheckChoice("pricing_model", form.PricingModel, pricingModelChoices()),
		field.CheckLength("q", form.Text, field.Length{Max: MaxSearchLength}),
	)...)
	problems = append(problems, categoryProblems...)
	if len(problems) > 0 {
		return StoreSearch{}, apperror.Validation(problems...)
	}
	return StoreSearch{
		Origin: origin, RadiusKm: radius, Sort: sortOf(form.Sort), UsesOwnCategories: form.CategoryIDs == nil,
		Filters: StoreFilters{
			StoreType: form.StoreType, CategoryIDs: categories, PricingModel: form.PricingModel,
			HasMaxCommission: commission.isGiven, MaxCommissionBPS: commission.basisPoints,
			WithoutPermit: form.WithoutPermitRequirement, WithoutHalal: form.WithoutHalalRequirement, Text: form.Text,
		},
	}, nil
}

func sortOf(raw string) SortOrder {
	if raw == "" {
		return Nearest
	}
	return SortOrder(raw)
}

func readRadius(raw string) (int, field.Problem) {
	if raw == "" {
		return DefaultRadiusKm, field.Problem{}
	}
	if raw == everyRadiusChoice {
		return EveryDistance, field.Problem{}
	}
	radius, err := strconv.Atoi(raw)
	if err != nil || !slices.Contains(RadiusSteps(), radius) {
		return 0, field.Invalid("radius_km", apperror.FieldUnknownChoice)
	}
	return radius, field.Problem{}
}

type commissionLimit struct {
	basisPoints int
	isGiven     bool
}

func readMaxCommission(raw string) (commissionLimit, field.Problem) {
	if raw == "" {
		return commissionLimit{}, field.Problem{}
	}
	commission, err := strconv.Atoi(raw)
	if err != nil || commission < 0 || commission > maxCommissionBPS {
		return commissionLimit{}, field.Invalid("max_commission_bps", apperror.FieldOutOfRange)
	}
	return commissionLimit{basisPoints: commission, isGiven: true}, field.Problem{}
}

func readCategories(raw []string) ([]string, []apperror.FieldError) {
	if len(raw) == 1 && raw[0] == anyCategoryChoice {
		return []string{}, nil
	}
	if len(raw) > MaxCategoryFilters {
		return nil, field.Collect(field.Invalid("category_id", apperror.FieldTooMany,
			field.Limit("max", MaxCategoryFilters)))
	}
	for _, categoryID := range raw {
		if !idFormat.MatchString(categoryID) {
			return nil, field.Collect(field.Invalid("category_id", apperror.FieldUnknownChoice))
		}
	}
	return slices.Clone(raw), nil
}

func EffectiveRadius(selectedKm int, storesWithinKm map[int]int) int {
	if selectedKm == EveryDistance {
		return EveryDistance
	}
	for _, step := range RadiusSteps() {
		if step >= selectedKm && storesWithinKm[step] >= MinResultsBeforeExpanding {
			return step
		}
	}
	return EveryDistance
}

type ViewerKey struct {
	OrganizationID string
	CityCode       string
}

type StoreQuery struct {
	ViewerID string
	Origin   Origin
	RadiusKm int
	Filters  StoreFilters
	Sort     SortOrder
	After    Position
	Limit    int
}
