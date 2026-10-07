package main

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	orghandler "github.com/benditandayusaputra/trustip/api/internal/modules/org/handler"
	orgrepository "github.com/benditandayusaputra/trustip/api/internal/modules/org/repository"
	orgservice "github.com/benditandayusaputra/trustip/api/internal/modules/org/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
)

const writesPerMinutePerUser = 60

type orgComponents struct {
	founder orgservice.OrganizationFounder
	regions orgservice.RegionReader
	tx      orgservice.Transactor
	clock   orgservice.Clock
	ids     orgservice.IDGenerator
}

func buildOrgComponents(pool *database.Pool, clock orgservice.Clock) orgComponents {
	return orgComponents{
		founder: orgrepository.NewFounders(pool),
		regions: orgrepository.NewRegions(pool),
		tx:      pool,
		clock:   clock,
		ids:     idgen.UUIDv7{},
	}
}

type orgRoutes struct {
	authenticate fiber.Handler
	writeLimiter *ratelimit.Limiter
	securityLog  *slog.Logger
}

func newWriteLimiter(clock ratelimit.Clock) *ratelimit.Limiter {
	return ratelimit.New(ratelimit.Rule{Limit: writesPerMinutePerUser, Window: time.Minute}, clock)
}

func orgDeps(components orgComponents, routes orgRoutes) orghandler.Deps {
	return orghandler.Deps{
		Authenticate: routes.authenticate,
		Create: orgservice.NewCreateOrganization(orgservice.CreateDeps{
			Founder: components.founder, Tx: components.tx, Clock: components.clock, IDs: components.ids,
		}),
		Provinces:    orgservice.NewListProvinces(components.regions),
		Cities:       orgservice.NewListCities(components.regions),
		WriteLimiter: routes.writeLimiter,
		SecurityLog:  routes.securityLog,
	}
}
