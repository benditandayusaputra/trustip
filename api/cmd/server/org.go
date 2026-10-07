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
	founder    orgservice.OrganizationFounder
	profiles   orgservice.ProfileStore
	stores     orgservice.StoreProfileStore
	categories orgservice.CategoryReader
	regions    orgservice.RegionReader
	tx         orgservice.Transactor
	clock      orgservice.Clock
	ids        orgservice.IDGenerator
}

func buildOrgComponents(pool *database.Pool, clock orgservice.Clock) orgComponents {
	profiles := orgrepository.NewProfiles(pool)
	regions := orgrepository.NewRegions(pool)
	return orgComponents{
		categories: regions,
		founder:    orgrepository.NewFounders(pool),
		profiles:   profiles,
		stores:     profiles,
		regions:    orgrepository.NewRegions(pool),
		tx:         pool,
		clock:      clock,
		ids:        idgen.UUIDv7{},
	}
}

type orgRoutes struct {
	authenticate fiber.Handler
	filesBaseURL string
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
		Settings:     orgservice.NewGetSettings(components.profiles),
		EditProfile:  orgservice.NewUpdateProfile(components.profiles, components.tx, components.clock),
		FilesBaseURL: routes.filesBaseURL,
		EditStore: orgservice.NewUpdateStoreProfile(orgservice.StoreProfileDeps{
			Profiles: components.profiles, Stores: components.stores, Tx: components.tx, Clock: components.clock,
		}),
		Categories:   orgservice.NewListCategories(components.categories),
		Provinces:    orgservice.NewListProvinces(components.regions),
		Cities:       orgservice.NewListCities(components.regions),
		WriteLimiter: routes.writeLimiter,
		SecurityLog:  routes.securityLog,
	}
}
