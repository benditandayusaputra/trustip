package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	storesPath     = "/directory/stores"
	consignorsPath = "/directory/consignors"
	productsPath   = "/directory/products"
	trueValue      = "true"
)

type Lister[F, T any] interface {
	Execute(ctx context.Context, signedIn actor.Actor, request service.ListRequest[F]) (service.Page[T], error)
}

type Deps struct {
	Authenticate   fiber.Handler
	ListStores     Lister[domain.StoreSearchForm, domain.StoreCard]
	ListConsignors Lister[domain.ConsignorSearchForm, domain.ConsignorCard]
	ListProducts   Lister[domain.ConsignorSearchForm, domain.ProductCard]
	FilesBaseURL   string
	SearchLimiter  httpserver.RateLimiter
	SecurityLog    *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	searchLimit := httpserver.RateLimit(deps.SearchLimiter, httpserver.ByClientIP, deps.SecurityLog)
	present := presenter{filesBaseURL: deps.FilesBaseURL}
	api.Get(storesPath, deps.Authenticate, searchLimit, list(deps.ListStores, storeSearchForm, present.storeCard))
	api.Get(consignorsPath, deps.Authenticate, searchLimit,
		list(deps.ListConsignors, consignorSearchForm, present.consignorCard))
	api.Get(productsPath, deps.Authenticate, searchLimit,
		list(deps.ListProducts, consignorSearchForm, present.productCard))
}

func signedInActor(c fiber.Ctx) (actor.Actor, error) {
	signedIn, ok := httpserver.ActorFrom(c)
	if !ok {
		return actor.Actor{}, apperror.New(apperror.CodeUnauthenticated)
	}
	return signedIn, nil
}

func list[F, T, R any](
	lister Lister[F, T], formOf func(c fiber.Ctx, search domain.SearchForm) F, present func(T) R,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		text := httpserver.LineInput("q", c.Query("q"))
		if err := httpserver.RejectUncleanInputs(text); err != nil {
			return fmt.Errorf("check directory search: %w", err)
		}
		page, err := lister.Execute(c.Context(), signedIn, service.ListRequest[F]{
			Search: formOf(c, searchFormOf(c, text.Value)), Cursor: c.Query("cursor"), Limit: c.Query("limit"),
		})
		if err != nil {
			return fmt.Errorf("list directory: %w", err)
		}
		items := make([]R, 0, len(page.Items))
		for _, card := range page.Items {
			items = append(items, present(card))
		}
		return httpserver.RespondList(c, httpserver.Page[R]{
			Items: items, NextCursor: page.NextCursor, HasMore: page.HasMore, Search: searchContextOf(page.Context),
		})
	}
}

func searchFormOf(c fiber.Ctx, text string) domain.SearchForm {
	return domain.SearchForm{
		Origin: domain.OriginForm{
			Kind: c.Query("origin"), Latitude: c.Query("lat"), Longitude: c.Query("lng"),
			CityCode: c.Query("city_code"),
		},
		Radius: c.Query("radius_km"), Sort: c.Query("sort"), CategoryIDs: repeatedQuery(c, "category_id"), Text: text,
	}
}

func storeSearchForm(c fiber.Ctx, search domain.SearchForm) domain.StoreSearchForm {
	return domain.StoreSearchForm{
		Search: search, StoreType: c.Query("store_type"), PricingModel: c.Query("pricing_model"),
		MaxCommissionBPS:         c.Query("max_commission_bps"),
		WithoutPermitRequirement: c.Query("without_permit") == trueValue,
		WithoutHalalRequirement:  c.Query("without_halal") == trueValue,
	}
}

func consignorSearchForm(c fiber.Ctx, search domain.SearchForm) domain.ConsignorSearchForm {
	return domain.ConsignorSearchForm{
		Search: search, MinPrice: c.Query("min_price"), MaxPrice: c.Query("max_price"),
		MinShelfLifeDays: c.Query("min_shelf_life_days"), HasPermit: c.Query("has_permit") == trueValue,
		HasHalal: c.Query("has_halal") == trueValue, CanSample: c.Query("can_sample") == trueValue,
		DeliversToViewer: c.Query("delivers_to_me") == trueValue,
	}
}

func repeatedQuery(c fiber.Ctx, name string) []string {
	raw := c.Request().URI().QueryArgs().PeekMulti(name)
	if len(raw) == 0 {
		return nil
	}
	values := make([]string, 0, len(raw))
	for _, value := range raw {
		values = append(values, string(value))
	}
	return values
}
