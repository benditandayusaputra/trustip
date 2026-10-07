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
	sections   orgservice.ProfileSectionStore
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
		sections:   profiles,
		regions:    regions,
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
		Settings:      orgservice.NewGetSettings(components.profiles),
		EditProfile:   orgservice.NewUpdateProfile(components.profiles, components.tx, components.clock),
		FilesBaseURL:  routes.filesBaseURL,
		EditStore:     orgservice.NewUpdateStoreProfile(sectionDeps(components)),
		EditConsignor: orgservice.NewUpdateConsignorProfile(sectionDeps(components)),
		EditPayment:   orgservice.NewUpdatePaymentInfo(sectionDeps(components)),
		Categories:    orgservice.NewListCategories(components.categories),
		Provinces:     orgservice.NewListProvinces(components.regions),
		Cities:        orgservice.NewListCities(components.regions),
		WriteLimiter:  routes.writeLimiter,
		SecurityLog:   routes.securityLog,
	}
}

func sectionDeps(components orgComponents) orgservice.SectionDeps {
	return orgservice.SectionDeps{
		Profiles: components.profiles, Sections: components.sections, Tx: components.tx, Clock: components.clock,
	}
}
