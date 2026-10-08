package main

import (
	stockhandler "github.com/benditandayusaputra/trustip/api/internal/modules/stock/handler"
	stockservice "github.com/benditandayusaputra/trustip/api/internal/modules/stock/service"
)

func stockDeps(stocks stockservice.StockReader, routes tradeRoutes) stockhandler.Deps {
	return stockhandler.Deps{
		Authenticate: routes.authenticate,
		Partnership:  stockservice.NewGetPartnershipStock(stocks, routes.clock),
		Organization: stockservice.NewListStock(stocks, routes.clock),
		Card:         stockservice.NewGetStockCard(stocks),
	}
}
