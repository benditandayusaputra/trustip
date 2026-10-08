package main

import (
	"context"

	auditrepository "github.com/benditandayusaputra/trustip/api/internal/modules/audit/repository"
	partnershipdomain "github.com/benditandayusaputra/trustip/api/internal/modules/partnership/domain"
	partnershiphandler "github.com/benditandayusaputra/trustip/api/internal/modules/partnership/handler"
	partnershiprepository "github.com/benditandayusaputra/trustip/api/internal/modules/partnership/repository"
	partnershipservice "github.com/benditandayusaputra/trustip/api/internal/modules/partnership/service"
	settlementservice "github.com/benditandayusaputra/trustip/api/internal/modules/settlement/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

type partnershipSealer struct {
	documentSealer
}

func (s partnershipSealer) Seal(ctx context.Context, request partnershipdomain.SealRequest) error {
	return s.seal(ctx, sealRequest(request))
}

type partnershipAnnouncer struct {
	organizationAnnouncer
}

func (a partnershipAnnouncer) Announce(ctx context.Context, notice partnershipdomain.Notice) error {
	return a.announce(ctx, organizationNotice(notice))
}

type endingParts struct {
	clock       serverClock
	seals       sealComponents
	settlements settlementservice.Deps
}

func buildEndingDeps(pool *database.Pool, parts endingParts) partnershipservice.EndDeps {
	partnerships := partnershiprepository.NewPartnerships(pool)
	return partnershipservice.EndDeps{
		Partnerships: partnerships,
		Reader:       partnerships,
		Settler:      outstandingSettler{issuer: settlementservice.NewIssuer(parts.settlements)},
		Sealer:       partnershipSealer{documentSealer{recorder: parts.seals.recorder}},
		Announcer:    partnershipAnnouncer{newOrganizationAnnouncer(pool, parts.clock)},
		Audit:        newAuditTrail(auditrepository.NewRecords(pool)),
		Tx:           pool,
		Clock:        parts.clock,
	}
}

func partnershipDeps(components tradeComponents, routes tradeRoutes) partnershiphandler.Deps {
	return partnershiphandler.Deps{
		Authenticate: routes.authenticate,
		List:         partnershipservice.NewListPartnerships(components.partnerships),
		Get:          partnershipservice.NewGetPartnership(components.partnerships),
		Catalog:      partnershipservice.NewListCatalog(components.partnerships),
		Ending: partnershiphandler.EndingDeps{
			Blockers:     partnershipservice.NewGetBlockers(components.ending),
			End:          partnershipservice.NewEndPartnership(components.ending),
			Listing:      partnershipservice.NewSetShownOnPublic(components.ending),
			WriteLimiter: routes.writeLimiter,
			SecurityLog:  routes.securityLog,
		},
		FilesBaseURL: routes.filesBaseURL,
	}
}
