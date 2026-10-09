package domain

import "time"

const VisitorID = "00000000-0000-0000-0000-000000000000"

type NeedPageStore struct {
	ID          string
	ProfileSlug string
	Name        string
	StoreType   string
	District    string
	CityName    string
}

type PublicNeed struct {
	ID                string
	Title             string
	CategoryName      string
	Description       string
	EstQtyPerWeek     int
	MaxRetailPrice    int64
	PreferredModel    string
	OfferedCommission CommissionOffer
	Status            string
	ExpiresAt         time.Time
	Store             NeedPageStore
}

type CommissionOffer struct {
	BPS     int
	IsGiven bool
}
