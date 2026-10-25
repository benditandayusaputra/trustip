package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
)

//go:embed demo_world.json
var demoWorldJSON []byte

const (
	kilometresPerDegree = 111.32
	degreesPerHalfTurn  = 180
	permille            = 1000
)

type demoWorld struct {
	CityCode      string             `json:"city_code"`
	District      string             `json:"district"`
	Center        demoPoint          `json:"center"`
	Password      string             `json:"password"`
	HistoryDays   int                `json:"history_days"`
	SetupDaysAgo  int                `json:"setup_days_ago"`
	Organizations []demoOrganization `json:"organizations"`
	Products      []demoProduct      `json:"products"`
	Needs         []demoNeed         `json:"needs"`
	Partnerships  []demoPartnership  `json:"partnerships"`
	Proposals     []demoProposal     `json:"proposals"`
}

type demoPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type demoOrganization struct {
	Key         string         `json:"key"`
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Owner       string         `json:"owner"`
	Email       string         `json:"email"`
	Address     string         `json:"address"`
	Description string         `json:"description"`
	CityCode    string         `json:"city_code"`
	District    string         `json:"district"`
	NorthKm     float64        `json:"north_km"`
	EastKm      float64        `json:"east_km"`
	Consignor   *demoConsignor `json:"consignor"`
	Store       *demoStore     `json:"store"`
}

type demoConsignor struct {
	DeliveryRadiusKm    int      `json:"delivery_radius_km"`
	DeliveryDays        []string `json:"delivery_days"`
	PreferredStoreTypes []string `json:"preferred_store_types"`
	CanProvideSample    bool     `json:"can_provide_sample"`
	PaymentInfo         string   `json:"payment_info"`
}

type demoStore struct {
	StoreType        string   `json:"store_type"`
	Categories       []string `json:"categories"`
	FixedDeposit     bool     `json:"fixed_deposit"`
	Commission       bool     `json:"commission"`
	CommissionMinBPS int      `json:"commission_min_bps"`
	CommissionMaxBPS int      `json:"commission_max_bps"`
	OpeningHours     string   `json:"opening_hours"`
}

type demoProduct struct {
	Key           string `json:"key"`
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	Category      string `json:"category"`
	Unit          string `json:"unit"`
	Retail        int64  `json:"retail"`
	Deposit       int64  `json:"deposit"`
	ShelfLifeDays int    `json:"shelf_life_days"`
	NetContent    string `json:"net_content"`
	Seeking       bool   `json:"seeking"`
}

type demoNeed struct {
	Key              string `json:"key"`
	Store            string `json:"store"`
	Title            string `json:"title"`
	Category         string `json:"category"`
	Description      string `json:"description"`
	WeeklyQty        int64  `json:"weekly_qty"`
	MaxRetailPrice   int64  `json:"max_retail_price"`
	PreferredModel   string `json:"preferred_model"`
	CommissionBPS    int64  `json:"commission_bps"`
	DaysAgo          int    `json:"days_ago"`
	FulfilledDaysAgo int    `json:"fulfilled_days_ago"`
}

type demoTerms struct {
	Schedule         string `json:"schedule"`
	Rhythm           string `json:"rhythm"`
	PaymentTermsDays int    `json:"payment_terms_days"`
	Notes            string `json:"notes"`
}

type demoItem struct {
	Product       string `json:"product"`
	Model         string `json:"model"`
	Retail        int64  `json:"retail"`
	Deposit       int64  `json:"deposit"`
	CommissionBPS int    `json:"commission_bps"`
	Qty           int    `json:"qty"`
}

type demoRoutine struct {
	From            int    `json:"from"`
	Until           int    `json:"until"`
	Every           int    `json:"every"`
	WeekendPermille int    `json:"weekend_permille"`
	SellPermille    int    `json:"sell_permille"`
	ReturnRest      bool   `json:"return_rest"`
	Reason          string `json:"reason"`
}

func (r demoRoutine) fallsOn(daysAgo int) bool {
	return daysAgo <= r.From && daysAgo >= r.Until && (r.From-daysAgo)%r.Every == 0
}

type demoEnding struct {
	DaysAgo int    `json:"days_ago"`
	By      string `json:"by"`
	Reason  string `json:"reason"`
}

type demoPartnership struct {
	demoTerms

	Key              string      `json:"key"`
	Consignor        string      `json:"consignor"`
	Store            string      `json:"store"`
	Need             string      `json:"need"`
	Items            []demoItem  `json:"items"`
	Drops            demoRoutine `json:"drops"`
	Counts           demoRoutine `json:"counts"`
	PayAfterDays     int         `json:"pay_after_days"`
	ConfirmAfterDays int         `json:"confirm_after_days"`
	Ended            *demoEnding `json:"ended"`
}

type demoDecline struct {
	DaysAgo int    `json:"days_ago"`
	Reason  string `json:"reason"`
}

type demoProposal struct {
	demoTerms

	From     string       `json:"from"`
	To       string       `json:"to"`
	DaysAgo  int          `json:"days_ago"`
	Message  string       `json:"message"`
	Items    []demoItem   `json:"items"`
	Declined *demoDecline `json:"declined"`
}

func loadDemoWorld() (demoWorld, error) {
	var world demoWorld
	if err := json.Unmarshal(demoWorldJSON, &world); err != nil {
		return demoWorld{}, fmt.Errorf("read demo world: %w", err)
	}
	return world, nil
}

func (w demoWorld) locate(organization demoOrganization) demoPoint {
	latitude := w.Center.Latitude + organization.NorthKm/kilometresPerDegree
	perLongitude := kilometresPerDegree * math.Cos(latitude*math.Pi/degreesPerHalfTurn)
	return demoPoint{Latitude: latitude, Longitude: w.Center.Longitude + organization.EastKm/perLongitude}
}
