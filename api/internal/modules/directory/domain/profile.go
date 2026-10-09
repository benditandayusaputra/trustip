package domain

import (
	"regexp"
	"time"
)

const (
	PublicProductLimit = 20
	maxSlugLength      = 80
)

var slugFormat = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func IsSlug(raw string) bool {
	return len(raw) <= maxSlugLength && slugFormat.MatchString(raw)
}

type Place struct {
	Point   Point
	IsKnown bool
}

type StoreProfile struct {
	ID                 string
	Slug               string
	Name               string
	Description        string
	Address            string
	District           string
	CityName           string
	Location           Place
	LogoStorageKey     string
	PhotoStorageKeys   []string
	StoreType          string
	OpeningHours       string
	AcceptedCategories []Category
	Terms              StoreTerms
	OtherRequirements  string
	Reputation         StoreReputation
	JoinedAt           time.Time
	OpenNeeds          []OpenNeed
}

type OpenNeed struct {
	ID           string
	Title        string
	CategoryName string
	ExpiresAt    time.Time
}

type DeclaredLegality struct {
	BusinessIDNo         string
	DistributionPermitNo string
	HalalCertNo          string
}

type ConsignorProfile struct {
	ID             string
	Slug           string
	Name           string
	Description    string
	District       string
	CityName       string
	Location       Place
	LogoStorageKey string
	Legality       DeclaredLegality
	ActivePartners int
	JoinedAt       time.Time
	Products       []PublicProduct
}

type PublicProduct struct {
	SKU                  string
	Name                 string
	Unit                 string
	CategoryName         string
	SuggestedRetailPrice int64
	PhotoStorageKey      string
}
