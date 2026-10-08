package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/partnership/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	partnershipsPath = "/partnerships"
	partnershipPath  = partnershipsPath + "/:id"
	catalogPath      = partnershipPath + "/catalog"
)

type Lister interface {
	Execute(ctx context.Context, signedIn actor.Actor, request service.ListRequest) (service.Page, error)
}

type Reader interface {
	Execute(ctx context.Context, signedIn actor.Actor, partnershipID string) (domain.Partnership, error)
}

type CatalogReader interface {
	Execute(ctx context.Context, signedIn actor.Actor, partnershipID string) ([]domain.CatalogItem, error)
}

type Deps struct {
	Authenticate fiber.Handler
	List         Lister
	Get          Reader
	Catalog      CatalogReader
	Ending       EndingDeps
	FilesBaseURL string
}

func Mount(api fiber.Router, deps Deps) {
	files := fileLinks{baseURL: deps.FilesBaseURL}
	api.Get(partnershipsPath, deps.Authenticate, list(deps.List))
	api.Get(partnershipPath, deps.Authenticate, read(deps.Get, files))
	api.Get(catalogPath, deps.Authenticate, catalog(deps.Catalog, files))
	mountEnding(api, deps, files)
}

func signedInActor(c fiber.Ctx) (actor.Actor, error) {
	signedIn, ok := httpserver.ActorFrom(c)
	if !ok {
		return actor.Actor{}, apperror.New(apperror.CodeUnauthenticated)
	}
	return signedIn, nil
}

func list(lister Lister) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		page, err := lister.Execute(c.Context(), signedIn,
			service.ListRequest{Status: c.Query("status"), Cursor: c.Query("cursor"), Limit: c.Query("limit")})
		if err != nil {
			return fmt.Errorf("list partnerships: %w", err)
		}
		items := make([]summaryResponse, 0, len(page.Items))
		for _, summary := range page.Items {
			items = append(items, toSummaryResponse(summary))
		}
		return httpserver.RespondList(c, httpserver.Page[summaryResponse]{
			Items: items, NextCursor: page.NextCursor, HasMore: page.HasMore,
		})
	}
}

func read(reader Reader, files fileLinks) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		partnershipID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read partnership id: %w", err)
		}
		partnership, err := reader.Execute(c.Context(), signedIn, partnershipID)
		if err != nil {
			return fmt.Errorf("read partnership: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: files.partnership(partnership)})
	}
}

func catalog(reader CatalogReader, files fileLinks) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		partnershipID, err := httpserver.IDParam(c, "id")
		if err != nil {
			return fmt.Errorf("read partnership id: %w", err)
		}
		found, err := reader.Execute(c.Context(), signedIn, partnershipID)
		if err != nil {
			return fmt.Errorf("read partner catalog: %w", err)
		}
		items := make([]catalogItemResponse, 0, len(found))
		for _, item := range found {
			items = append(items, files.catalogItem(item))
		}
		return httpserver.RespondList(c, httpserver.Page[catalogItemResponse]{Items: items})
	}
}
