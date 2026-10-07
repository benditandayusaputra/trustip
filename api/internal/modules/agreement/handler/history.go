package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

func versionParam(c fiber.Ctx) (int, error) {
	raw := c.Params("version")
	if raw == "" {
		return 0, nil
	}
	version, err := strconv.Atoi(raw)
	if err != nil || version < 1 {
		return 0, apperror.NotFound()
	}
	return version, nil
}

type termChangeResponse struct {
	Field  string `json:"field"`
	Before string `json:"before"`
	After  string `json:"after"`
}

type termsSnapshot struct {
	PricingModel  string `json:"pricing_model"`
	RetailPrice   int64  `json:"retail_price"`
	CommissionBPS int    `json:"commission_bps"`
	DepositPrice  int64  `json:"deposit_price"`
}

type itemChangeResponse struct {
	ProductID   string         `json:"product_id"`
	ProductName string         `json:"product_name"`
	Kind        string         `json:"kind"`
	Before      *termsSnapshot `json:"before"`
	After       termsSnapshot  `json:"after"`
}

type versionResponse struct {
	Version     int                  `json:"version"`
	Status      string               `json:"status"`
	Number      string               `json:"number"`
	AcceptedAt  string               `json:"accepted_at"`
	TermChanges []termChangeResponse `json:"term_changes"`
	ItemChanges []itemChangeResponse `json:"item_changes"`
}

func history(reader HistoryReader) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, ok := httpserver.ActorFrom(c)
		if !ok {
			return apperror.New(apperror.CodeUnauthenticated)
		}
		partnershipID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read partnership id: %w", err)
		}
		versions, err := reader.Execute(c.Context(), signedIn, partnershipID)
		if err != nil {
			return fmt.Errorf("read agreement history: %w", err)
		}
		items := make([]versionResponse, 0, len(versions))
		for _, version := range versions {
			items = append(items, toVersionResponse(version))
		}
		return httpserver.RespondList(c, httpserver.Page[versionResponse]{Items: items})
	}
}

func toVersionResponse(version domain.Version) versionResponse {
	agreement := version.Agreement
	terms := make([]termChangeResponse, 0, len(version.Diff.Terms))
	for _, change := range version.Diff.Terms {
		terms = append(terms, termChangeResponse(change))
	}
	items := make([]itemChangeResponse, 0, len(version.Diff.Items))
	for _, change := range version.Diff.Items {
		response := itemChangeResponse{
			ProductID: change.ProductID, ProductName: change.ProductName, Kind: string(change.Kind),
			After: snapshotOf(change.After),
		}
		if change.Kind != domain.ItemAdded {
			before := snapshotOf(change.Before)
			response.Before = &before
		}
		items = append(items, response)
	}
	return versionResponse{
		Version: agreement.Version, Status: string(agreement.Status), Number: agreement.Number(),
		AcceptedAt: timestamp(agreement.AcceptedAt), TermChanges: terms, ItemChanges: items,
	}
}

func snapshotOf(terms pricing.Terms) termsSnapshot {
	return termsSnapshot{
		PricingModel: string(terms.Model), RetailPrice: terms.RetailPrice.Int64(), CommissionBPS: terms.CommissionBPS,
		DepositPrice: terms.DepositPrice.Int64(),
	}
}
