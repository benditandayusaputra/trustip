package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/modules/report/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Reports) DailyRows(ctx context.Context, query domain.DailyQuery) ([]domain.RecapRow, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).DailyRecap(readCtx, sqlc.DailyRecapParams{
		StoreOrgID: query.OrganizationID, FromAt: query.From, ToAt: query.To,
	})
	if err != nil {
		return nil, fmt.Errorf("recap %s of %s: %w", query.Day, query.OrganizationID, err)
	}
	recap := make([]domain.RecapRow, 0, len(rows))
	for _, row := range rows {
		recap = append(recap, domain.RecapRow{
			PartnerID: row.PartnerID, Name: row.Name, Paid: money.Rupiah(row.PaidAmount),
			Figures: domain.Figures{
				SoldQty: int(row.SoldQty), ReturnedQty: int(row.ReturnedQty), ReceivedQty: int(row.ReceivedQty),
				Split: pricing.Split{
					Gross: money.Rupiah(row.GrossAmount), StoreShare: money.Rupiah(row.StoreShareAmount),
					ConsignorShare: money.Rupiah(row.ConsignorShareAmount),
				},
			},
		})
	}
	return recap, nil
}

func (r *Reports) SettlementRows(ctx context.Context, query domain.SettlementQuery) ([]domain.SettlementRow, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).SettlementReport(readCtx, sqlc.SettlementReportParams{
		Today: query.Today, OrganizationID: query.OrganizationID, FromAt: query.From, ToAt: query.To,
	})
	if err != nil {
		return nil, fmt.Errorf("report settlements of %s: %w", query.OrganizationID, err)
	}
	found := make([]domain.SettlementRow, 0, len(rows))
	for _, row := range rows {
		found = append(found, domain.SettlementRow{
			PartnerID: row.PartnerID, Name: row.Name,
			Issued:  domain.Tally{Count: int(row.IssuedCount), Amount: money.Rupiah(row.IssuedAmount)},
			Paid:    domain.Tally{Count: int(row.PaidCount), Amount: money.Rupiah(row.PaidAmount)},
			Unpaid:  domain.Tally{Count: int(row.UnpaidCount), Amount: money.Rupiah(row.UnpaidAmount)},
			Overdue: domain.Tally{Count: int(row.OverdueCount), Amount: money.Rupiah(row.OverdueAmount)},
		})
	}
	return found, nil
}
