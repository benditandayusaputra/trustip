package main

import (
	insighthandler "github.com/benditandayusaputra/trustip/api/internal/modules/insight/handler"
	insightservice "github.com/benditandayusaputra/trustip/api/internal/modules/insight/service"
)

type insightComponents struct {
	histories  insightservice.DropHistoryReader
	dashboards insightservice.DashboardReader
}

func insightDeps(components insightComponents, routes tradeRoutes) insighthandler.Deps {
	return insighthandler.Deps{
		Authenticate: routes.authenticate,
		DropSuggest:  insightservice.NewSuggestDrops(components.histories, routes.clock),
		Insights:     insightservice.NewListInsights(components.dashboards, routes.clock),
		Dashboard:    insightservice.NewGetDashboard(components.dashboards, routes.clock),
	}
}
