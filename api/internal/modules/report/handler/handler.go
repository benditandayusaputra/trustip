package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/modules/report/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const salesPath = "/reports/sales"

type SalesReader interface {
	Execute(ctx context.Context, signedIn actor.Actor, filter domain.Filter) (domain.Report, error)
}

type Deps struct {
	Authenticate fiber.Handler
	Sales        SalesReader
	Daily        RecapReader
	Settlements  SettlementReader
}

func Mount(api fiber.Router, deps Deps) {
	api.Get(salesPath, deps.Authenticate, sales(deps.Sales))
	api.Get(dailyPath, deps.Authenticate, daily(deps.Daily))
	api.Get(settlementsPath, deps.Authenticate, settlementReport(deps.Settlements))
}

func signedInActor(c fiber.Ctx) (actor.Actor, error) {
	signedIn, ok := httpserver.ActorFrom(c)
	if !ok {
		return actor.Actor{}, apperror.New(apperror.CodeUnauthenticated)
	}
	return signedIn, nil
}

func sales(reader SalesReader) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		report, err := reader.Execute(c.Context(), signedIn, domain.Filter{
			From: c.Query("from"), To: c.Query("to"), GroupBy: c.Query("group_by"),
			PartnerID: c.Query("partner_id"), ProductID: c.Query("product_id"),
		})
		if err != nil {
			return fmt.Errorf("read sales report: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toReportResponse(report)})
	}
}

type splitResponse struct {
	Gross          int64 `json:"gross"`
	StoreShare     int64 `json:"store_share"`
	ConsignorShare int64 `json:"consignor_share"`
}

type figuresResponse struct {
	SoldQty     int           `json:"sold_qty"`
	ReturnedQty int           `json:"returned_qty"`
	ReceivedQty int           `json:"received_qty"`
	ReturnedBPS int           `json:"returned_bps"`
	Split       splitResponse `json:"split"`
}

type rowResponse struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	figuresResponse
}

type reportResponse struct {
	From    string          `json:"from"`
	To      string          `json:"to"`
	GroupBy string          `json:"group_by"`
	Summary figuresResponse `json:"summary"`
	Rows    []rowResponse   `json:"rows"`
}

func toReportResponse(report domain.Report) reportResponse {
	rows := make([]rowResponse, 0, len(report.Rows))
	for _, row := range report.Rows {
		rows = append(rows, rowResponse{Key: row.Key, Label: row.Label, figuresResponse: toFigures(row.Figures)})
	}
	return reportResponse{
		From: report.Query.FirstDay(), To: report.Query.LastDay(), GroupBy: string(report.Query.GroupBy),
		Summary: toFigures(report.Summary), Rows: rows,
	}
}

func toFigures(figures domain.Figures) figuresResponse {
	return figuresResponse{
		SoldQty: figures.SoldQty, ReturnedQty: figures.ReturnedQty, ReceivedQty: figures.ReceivedQty,
		ReturnedBPS: figures.ReturnedBPS(), Split: toSplit(figures.Split),
	}
}

func toSplit(split pricing.Split) splitResponse {
	return splitResponse{
		Gross: split.Gross.Int64(), StoreShare: split.StoreShare.Int64(), ConsignorShare: split.ConsignorShare.Int64(),
	}
}
