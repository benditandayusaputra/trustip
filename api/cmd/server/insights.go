package main

import (
	insighthandler "github.com/benditandayusaputra/trustip/api/internal/modules/insight/handler"
	insightservice "github.com/benditandayusaputra/trustip/api/internal/modules/insight/service"
)

func insightDeps(histories insightservice.DropHistoryReader, routes tradeRoutes) insighthandler.Deps {
	return insighthandler.Deps{
		Authenticate: routes.authenticate,
		DropSuggest:  insightservice.NewSuggestDrops(histories, routes.clock),
	}
}
