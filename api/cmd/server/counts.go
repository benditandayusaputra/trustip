package main

import (
	"context"

	auditrepository "github.com/benditandayusaputra/trustip/api/internal/modules/audit/repository"
	countdomain "github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
	counthandler "github.com/benditandayusaputra/trustip/api/internal/modules/count/handler"
	countrepository "github.com/benditandayusaputra/trustip/api/internal/modules/count/repository"
	countservice "github.com/benditandayusaputra/trustip/api/internal/modules/count/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

type countSealer struct {
	documentSealer
}

func (s countSealer) Seal(ctx context.Context, request countdomain.SealRequest) error {
	return s.seal(ctx, sealRequest(request))
}

type countSealChecker struct {
	documentSealChecker
}

func (c countSealChecker) Check(ctx context.Context, check countdomain.SealCheck) (countdomain.SealVerdict, error) {
	isValid, trail, err := c.check(ctx, sealCheck(check))
	if err != nil {
		return countdomain.SealVerdict{}, err
	}
	events := make([]countdomain.SealEvent, 0, len(trail))
	for _, event := range trail {
		events = append(events, countdomain.SealEvent(event))
	}
	return countdomain.SealVerdict{IsValid: isValid, Events: events}, nil
}

type countAnnouncer struct {
	organizationAnnouncer
}

func (a countAnnouncer) Announce(ctx context.Context, notice countdomain.Notice) error {
	return a.announce(ctx, organizationNotice(notice))
}

func buildCountDeps(base foundation, seals sealComponents) countservice.Deps {
	pool, clock := base.pool, base.clock
	return countservice.Deps{
		Counts:    countrepository.NewCounts(pool),
		Sealer:    countSealer{documentSealer{recorder: seals.recorder}},
		Seals:     countSealChecker{documentSealChecker{verifier: seals.verifier}},
		Announcer: countAnnouncer{newOrganizationAnnouncer(base)},
		Audit:     newAuditTrail(auditrepository.NewRecords(pool)),
		Tx:        pool,
		Clock:     clock,
		IDs:       idgen.UUIDv7{},
	}
}

func countDeps(deps countservice.Deps, routes tradeRoutes) counthandler.Deps {
	return counthandler.Deps{
		Authenticate: routes.authenticate,
		Record:       countservice.NewRecordCount(deps),
		Get:          countservice.NewGetCount(deps),
		Sheet:        countservice.NewGetSheet(deps),
		Confirm:      countservice.NewConfirmCount(deps),
		List:         countservice.NewListCounts(deps),
		WriteLimiter: routes.writeLimiter,
		SecurityLog:  routes.securityLog,
	}
}
