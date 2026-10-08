package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/report/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	dailyPath       = "/reports/daily"
	settlementsPath = "/reports/settlements"
)

type RecapReader interface {
	Execute(ctx context.Context, signedIn actor.Actor, day string) (domain.DailyRecap, error)
}

type SettlementReader interface {
	Execute(ctx context.Context, signedIn actor.Actor, filter domain.Filter) (domain.SettlementReport, error)
}

type recapRowResponse struct {
	PartnerID string `json:"partner_id"`
	Name      string `json:"name"`
	figuresResponse
	Paid int64 `json:"paid"`
}

type recapResponse struct {
	Date  string             `json:"date"`
	Total recapRowResponse   `json:"total"`
	Rows  []recapRowResponse `json:"rows"`
}

func toRecapRow(row domain.RecapRow) recapRowResponse {
	return recapRowResponse{
		PartnerID: row.PartnerID, Name: row.Name, figuresResponse: toFigures(row.Figures), Paid: row.Paid.Int64(),
	}
}

func daily(reader RecapReader) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		recap, err := reader.Execute(c.Context(), signedIn, c.Query("date"))
		if err != nil {
			return fmt.Errorf("read daily recap: %w", err)
		}
		rows := make([]recapRowResponse, 0, len(recap.Rows))
		for _, row := range recap.Rows {
			rows = append(rows, toRecapRow(row))
		}
		return httpserver.Respond(c, httpserver.Reply{Data: recapResponse{
			Date: recap.Query.Day, Total: toRecapRow(recap.Total), Rows: rows,
		}})
	}
}

type tallyResponse struct {
	Count  int   `json:"count"`
	Amount int64 `json:"amount"`
}

type settlementRowResponse struct {
	PartnerID string        `json:"partner_id"`
	Name      string        `json:"name"`
	Issued    tallyResponse `json:"issued"`
	Paid      tallyResponse `json:"paid"`
	Unpaid    tallyResponse `json:"unpaid"`
	Overdue   tallyResponse `json:"overdue"`
}

type settlementReportResponse struct {
	From  string                  `json:"from"`
	To    string                  `json:"to"`
	Total settlementRowResponse   `json:"total"`
	Rows  []settlementRowResponse `json:"rows"`
}

func toTally(tally domain.Tally) tallyResponse {
	return tallyResponse{Count: tally.Count, Amount: tally.Amount.Int64()}
}

func toSettlementRow(row domain.SettlementRow) settlementRowResponse {
	return settlementRowResponse{
		PartnerID: row.PartnerID, Name: row.Name, Issued: toTally(row.Issued), Paid: toTally(row.Paid),
		Unpaid: toTally(row.Unpaid), Overdue: toTally(row.Overdue),
	}
}

func settlementReport(reader SettlementReader) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, err := signedInActor(c)
		if err != nil {
			return err
		}
		report, err := reader.Execute(c.Context(), signedIn, domain.Filter{From: c.Query("from"), To: c.Query("to")})
		if err != nil {
			return fmt.Errorf("read settlement report: %w", err)
		}
		rows := make([]settlementRowResponse, 0, len(report.Rows))
		for _, row := range report.Rows {
			rows = append(rows, toSettlementRow(row))
		}
		return httpserver.Respond(c, httpserver.Reply{Data: settlementReportResponse{
			From: report.Query.FirstDay(), To: report.Query.LastDay(), Total: toSettlementRow(report.Total),
			Rows: rows,
		}})
	}
}
