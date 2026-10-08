package handler

import (
	"github.com/benditandayusaputra/trustip/api/internal/modules/stock/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/stock/service"
)

type batchResponse struct {
	ID         string  `json:"id"`
	BatchCode  string  `json:"batch_code"`
	ExpiryDate *string `json:"expiry_date"`
	ReceivedAt string  `json:"received_at"`
	Balance    int     `json:"balance"`
	AgeDays    int     `json:"age_days"`
	Flag       string  `json:"flag"`
}

type productStockResponse struct {
	ProductID     string          `json:"product_id"`
	ProductName   string          `json:"product_name"`
	SKU           string          `json:"sku"`
	Unit          string          `json:"unit"`
	Balance       int             `json:"balance"`
	Flag          string          `json:"flag"`
	NearestExpiry *string         `json:"nearest_expiry"`
	Batches       []batchResponse `json:"batches"`
}

type counterpartResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type holdingResponse struct {
	PartnershipID string              `json:"partnership_id"`
	Counterpart   counterpartResponse `json:"counterpart"`
	ProductID     string              `json:"product_id"`
	ProductName   string              `json:"product_name"`
	SKU           string              `json:"sku"`
	Unit          string              `json:"unit"`
	Balance       int                 `json:"balance"`
	Flag          string              `json:"flag"`
	NearestExpiry *string             `json:"nearest_expiry"`
}

type cardEntryResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Qty          int    `json:"qty"`
	OccurredAt   string `json:"occurred_at"`
	BatchCode    string `json:"batch_code"`
	SourceType   string `json:"source_type"`
	SourceID     string `json:"source_id"`
	SourceNumber string `json:"source_number"`
	BalanceAfter int    `json:"balance_after"`
}

type cardResponse struct {
	ProductID   string              `json:"product_id"`
	ProductName string              `json:"product_name"`
	SKU         string              `json:"sku"`
	Unit        string              `json:"unit"`
	Entries     []cardEntryResponse `json:"entries"`
	NextCursor  *string             `json:"next_cursor"`
	HasMore     bool                `json:"has_more"`
}

func nearestOf(view domain.ProductView) *string {
	if view.NearestExpiry == "" {
		return nil
	}
	nearest := view.NearestExpiry
	return &nearest
}

func toProductStocks(views []domain.ProductView) []productStockResponse {
	responses := make([]productStockResponse, 0, len(views))
	for _, view := range views {
		batches := make([]batchResponse, 0, len(view.BatchViews))
		for _, batch := range view.BatchViews {
			batches = append(batches, batchResponse{
				ID: batch.ID, BatchCode: batch.Code, ExpiryDate: expiryDate(batch.Expiry),
				ReceivedAt: timestamp(batch.ReceivedAt), Balance: batch.Balance, AgeDays: batch.AgeDays,
				Flag: string(batch.Flag),
			})
		}
		responses = append(responses, productStockResponse{
			ProductID: view.ID, ProductName: view.Name, SKU: view.SKU, Unit: view.Unit, Balance: view.Balance,
			Flag: string(view.Flag), NearestExpiry: nearestOf(view), Batches: batches,
		})
	}
	return responses
}

func toHoldings(views []domain.ProductView) []holdingResponse {
	responses := make([]holdingResponse, 0, len(views))
	for _, view := range views {
		counterpart := view.Counterpart
		responses = append(responses, holdingResponse{
			PartnershipID: counterpart.PartnershipID,
			Counterpart:   counterpartResponse{ID: counterpart.ID, Name: counterpart.Name, Type: counterpart.Type},
			ProductID:     view.ID, ProductName: view.Name, SKU: view.SKU, Unit: view.Unit, Balance: view.Balance,
			Flag: string(view.Flag), NearestExpiry: nearestOf(view),
		})
	}
	return responses
}

func toCardResponse(page service.CardPage) cardResponse {
	entries := make([]cardEntryResponse, 0, len(page.Entries))
	for _, entry := range page.Entries {
		entries = append(entries, cardEntryResponse{
			ID: entry.ID, Type: entry.Type, Qty: entry.Qty, OccurredAt: timestamp(entry.OccurredAt),
			BatchCode: entry.BatchCode, SourceType: entry.SourceType, SourceID: entry.SourceID,
			SourceNumber: entry.SourceNumber, BalanceAfter: entry.BalanceAfter,
		})
	}
	response := cardResponse{
		ProductID: page.Product.ID, ProductName: page.Product.Name, SKU: page.Product.SKU, Unit: page.Product.Unit,
		Entries: entries, HasMore: page.HasMore,
	}
	if page.NextCursor != "" {
		cursor := page.NextCursor
		response.NextCursor = &cursor
	}
	return response
}
