package main

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"

	auditrepository "github.com/benditandayusaputra/trustip/api/internal/modules/audit/repository"
	producthandler "github.com/benditandayusaputra/trustip/api/internal/modules/product/handler"
	productrepository "github.com/benditandayusaputra/trustip/api/internal/modules/product/repository"
	productservice "github.com/benditandayusaputra/trustip/api/internal/modules/product/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

type productComponents struct {
	products productservice.ProductStore
	audit    productservice.AuditTrail
	tx       productservice.Transactor
	clock    productservice.Clock
	ids      productservice.IDGenerator
}

func buildProductComponents(pool *database.Pool, clock productservice.Clock) productComponents {
	return productComponents{
		products: productrepository.NewProducts(pool),
		audit:    newAuditTrail(auditrepository.NewRecords(pool)),
		tx:       pool,
		clock:    clock,
		ids:      idgen.UUIDv7{},
	}
}

type productRoutes struct {
	authenticate fiber.Handler
	filesBaseURL string
	writeLimiter httpserver.RateLimiter
	securityLog  *slog.Logger
}

func productDeps(components productComponents, routes productRoutes) producthandler.Deps {
	deps := productservice.ProductDeps{
		Products: components.products, Audit: components.audit, Tx: components.tx, Clock: components.clock,
		IDs: components.ids,
	}
	return producthandler.Deps{
		Authenticate: routes.authenticate,
		Create:       productservice.NewCreateProduct(deps),
		Update:       productservice.NewUpdateProduct(deps),
		Get:          productservice.NewGetProduct(components.products),
		Archive:      productservice.NewArchiveProduct(deps),
		Reactivate:   productservice.NewReactivateProduct(deps),
		List:         productservice.NewListProducts(components.products),
		FilesBaseURL: routes.filesBaseURL,
		WriteLimiter: routes.writeLimiter,
		SecurityLog:  routes.securityLog,
	}
}
