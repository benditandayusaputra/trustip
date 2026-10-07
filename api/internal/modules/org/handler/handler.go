package handler

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	organizationsPath = "/organizations"
	organizationPath  = organizationsPath + "/:id"
	provincesPath     = "/provinces"
	citiesPath        = "/cities"
)

type Creator interface {
	Execute(ctx context.Context, signedIn actor.Actor, form domain.OnboardingForm) (domain.Organization, error)
}

type ProvinceLister interface {
	Execute(ctx context.Context) ([]domain.Province, error)
}

type CityLister interface {
	Execute(ctx context.Context, provinceCode string) ([]domain.City, error)
}

type Deps struct {
	Authenticate fiber.Handler
	Create       Creator
	Settings     SettingsReader
	EditProfile  ProfileEditor
	FilesBaseURL string
	Provinces    ProvinceLister
	Cities       CityLister
	WriteLimiter httpserver.RateLimiter
	SecurityLog  *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	writeLimit := httpserver.RateLimit(deps.WriteLimiter, httpserver.BySignedInUser, deps.SecurityLog)
	api.Post(organizationsPath, deps.Authenticate, writeLimit, createOrganization(deps.Create))
	api.Get(organizationPath, deps.Authenticate, readSettings(deps.Settings, deps.FilesBaseURL))
	api.Patch(organizationPath, deps.Authenticate, writeLimit, updateProfile(deps.EditProfile, deps.FilesBaseURL))
	api.Get(provincesPath, listProvinces(deps.Provinces))
	api.Get(citiesPath, listCities(deps.Cities))
}

func PublicRoutes() []httpserver.PublicRoute {
	return []httpserver.PublicRoute{
		{Method: fiber.MethodGet, Path: httpserver.APIPrefix + provincesPath},
		{Method: fiber.MethodGet, Path: httpserver.APIPrefix + citiesPath},
	}
}
