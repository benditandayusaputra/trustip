package main

import (
	auditrepository "github.com/benditandayusaputra/trustip/api/internal/modules/audit/repository"
	needhandler "github.com/benditandayusaputra/trustip/api/internal/modules/need/handler"
	needrepository "github.com/benditandayusaputra/trustip/api/internal/modules/need/repository"
	needservice "github.com/benditandayusaputra/trustip/api/internal/modules/need/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

func buildNeedDeps(pool *database.Pool, clock serverClock) needservice.Deps {
	needs := needrepository.NewNeeds(pool)
	return needservice.Deps{
		Writer: needs, Reader: needs, Audit: newAuditTrail(auditrepository.NewRecords(pool)), Tx: pool,
		Clock: clock, IDs: idgen.UUIDv7{},
	}
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
