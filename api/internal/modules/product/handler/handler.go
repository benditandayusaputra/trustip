package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/product/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	productsPath   = "/products"
	productPath    = productsPath + "/:id"
	archivePath    = productPath + "/archive"
	reactivatePath = productPath + "/reactivate"
)

type Creator interface {
	Execute(ctx context.Context, signedIn actor.Actor, form domain.ProductForm) (domain.Product, error)
}

type Updater interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.UpdateProductCommand) (domain.Product, error)
}

type Reader interface {
	Execute(ctx context.Context, signedIn actor.Actor, productID string) (domain.Product, error)
}

type Lister interface {
	Execute(ctx context.Context, signedIn actor.Actor, request service.ListRequest) (service.Page, error)
}

type Deps struct {
	Authenticate fiber.Handler
	Create       Creator
	Update       Updater
	Get          Reader
	Archive      Reader
	Reactivate   Reader
	List         Lister
	FilesBaseURL string
	WriteLimiter httpserver.RateLimiter
	SecurityLog  *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	writeLimit := httpserver.RateLimit(deps.WriteLimiter, httpserver.BySignedInUser, deps.SecurityLog)
	present := presenter{filesBaseURL: deps.FilesBaseURL}
	api.Get(productsPath, deps.Authenticate, list(deps.List, present))
	api.Post(productsPath, deps.Authenticate, writeLimit, create(deps.Create, present))
	api.Get(productPath, deps.Authenticate, act(deps.Get, present))
	api.Patch(productPath, deps.Authenticate, writeLimit, update(deps.Update, present))
	api.Post(archivePath, deps.Authenticate, writeLimit, act(deps.Archive, present))
	api.Post(reactivatePath, deps.Authenticate, writeLimit, act(deps.Reactivate, present))
}

func signedInActor(c fiber.Ctx) (actor.Actor, error) {
	signedIn, ok := httpserver.ActorFrom(c)
	if !ok {
		return actor.Actor{}, apperror.New(apperror.CodeUnauthenticated)
	}
	return signedIn, nil
}

func create(creator Creator, present presenter) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		form, err := readForm(c)
		if err != nil {
			return err
		}
		product, err := creator.Execute(c.Context(), signedIn, form)
		if err != nil {
			return fmt.Errorf("create product: %w", err)
		}
		c.Location(httpserver.APIPrefix + productsPath + "/" + product.ID)
		return httpserver.Respond(c, httpserver.Reply{Status: fiber.StatusCreated, Data: present.product(product)})
	}
}

func update(updater Updater, present presenter) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		productID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read product id: %w", err)
		}
		form, err := readForm(c)
		if err != nil {
			return err
		}
		product, err := updater.Execute(c.Context(), signedIn,
			service.UpdateProductCommand{ProductID: productID, Form: form})
		if err != nil {
			return fmt.Errorf("update product: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: present.product(product)})
	}
}

func act(reader Reader, present presenter) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		productID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read product id: %w", err)
		}
		product, err := reader.Execute(c.Context(), signedIn, productID)
		if err != nil {
			return fmt.Errorf("act on product: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: present.product(product)})
	}
}

func list(lister Lister, present presenter) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		search := httpserver.LineInput("q", c.Query("q"))
		if err := httpserver.RejectUncleanInputs(search); err != nil {
			return fmt.Errorf("check product search: %w", err)
		}
		page, err := lister.Execute(c.Context(), signedIn, service.ListRequest{
			Search: domain.SearchForm{
				Text: search.Value, CategoryID: c.Query("category_id"), Status: c.Query("status"),
			},
			Cursor: c.Query("cursor"), Limit: c.Query("limit"),
		})
		if err != nil {
			return fmt.Errorf("list products: %w", err)
		}
		items := make([]productResponse, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, present.product(item))
		}
		return httpserver.RespondList(c, httpserver.Page[productResponse]{
			Items: items, NextCursor: page.NextCursor, HasMore: page.HasMore,
		})
	}
}
