package main

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	directoryhandler "github.com/benditandayusaputra/trustip/api/internal/modules/directory/handler"
	directoryrepository "github.com/benditandayusaputra/trustip/api/internal/modules/directory/repository"
	directoryservice "github.com/benditandayusaputra/trustip/api/internal/modules/directory/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
)

const searchesPerMinutePerIP = 60

type directoryComponents struct {
	viewers    directoryservice.ViewerReader
	stores     directoryservice.StoreDirectory
	consignors directoryservice.ConsignorDirectory
}

func buildDirectoryComponents(pool *database.Pool) directoryComponents {
	directory := directoryrepository.NewDirectory(pool)
	return directoryComponents{viewers: directory, stores: directory, consignors: directory}
}

type directoryRoutes struct {
	authenticate fiber.Handler
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
		FilesBaseURL: routes.filesBaseURL,
		SearchLimiter: ratelimit.New(ratelimit.Rule{Limit: searchesPerMinutePerIP, Window: time.Minute},
			routes.clock),
		SecurityLog: routes.securityLog,
	})
}
