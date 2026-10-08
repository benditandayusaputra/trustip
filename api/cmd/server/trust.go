package main

import (
	"time"

	trusthandler "github.com/benditandayusaputra/trustip/api/internal/modules/trust/handler"
	trustservice "github.com/benditandayusaputra/trustip/api/internal/modules/trust/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/ratelimit"
)

const reportsPerHourPerUser = 5

func trustDeps(reports trustservice.ReportStore, routes tradeRoutes) trusthandler.Deps {
	return trusthandler.Deps{
		Authenticate: routes.authenticate,
		Report: trustservice.NewReportListing(trustservice.ReportDeps{
			Reports: reports, Clock: routes.clock, IDs: idgen.UUIDv7{},
		}),
		ReportLimiter: ratelimit.New(ratelimit.Rule{Limit: reportsPerHourPerUser, Window: time.Hour}, routes.clock),
		SecurityLog:   routes.securityLog,
	}
}
