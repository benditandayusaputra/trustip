package handler

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
)

type partyResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type splitResponse struct {
	Gross          int64 `json:"gross"`
	StoreShare     int64 `json:"store_share"`
	ConsignorShare int64 `json:"consignor_share"`
}

type termsResponse struct {
	PricingModel  string `json:"pricing_model"`
	RetailPrice   int64  `json:"retail_price"`
	CommissionBPS int    `json:"commission_bps"`
	DepositPrice  int64  `json:"deposit_price"`
}

type sheetBatchResponse struct {
	ID         string        `json:"id"`
	ExpiryDate *string       `json:"expiry_date"`
	ReceivedAt string        `json:"received_at"`
	Balance    int           `json:"balance"`
	Terms      termsResponse `json:"terms"`
}

type sheetProductResponse struct {
	ProductID   string               `json:"product_id"`
	ProductName string               `json:"product_name"`
	SKU         string               `json:"sku"`
	Unit        string               `json:"unit"`
	RecordedQty int                  `json:"recorded_qty"`
	Batches     []sheetBatchResponse `json:"batches"`
}

type sheetResponse struct {
	PartnershipID    string                 `json:"partnership_id"`
	IsActive         bool                   `json:"is_active"`
	AgreementVersion int                    `json:"agreement_version"`
	Store            partyResponse          `json:"store"`
	Consignor        partyResponse          `json:"consignor"`
	LastDocumentAt   *string                `json:"last_document_at"`
	Products         []sheetProductResponse `json:"products"`
}

func toSheetResponse(sheet domain.Sheet) sheetResponse {
	response := sheetResponse{
		PartnershipID: sheet.PartnershipID, IsActive: sheet.IsActive, AgreementVersion: sheet.AgreementVersion,
		Store:     partyResponse{ID: sheet.StoreID, Name: sheet.StoreName},
		Consignor: partyResponse{ID: sheet.ConsignorID, Name: sheet.ConsignorName},
		Products:  make([]sheetProductResponse, 0, len(sheet.Products)),
	}
	if sheet.LastDocument.Given {
		last := timestamp(sheet.LastDocument.At)
		response.LastDocumentAt = &last
	}
	for _, product := range sheet.Products {
		response.Products = append(response.Products, toSheetProduct(product))
	}
	return response
}

func toSheetProduct(product domain.StockedProduct) sheetProductResponse {
	batches := make([]sheetBatchResponse, 0, len(product.Batches))
	for _, batch := range product.Batches {
		batches = append(batches, sheetBatchResponse{
			ID: batch.Stock.ID, ExpiryDate: expiryOf(batch), ReceivedAt: timestamp(batch.Stock.ReceivedAt),
			Balance: batch.Stock.Balance, Terms: termsOf(batch.Terms),
		})
	}
	return sheetProductResponse{
		ProductID: product.ProductID, ProductName: product.Name, SKU: product.SKU, Unit: product.Unit,
		RecordedQty: product.RecordedQty(), Batches: batches,
	}
}

func expiryOf(batch domain.PricedBatch) *string {
	date, isGiven := batch.Stock.Expiry.Date()
	if !isGiven {
		return nil
	}
	formatted := date.Format(time.DateOnly)
	return &formatted
}

func termsOf(terms pricing.Terms) termsResponse {
	return termsResponse{
		PricingModel: string(terms.Model), RetailPrice: terms.RetailPrice.Int64(), CommissionBPS: terms.CommissionBPS,
		DepositPrice: terms.DepositPrice.Int64(),
	}
}

func toSplit(split pricing.Split) splitResponse {
	return splitResponse{
		Gross: split.Gross.Int64(), StoreShare: split.StoreShare.Int64(), ConsignorShare: split.ConsignorShare.Int64(),
	}
}
