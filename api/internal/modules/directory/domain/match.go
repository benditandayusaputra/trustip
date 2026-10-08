package domain

import (
	"slices"
	"time"
)

const (
	distanceMaxPoints      = 40
	defaultReachKm         = 25
	metersPerKm            = 1000
	needCategoryPoints     = 30
	acceptedCategoryPoints = 20
	preferredTypePoints    = 10
	noPreferencePoints     = 5
	activeRecentlyPoints   = 10
	activeThisMonthPoints  = 5
	activeRecentlyWithin   = 14 * 24 * time.Hour
	activeThisMonthWithin  = 30 * 24 * time.Hour
	veryGoodMatchFrom      = 75
	goodMatchFrom          = 50
	maxReasons             = 3
	halves                 = 2
)

type Label string

const (
	NoLabel       Label = ""
	GoodMatch     Label = "GOOD"
	VeryGoodMatch Label = "VERY_GOOD"
)

type ReasonCode string

const (
	ReasonDistance        ReasonCode = "DISTANCE"
	ReasonNeedsCategory   ReasonCode = "NEEDS_CATEGORY"
	ReasonAcceptsCategory ReasonCode = "ACCEPTS_CATEGORY"
	ReasonOnTimePayment   ReasonCode = "ON_TIME_PAYMENT"
	ReasonEstablished     ReasonCode = "ESTABLISHED"
	ReasonPreferredType   ReasonCode = "PREFERRED_STORE_TYPE"
	ReasonActiveRecently  ReasonCode = "ACTIVE_RECENTLY"
)

type Reason struct {
	Code           ReasonCode
	DistanceMeters int
	CategoryName   string
	OnTimeBPS      int
}

type Category struct {
	ID   string
	Name string
}

type Distance struct {
	Meters  int
	IsKnown bool
}

func KnownDistance(meters int) Distance {
	return Distance{Meters: meters, IsKnown: true}
}

type MatchFacts struct {
	Distance            Distance
	DeliveryRadiusKm    int
	ConsignorCategories []Category
	AcceptedCategoryIDs []string
	NeedCategoryIDs     []string
	PreferredStoreTypes []string
	StoreType           string
	Reputation          Reputation
	LastActiveAt        time.Time
	Now                 time.Time
}

type Match struct {
	Score   int
	Label   Label
	Reasons []Reason
}

type component struct {
	points int
	reason Reason
	isWon  bool
}

func ScoreMatch(facts MatchFacts) Match {
	components := []component{
		distanceComponent(facts.Distance, facts.DeliveryRadiusKm),
		categoryComponent(facts),
		facts.Reputation.component(),
		storeTypeComponent(facts.PreferredStoreTypes, facts.StoreType),
		activityComponent(facts.LastActiveAt, facts.Now),
	}
	match := Match{Reasons: []Reason{}}
	for _, part := range components {
		match.Score += part.points
		if part.isWon && len(match.Reasons) < maxReasons {
			match.Reasons = append(match.Reasons, part.reason)
		}
	}
	match.Label = labelOf(match.Score)
	return match
}

func labelOf(score int) Label {
	if score >= veryGoodMatchFrom {
		return VeryGoodMatch
	}
	if score >= goodMatchFrom {
		return GoodMatch
	}
	return NoLabel
}

func distanceComponent(distance Distance, deliveryRadiusKm int) component {
	if !distance.IsKnown {
		return component{}
	}
	reachKm := deliveryRadiusKm
	if reachKm <= 0 {
		reachKm = defaultReachKm
	}
	return component{
		points: proportionalPoints(distance.Meters, reachKm*metersPerKm),
		reason: Reason{Code: ReasonDistance, DistanceMeters: distance.Meters}, isWon: true,
	}
}

func proportionalPoints(distanceMeters, reachMeters int) int {
	remaining := reachMeters - distanceMeters
	if remaining <= 0 {
		return 0
	}
	return (halves*distanceMaxPoints*remaining + reachMeters) / (halves * reachMeters)
}

func categoryComponent(facts MatchFacts) component {
	for _, category := range facts.ConsignorCategories {
		if slices.Contains(facts.NeedCategoryIDs, category.ID) {
			return categoryWin(needCategoryPoints, ReasonNeedsCategory, category)
		}
	}
	for _, category := range facts.ConsignorCategories {
		if slices.Contains(facts.AcceptedCategoryIDs, category.ID) {
			return categoryWin(acceptedCategoryPoints, ReasonAcceptsCategory, category)
		}
	}
	return component{}
}

func categoryWin(points int, code ReasonCode, category Category) component {
	return component{points: points, reason: Reason{Code: code, CategoryName: category.Name}, isWon: true}
}

func storeTypeComponent(preferred []string, storeType string) component {
	if len(preferred) == 0 {
		return component{points: noPreferencePoints}
	}
	if slices.Contains(preferred, storeType) {
		return component{points: preferredTypePoints, reason: Reason{Code: ReasonPreferredType}, isWon: true}
	}
	return component{}
}

func activityComponent(lastActiveAt, now time.Time) component {
	if lastActiveAt.IsZero() {
		return component{}
	}
	idle := now.Sub(lastActiveAt)
	if idle <= activeRecentlyWithin {
		return component{points: activeRecentlyPoints, reason: Reason{Code: ReasonActiveRecently}, isWon: true}
	}
	if idle <= activeThisMonthWithin {
		return component{points: activeThisMonthPoints}
	}
	return component{}
}
