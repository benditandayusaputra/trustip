package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/report/domain"
)

const dateLayout = "2006-01-02"

type SettlementReader interface {
	SettlementRows(ctx context.Context, query domain.SettlementQuery) ([]domain.SettlementRow, error)
}

type GetSettlementReport struct {
	settlements SettlementReader
	clock       Clock
}

func NewGetSettlementReport(settlements SettlementReader, clock Clock) *GetSettlementReport {
	return &GetSettlementReport{settlements: settlements, clock: clock}
}

func (uc *GetSettlementReport) Execute(
	ctx context.Context, signedIn actor.Actor, filter domain.Filter,
) (domain.SettlementReport, error) {
	if err := policy.Require(signedIn, policy.ViewReports); err != nil {
		return domain.SettlementReport{}, fmt.Errorf("check report reader: %w", err)
	}
	filter.GroupBy = string(domain.ByPartner)
	query, err := domain.NewQuery(filter)
	if err != nil {
		return domain.SettlementReport{}, fmt.Errorf("read report filter: %w", err)
	}
	query.OrganizationID = signedIn.Membership.OrganizationID
	settlementQuery := domain.SettlementQuery{
		Query: query, Today: uc.clock.Now().In(period.Jakarta()).Format(dateLayout),
	}
	rows, err := uc.settlements.SettlementRows(ctx, settlementQuery)
	if err != nil {
		return domain.SettlementReport{}, fmt.Errorf("read settlement rows: %w", err)
	}
	return domain.SettlementReport{
		Query: settlementQuery, Rows: rows, Total: domain.SummarizeSettlements(rows),
	}, nil
}
