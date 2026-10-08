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
	storesPath = "/directory/stores"
	trueValue  = "true"
)

type StoreLister interface {
	Execute(ctx context.Context, signedIn actor.Actor, request service.StoreListRequest) (service.StorePage, error)
}

type Deps struct {
	Authenticate  fiber.Handler
	ListStores    StoreLister
	FilesBaseURL  string
	SearchLimiter httpserver.RateLimiter
	SecurityLog   *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	searchLimit := httpserver.RateLimit(deps.SearchLimiter, httpserver.ByClientIP, deps.SecurityLog)
	present := presenter{filesBaseURL: deps.FilesBaseURL}
	api.Get(storesPath, deps.Authenticate, searchLimit, listStores(deps.ListStores, present))
}

func signedInActor(c fiber.Ctx) (actor.Actor, error) {
	signedIn, ok := httpserver.ActorFrom(c)
	if !ok {
		return actor.Actor{}, apperror.New(apperror.CodeUnauthenticated)
	}
	return signedIn, nil
}

func listStores(lister StoreLister, present presenter) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		search := httpserver.LineInput("q", c.Query("q"))
		if err := httpserver.RejectUncleanInputs(search); err != nil {
			return fmt.Errorf("check store search: %w", err)
		}
		page, err := lister.Execute(c.Context(), signedIn, service.StoreListRequest{
			Search: storeSearchForm(c, search.Value), Cursor: c.Query("cursor"), Limit: c.Query("limit"),
		})
		if err != nil {
			return fmt.Errorf("list directory stores: %w", err)
		}
		items := make([]storeCardResponse, 0, len(page.Items))
		for _, card := range page.Items {
			items = append(items, present.storeCard(card))
		}
		return httpserver.RespondList(c, httpserver.Page[storeCardResponse]{
			Items: items, NextCursor: page.NextCursor, HasMore: page.HasMore, Search: searchContextOf(page.Context),
		})
	}
}

func storeSearchForm(c fiber.Ctx, text string) domain.StoreSearchForm {
	return domain.StoreSearchForm{
		Origin: domain.OriginForm{
			Kind: c.Query("origin"), Latitude: c.Query("lat"), Longitude: c.Query("lng"),
			CityCode: c.Query("city_code"),
		},
		Radius: c.Query("radius_km"), Sort: c.Query("sort"), StoreType: c.Query("store_type"),
		CategoryIDs: repeatedQuery(c, "category_id"), PricingModel: c.Query("pricing_model"),
		MaxCommissionBPS:         c.Query("max_commission_bps"),
		WithoutPermitRequirement: c.Query("without_permit") == trueValue,
		WithoutHalalRequirement:  c.Query("without_halal") == trueValue,
		Text:                     text,
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
