package main

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	directoryhandler "github.com/benditandayusaputra/trustip/api/internal/modules/directory/handler"
	directoryrepository "github.com/benditandayusaputra/trustip/api/internal/modules/directory/repository"
	directoryservice "github.com/benditandayusaputra/trustip/api/internal/modules/directory/service"
	needhandler "github.com/benditandayusaputra/trustip/api/internal/modules/need/handler"
	needservice "github.com/benditandayusaputra/trustip/api/internal/modules/need/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
)

const searchesPerMinutePerIP = 60

type directoryComponents struct {
	viewers    directoryservice.ViewerReader
	stores     directoryservice.StoreDirectory
	consignors directoryservice.ConsignorDirectory
	contacts   directoryservice.ContactReader
	needs      needservice.Deps
}

func buildDirectoryComponents(pool *database.Pool, clock serverClock) directoryComponents {
	directory := directoryrepository.NewDirectory(pool)
	return directoryComponents{
		viewers: directory, stores: directory, consignors: directory, contacts: directory,
		needs: buildNeedDeps(pool, clock),
	}
}

type directoryRoutes struct {
	authenticate fiber.Handler
	writeLimiter httpserver.RateLimiter
	filesBaseURL string
	securityLog  *slog.Logger
	clock        serverClock
}

func mountDirectory(api fiber.Router, components directoryComponents, routes directoryRoutes) {
	directoryhandler.Mount(api, directoryhandler.Deps{
		Authenticate: routes.authenticate,
		ListStores:   directoryservice.NewListStores(components.viewers, components.stores, routes.clock),
		ListConsignors: directoryservice.NewListConsignors(components.viewers, components.consignors,
			routes.clock),
		ListProducts: directoryservice.NewListProducts(components.viewers, components.consignors, routes.clock),
		GetContact:   directoryservice.NewGetContact(components.contacts),
		FilesBaseURL: routes.filesBaseURL,
		SearchLimiter: ratelimit.New(ratelimit.Rule{Limit: searchesPerMinutePerIP, Window: time.Minute},
			routes.clock),
		SecurityLog: routes.securityLog,
	})
	needhandler.Mount(api, needDeps(components.needs, routes))
}
