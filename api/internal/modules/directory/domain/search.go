package domain

import (
	"regexp"
	"slices"
	"strconv"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
)

const (
	DefaultRadiusKm           = 10
	EveryDistance             = 0
	MinResultsBeforeExpanding = 5
	MaxSearchLength           = 80
	MaxCategoryFilters        = 12
	MatchCandidateLimit       = 200
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

type SearchForm struct {
	Origin      OriginForm
	Radius      string
	Sort        string
	CategoryIDs []string
	Text        string
}

type Search struct {
	Origin            OriginRequest
	RadiusKm          int
	Sort              SortOrder
	CategoryIDs       []string
	UsesOwnCategories bool
	Text              string
}

func readSearch(form SearchForm) (Search, []apperror.FieldError) {
	origin, problems := readOrigin(form.Origin)
	radius, radiusProblem := readRadius(form.Radius)
	categories, categoryProblems := readCategories(form.CategoryIDs)
	problems = append(problems, field.Collect(
		radiusProblem,
		field.CheckChoice("sort", form.Sort, sortChoices()),
		field.CheckLength("q", form.Text, field.Length{Max: MaxSearchLength}),
	)...)
	return Search{
		Origin: origin, RadiusKm: radius, Sort: sortOf(form.Sort), CategoryIDs: categories,
		UsesOwnCategories: form.CategoryIDs == nil, Text: form.Text,
	}, append(problems, categoryProblems...)
}

func (s Search) SortFor(origin Origin) SortOrder {
	if !origin.IsKnown && s.Sort == Nearest {
		return ByCity
	}
	return s.Sort
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
	return append([]string{}, raw...), nil
}

type Bound struct {
	Value   int
	IsGiven bool
}

func readBound(name, raw string, allowed field.Length) (Bound, field.Problem) {
	if raw == "" {
		return Bound{}, field.Problem{}
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < allowed.Min || value > allowed.Max {
		return Bound{}, field.Invalid(name, apperror.FieldOutOfRange)
	}
	return Bound{Value: value, IsGiven: true}, field.Problem{}
}

func EffectiveRadius(selectedKm int, resultsWithinKm map[int]int) int {
	if selectedKm == EveryDistance {
		return EveryDistance
	}
	for _, step := range RadiusSteps() {
		if step >= selectedKm && resultsWithinKm[step] >= MinResultsBeforeExpanding {
			return step
		}
	}
	return EveryDistance
}

type ViewerKey struct {
	OrganizationID string
	CityCode       string
}

type Listing struct {
	ViewerID string
	Origin   Origin
	RadiusKm int
	Sort     SortOrder
	After    Position
	Limit    int
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
