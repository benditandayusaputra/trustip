package main

import (
	"context"

	auditrepository "github.com/benditandayusaputra/trustip/api/internal/modules/audit/repository"
	needdomain "github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
	needhandler "github.com/benditandayusaputra/trustip/api/internal/modules/need/handler"
	needrepository "github.com/benditandayusaputra/trustip/api/internal/modules/need/repository"
	needservice "github.com/benditandayusaputra/trustip/api/internal/modules/need/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

func buildNeedDeps(pool *database.Pool, clock serverClock) needservice.Deps {
	needs := needrepository.NewNeeds(pool)
	return needservice.Deps{
		Writer: needs, Reader: needs, Interest: needs, Announcer: needAnnouncer{newOrganizationAnnouncer(pool, clock)},
		Audit: newAuditTrail(auditrepository.NewRecords(pool)), Tx: pool, Clock: clock, IDs: idgen.UUIDv7{},
	}
}

type needAnnouncer struct {
	organizationAnnouncer
}

func (a needAnnouncer) Announce(ctx context.Context, notice needdomain.Notice) error {
	return a.announce(ctx, organizationNotice(notice))
}

func needDeps(deps needservice.Deps, routes directoryRoutes) needhandler.Deps {
	return needhandler.Deps{
		Authenticate: routes.authenticate,
		Post:         needservice.NewPostNeed(deps),
		Get:          needservice.NewGetNeed(deps.Reader, deps.Clock),
		List:         needservice.NewListNeeds(deps.Reader, deps.Clock),
		Extend:       needservice.NewExtendNeed(deps),
		Fulfill:      needservice.NewFulfillNeed(deps),
		Close:        needservice.NewCloseNeed(deps),
		WriteLimiter: routes.writeLimiter,
		SecurityLog:  routes.securityLog,
	}
}
