package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/report/domain"
)

type SalesReader interface {
	SalesRows(ctx context.Context, query domain.Query) ([]domain.Row, error)
}

type GetSalesReport struct {
	sales SalesReader
}

func NewGetSalesReport(sales SalesReader) *GetSalesReport {
	return &GetSalesReport{sales: sales}
}

func (uc *GetSalesReport) Execute(
	ctx context.Context, signedIn actor.Actor, filter domain.Filter,
) (domain.Report, error) {
	if err := policy.Require(signedIn, policy.ViewReports); err != nil {
		return domain.Report{}, fmt.Errorf("check report reader: %w", err)
	}
	query, err := domain.NewQuery(filter)
	if err != nil {
		return domain.Report{}, fmt.Errorf("read report filter: %w", err)
	}
	query.OrganizationID = signedIn.Membership.OrganizationID
	rows, err := uc.sales.SalesRows(ctx, query)
	if err != nil {
		return domain.Report{}, fmt.Errorf("read sales rows: %w", err)
	}
	if query.GroupBy == domain.ByDay {
		rows = domain.FillDays(query, rows)
	}
	return domain.Report{Query: query, Rows: rows, Summary: domain.Summarize(rows)}, nil
}
