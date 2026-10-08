package handler

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
)

type counterpartRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type summaryResponse struct {
	ID           string         `json:"id"`
	Status       string         `json:"status"`
	StartedAt    string         `json:"started_at"`
	EndedAt      *string        `json:"ended_at"`
	Counterpart  counterpartRef `json:"counterpart"`
	StockQty     int            `json:"stock_qty"`
	UnpaidAmount int64          `json:"unpaid_amount"`
}

type point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type counterpartResponse struct {
	counterpartRef
	Slug        string  `json:"slug"`
	Description string  `json:"description"`
	WhatsApp    string  `json:"whatsapp"`
	CityName    string  `json:"city_name"`
	District    string  `json:"district"`
	Address     string  `json:"address"`
	Location    *point  `json:"location"`
	LogoURL     *string `json:"logo_url"`
}

type partnershipResponse struct {
	ID           string              `json:"id"`
	Status       string              `json:"status"`
	StartedAt    string              `json:"started_at"`
	EndedAt      *string             `json:"ended_at"`
	ShowOnPublic bool                `json:"show_on_public"`
	EndReason    string              `json:"end_reason"`
	Counterpart  counterpartResponse `json:"counterpart"`
}

type splitResponse struct {
	Gross          int64 `json:"gross"`
	StoreShare     int64 `json:"store_share"`
	ConsignorShare int64 `json:"consignor_share"`
}

type termsResponse struct {
	PricingModel  string        `json:"pricing_model"`
	RetailPrice   int64         `json:"retail_price"`
	CommissionBPS int           `json:"commission_bps"`
	DepositPrice  int64         `json:"deposit_price"`
	PerUnit       splitResponse `json:"per_unit"`
}

type catalogItemResponse struct {
	ProductID    string         `json:"product_id"`
	Name         string         `json:"name"`
	SKU          string         `json:"sku"`
	Unit         string         `json:"unit"`
	Description  string         `json:"description"`
	NetContent   string         `json:"net_content"`
	CategoryName string         `json:"category_name"`
	PhotoURL     *string        `json:"photo_url"`
	ShelfLife    *int           `json:"shelf_life_days"`
	IsActive     bool           `json:"is_active"`
	Terms        *termsResponse `json:"terms"`
}

func timestamp(at time.Time) string {
	return at.UTC().Format(time.RFC3339)
}

func optionalTime(at time.Time) *string {
	if at.IsZero() {
		return nil
	}
	formatted := timestamp(at)
	return &formatted
}

func toSummaryResponse(summary domain.Summary) summaryResponse {
	return summaryResponse{
		ID: summary.ID, Status: string(summary.Status), StartedAt: timestamp(summary.StartedAt),
		EndedAt: optionalTime(summary.EndedAt),
		Counterpart: counterpartRef{
			ID: summary.CounterpartID, Name: summary.CounterpartName, Type: string(summary.CounterpartType),
		},
		StockQty: summary.StockQty, UnpaidAmount: summary.UnpaidAmount,
	}
}

type fileLinks struct {
	baseURL string
}

func (f fileLinks) urlOf(storageKey string) *string {
	if storageKey == "" {
		return nil
	}
	url := f.baseURL + "/" + storageKey
	return &url
}

func (f fileLinks) partnership(partnership domain.Partnership) partnershipResponse {
	view := partnership.View()
	counterpart := view.Counterpart
	response := counterpartResponse{
		counterpartRef: counterpartRef{ID: counterpart.ID, Name: counterpart.Name, Type: string(counterpart.Type)},
		Slug:           counterpart.Slug, Description: counterpart.Description, WhatsApp: counterpart.WhatsApp,
		CityName: counterpart.CityName, District: counterpart.District, Address: counterpart.Address,
		LogoURL: f.urlOf(counterpart.LogoKey),
	}
	if counterpart.HasLocation {
		response.Location = &point{Latitude: view.Location.Latitude, Longitude: view.Location.Longitude}
	}
	return partnershipResponse{
		ID: partnership.ID, Status: string(partnership.Status), StartedAt: timestamp(partnership.StartedAt),
		EndedAt: optionalTime(partnership.EndedAt), ShowOnPublic: partnership.ShowOnPublic,
		EndReason: partnership.EndReason, Counterpart: response,
	}
}

func shelfLifeOf(days int) *int {
	if days == 0 {
		return nil
	}
	return &days
}

func (f fileLinks) catalogItem(item domain.CatalogItem) catalogItemResponse {
	response := catalogItemResponse{
		ProductID: item.ProductID, Name: item.Name, SKU: item.SKU, Unit: item.Unit, Description: item.Description,
		NetContent: item.NetContent, CategoryName: item.CategoryName, PhotoURL: f.urlOf(item.PhotoKey),
		ShelfLife: shelfLifeOf(item.ShelfLife), IsActive: item.IsActive,
	}
	if !item.IsAgreed {
		return response
	}
	perUnit := item.Terms.PerUnit()
	response.Terms = &termsResponse{
		PricingModel: string(item.Terms.Model), RetailPrice: item.Terms.RetailPrice.Int64(),
		CommissionBPS: item.Terms.CommissionBPS, DepositPrice: item.Terms.DepositPrice.Int64(),
		PerUnit: splitResponse{
			Gross: perUnit.Gross.Int64(), StoreShare: perUnit.StoreShare.Int64(),
			ConsignorShare: perUnit.ConsignorShare.Int64(),
		},
	}
	return response
}
