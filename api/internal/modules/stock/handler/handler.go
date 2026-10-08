package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
	"github.com/benditandayusaputra/trustip/api/internal/modules/stock/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/stock/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	stockPath            = "/stock"
	partnershipStockPath = "/partnerships/:id/stock"
	cardPath             = partnershipStockPath + "/:product_id/card"
)

type PartnershipReader interface {
	Execute(ctx context.Context, signedIn actor.Actor, partnershipID string) ([]domain.ProductView, error)
}

type OrganizationReader interface {
	Execute(ctx context.Context, signedIn actor.Actor) ([]domain.ProductView, error)
}

type CardReader interface {
	Execute(ctx context.Context, signedIn actor.Actor, request service.CardRequest) (service.CardPage, error)
}

type Deps struct {
	Authenticate fiber.Handler
	Partnership  PartnershipReader
	Organization OrganizationReader
	Card         CardReader
}

func Mount(api fiber.Router, deps Deps) {
	api.Get(stockPath, deps.Authenticate, organizationStock(deps.Organization))
	api.Get(partnershipStockPath, deps.Authenticate, partnershipStock(deps.Partnership))
	api.Get(cardPath, deps.Authenticate, stockCard(deps.Card))
}

func signedInActor(c fiber.Ctx) (actor.Actor, error) {
	signedIn, ok := httpserver.ActorFrom(c)
	if !ok {
		return actor.Actor{}, apperror.New(apperror.CodeUnauthenticated)
	}
	return signedIn, nil
}

func organizationStock(reader OrganizationReader) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		views, err := reader.Execute(c.Context(), signedIn)
		if err != nil {
			return fmt.Errorf("list stock: %w", err)
		}
		return httpserver.RespondList(c, httpserver.Page[holdingResponse]{Items: toHoldings(views)})
	}
}

func partnershipStock(reader PartnershipReader) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		partnershipID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read partnership id: %w", err)
		}
		views, err := reader.Execute(c.Context(), signedIn, partnershipID)
		if err != nil {
			return fmt.Errorf("read partnership stock: %w", err)
		}
		return httpserver.RespondList(c, httpserver.Page[productStockResponse]{Items: toProductStocks(views)})
	}
}

func stockCard(reader CardReader) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		partnershipID, partnershipErr := httpserver.IDParam(c, "id")
		productID, productErr := httpserver.IDParam(c, "product_id")
		if partnershipErr != nil || productErr != nil {
			return apperror.NotFound()
		}
		page, err := reader.Execute(c.Context(), signedIn, service.CardRequest{
			PartnershipID: partnershipID, ProductID: productID, Cursor: c.Query("cursor"), Limit: c.Query("limit"),
		})
		if err != nil {
			return fmt.Errorf("read stock card: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toCardResponse(page)})
	}
}

func timestamp(at time.Time) string {
	return at.UTC().Format(time.RFC3339)
}

func expiryDate(expiry stock.Expiry) *string {
	date, isGiven := expiry.Date()
	if !isGiven {
		return nil
	}
	formatted := date.Format(time.DateOnly)
	return &formatted
}
