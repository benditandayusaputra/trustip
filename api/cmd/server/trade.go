package main

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"

	agreementhandler "github.com/benditandayusaputra/trustip/api/internal/modules/agreement/handler"
	partnershiphandler "github.com/benditandayusaputra/trustip/api/internal/modules/partnership/handler"
	partnershiprepository "github.com/benditandayusaputra/trustip/api/internal/modules/partnership/repository"
	partnershipservice "github.com/benditandayusaputra/trustip/api/internal/modules/partnership/service"
	producthandler "github.com/benditandayusaputra/trustip/api/internal/modules/product/handler"
	proposalhandler "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/handler"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

type tradeComponents struct {
	products     productComponents
	proposals    proposalComponents
	agreements   agreementComponents
	partnerships partnershipservice.PartnershipReader
}

func buildTradeComponents(pool *database.Pool, clock serverClock, sealKeys security.ActiveKeyring) tradeComponents {
	seals := buildSealComponents(pool, sealKeys)
	agreements := buildAgreementComponents(pool, seals)
	return tradeComponents{
		products: buildProductComponents(pool, clock),
		proposals: buildProposalComponents(pool, clock, proposalCollaborators{
			keeper: newAgreementKeeper(pool, seals), current: currentAgreements{reader: agreements.reader},
			announcer: proposalAnnouncer{newOrganizationAnnouncer(pool, clock)},
		}),
		agreements:   agreements,
		partnerships: partnershiprepository.NewPartnerships(pool),
	}
}

type tradeRoutes struct {
	authenticate fiber.Handler
	filesBaseURL string
	writeLimiter httpserver.RateLimiter
	securityLog  *slog.Logger
}

func mountTrade(api fiber.Router, components tradeComponents, routes tradeRoutes) {
	producthandler.Mount(api, productDeps(components.products, productRoutes(routes)))
	negotiation := proposalRoutes{
		authenticate: routes.authenticate, writeLimiter: routes.writeLimiter, securityLog: routes.securityLog,
	}
	proposalhandler.Mount(api, proposalDeps(components.proposals, negotiation))
	agreementhandler.Mount(api, agreementDeps(components.agreements, negotiation))
	partnershiphandler.Mount(api, partnershiphandler.Deps{
		Authenticate: routes.authenticate,
		List:         partnershipservice.NewListPartnerships(components.partnerships),
		Get:          partnershipservice.NewGetPartnership(components.partnerships),
		Catalog:      partnershipservice.NewListCatalog(components.partnerships),
		FilesBaseURL: routes.filesBaseURL,
	})
}
