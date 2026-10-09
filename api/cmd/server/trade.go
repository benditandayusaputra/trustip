package main

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"

	agreementhandler "github.com/benditandayusaputra/trustip/api/internal/modules/agreement/handler"
	counthandler "github.com/benditandayusaputra/trustip/api/internal/modules/count/handler"
	countservice "github.com/benditandayusaputra/trustip/api/internal/modules/count/service"
	drophandler "github.com/benditandayusaputra/trustip/api/internal/modules/drop/handler"
	dropservice "github.com/benditandayusaputra/trustip/api/internal/modules/drop/service"
	insightrepository "github.com/benditandayusaputra/trustip/api/internal/modules/insight/repository"
	partnershiphandler "github.com/benditandayusaputra/trustip/api/internal/modules/partnership/handler"
	partnershiprepository "github.com/benditandayusaputra/trustip/api/internal/modules/partnership/repository"
	partnershipservice "github.com/benditandayusaputra/trustip/api/internal/modules/partnership/service"
	producthandler "github.com/benditandayusaputra/trustip/api/internal/modules/product/handler"
	proposalhandler "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/handler"
	reporthandler "github.com/benditandayusaputra/trustip/api/internal/modules/report/handler"
	reportrepository "github.com/benditandayusaputra/trustip/api/internal/modules/report/repository"
	settlementhandler "github.com/benditandayusaputra/trustip/api/internal/modules/settlement/handler"
	settlementservice "github.com/benditandayusaputra/trustip/api/internal/modules/settlement/service"
	stockhandler "github.com/benditandayusaputra/trustip/api/internal/modules/stock/handler"
	stockrepository "github.com/benditandayusaputra/trustip/api/internal/modules/stock/repository"
	stockservice "github.com/benditandayusaputra/trustip/api/internal/modules/stock/service"
	trustrepository "github.com/benditandayusaputra/trustip/api/internal/modules/trust/repository"
	trustservice "github.com/benditandayusaputra/trustip/api/internal/modules/trust/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

type tradeComponents struct {
	products     productComponents
	proposals    proposalComponents
	agreements   agreementComponents
	partnerships partnershipservice.PartnershipReader
	drops        dropservice.Deps
	stocks       stockservice.StockReader
	counts       countservice.Deps
	settlements  settlementservice.Deps
	ending       partnershipservice.EndDeps
	reports      *reportrepository.Reports
	insights     insightComponents
	listings     trustservice.ReportStore
	seals        sealComponents
}

func buildTradeComponents(base foundation, sealKeys security.ActiveKeyring) tradeComponents {
	pool, clock := base.pool, base.clock
	seals := buildSealComponents(pool, sealKeys)
	agreements := buildAgreementComponents(pool, seals)
	settlements := buildSettlementDeps(base, seals)
	counts := buildCountDeps(base, seals)
	counts.Issuer = countSettlementIssuer{issuer: settlementservice.NewIssuer(settlements)}
	return tradeComponents{
		products: buildProductComponents(pool, clock),
		proposals: buildProposalComponents(pool, clock, proposalCollaborators{
			keeper: newAgreementKeeper(pool, seals), current: currentAgreements{reader: agreements.reader},
			announcer: proposalAnnouncer{newOrganizationAnnouncer(base)},
		}),
		agreements:   agreements,
		partnerships: partnershiprepository.NewPartnerships(pool),
		drops:        buildDropDeps(base, seals),
		stocks:       stockrepository.NewStocks(pool),
		counts:       counts,
		settlements:  settlements,
		ending:       buildEndingDeps(base, endingParts{seals: seals, settlements: settlements}),
		reports:      reportrepository.NewReports(pool),
		listings:     trustrepository.NewReports(pool),
		seals:        seals,
		insights: insightComponents{
			histories: insightrepository.NewHistories(pool), dashboards: insightrepository.NewDashboards(pool),
		},
	}
}

type tradeRoutes struct {
	authenticate fiber.Handler
	visit        fiber.Handler
	filesBaseURL string
	writeLimiter httpserver.RateLimiter
	securityLog  *slog.Logger
	clock        serverClock
}

func mountTrade(api fiber.Router, components tradeComponents, routes tradeRoutes) {
	producthandler.Mount(api, productDeps(components.products, productRoutes{
		authenticate: routes.authenticate, filesBaseURL: routes.filesBaseURL, writeLimiter: routes.writeLimiter,
		securityLog: routes.securityLog,
	}))
	negotiation := proposalRoutes{
		authenticate: routes.authenticate, writeLimiter: routes.writeLimiter, securityLog: routes.securityLog,
	}
	proposalhandler.Mount(api, proposalDeps(components.proposals, negotiation))
	agreementhandler.Mount(api, agreementDeps(components.agreements, negotiation))
	partnershiphandler.Mount(api, partnershipDeps(components, routes))
	drophandler.Mount(api, dropDeps(components.drops, routes))
	stockhandler.Mount(api, stockDeps(components.stocks, routes))
	counthandler.Mount(api, countDeps(components.counts, routes))
	settlementhandler.Mount(api, settlementDeps(components.settlements, routes))
	reporthandler.Mount(api, reportDeps(components.reports, routes))
}
