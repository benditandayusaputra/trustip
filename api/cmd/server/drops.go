package main

import (
	"context"

	auditrepository "github.com/benditandayusaputra/trustip/api/internal/modules/audit/repository"
	dropdomain "github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
	drophandler "github.com/benditandayusaputra/trustip/api/internal/modules/drop/handler"
	droprepository "github.com/benditandayusaputra/trustip/api/internal/modules/drop/repository"
	dropservice "github.com/benditandayusaputra/trustip/api/internal/modules/drop/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

type dropSealer struct {
	documentSealer
}

func (s dropSealer) Seal(ctx context.Context, request dropdomain.SealRequest) error {
	return s.seal(ctx, sealRequest(request))
}

type dropSealChecker struct {
	documentSealChecker
}

func (c dropSealChecker) Check(ctx context.Context, check dropdomain.SealCheck) (dropdomain.SealVerdict, error) {
	isValid, trail, err := c.check(ctx, sealCheck(check))
	if err != nil {
		return dropdomain.SealVerdict{}, err
	}
	events := make([]dropdomain.SealEvent, 0, len(trail))
	for _, event := range trail {
		events = append(events, dropdomain.SealEvent(event))
	}
	return dropdomain.SealVerdict{IsValid: isValid, Events: events}, nil
}

type dropAnnouncer struct {
	organizationAnnouncer
}

func (a dropAnnouncer) Announce(ctx context.Context, notice dropdomain.Notice) error {
	return a.announce(ctx, organizationNotice(notice))
}

func buildDropDeps(pool *database.Pool, clock serverClock, seals sealComponents) dropservice.Deps {
	return dropservice.Deps{
		Drops:     droprepository.NewDrops(pool),
		Sealer:    dropSealer{documentSealer{recorder: seals.recorder}},
		Seals:     dropSealChecker{documentSealChecker{verifier: seals.verifier}},
		Announcer: dropAnnouncer{newOrganizationAnnouncer(pool, clock)},
		Audit:     newAuditTrail(auditrepository.NewRecords(pool)),
		Tx:        pool,
		Clock:     clock,
		IDs:       idgen.UUIDv7{},
	}
}

func dropDeps(deps dropservice.Deps, routes tradeRoutes) drophandler.Deps {
	return drophandler.Deps{
		Authenticate: routes.authenticate,
		Record:       dropservice.NewRecordDrop(deps),
		Get:          dropservice.NewGetDrop(deps),
		Clock:        routes.clock,
		WriteLimiter: routes.writeLimiter,
		SecurityLog:  routes.securityLog,
	}
}
