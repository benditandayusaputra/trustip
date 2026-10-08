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
	stores directoryservice.StoreDirectory
}

func buildDirectoryComponents(pool *database.Pool) directoryComponents {
	return directoryComponents{stores: directoryrepository.NewDirectory(pool)}
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
		ListStores:   directoryservice.NewListStores(components.stores, routes.clock),
		FilesBaseURL: routes.filesBaseURL,
		SearchLimiter: ratelimit.New(ratelimit.Rule{Limit: searchesPerMinutePerIP, Window: time.Minute},
			routes.clock),
		SecurityLog: routes.securityLog,
	})
}
