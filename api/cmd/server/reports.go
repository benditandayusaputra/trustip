package main

import (
	reporthandler "github.com/benditandayusaputra/trustip/api/internal/modules/report/handler"
	reportrepository "github.com/benditandayusaputra/trustip/api/internal/modules/report/repository"
	reportservice "github.com/benditandayusaputra/trustip/api/internal/modules/report/service"
)

func reportDeps(reports *reportrepository.Reports, routes tradeRoutes) reporthandler.Deps {
	return reporthandler.Deps{
		Authenticate: routes.authenticate,
		Sales:        reportservice.NewGetSalesReport(reports),
	}
}
