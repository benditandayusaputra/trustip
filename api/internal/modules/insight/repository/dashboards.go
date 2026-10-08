package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/modules/insight/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

type Dashboards struct {
	pool *database.Pool
}

func NewDashboards(pool *database.Pool) *Dashboards {
	return &Dashboards{pool: pool}
}

func (r *Dashboards) ReadSnapshot(ctx context.Context, query domain.SnapshotQuery) (domain.Snapshot, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	stockRows, err := queries.ListInsightStock(readCtx, sqlc.ListInsightStockParams{
		OrganizationID: query.OrganizationID, SalesSince: query.SalesSince, Now: query.Now,
	})
	if err != nil {
		return domain.Snapshot{}, fmt.Errorf("list insight stock of %s: %w", query.OrganizationID, err)
	}
	settlementRows, err := queries.ListOpenSettlements(readCtx, query.OrganizationID)
	if err != nil {
		return domain.Snapshot{}, fmt.Errorf("list open settlements of %s: %w", query.OrganizationID, err)
	}
	settlements, err := openSettlementsOf(settlementRows)
	if err != nil {
		return domain.Snapshot{}, err
	}
	stock := stockOf(stockRows)
	return domain.Snapshot{
		Products: stock.products, Partnerships: stock.partnerships, Settlements: settlements,
	}, nil
}

func openSettlementsOf(rows []sqlc.ListOpenSettlementsRow) ([]domain.OpenSettlement, error) {
	settlements := make([]domain.OpenSettlement, 0, len(rows))
	for _, row := range rows {
		dueOn, err := time.ParseInLocation(time.DateOnly, row.DueDate, period.Jakarta())
		if err != nil {
			return nil, fmt.Errorf("read due date of settlement %s: %w", row.ID, err)
		}
		settlements = append(settlements, domain.OpenSettlement{
			ID: row.ID, Number: row.Number, PartnershipID: row.PartnershipID, CounterpartName: row.CounterpartName,
			DueOn: dueOn, Outstanding: money.Rupiah(row.ConsignorShareAmount),
		})
	}
	return settlements, nil
}

func (r *Dashboards) ReadDashboard(ctx context.Context, query domain.DashboardQuery) (domain.Dashboard, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	figures, err := queries.GetDashboardFigures(readCtx, sqlc.GetDashboardFiguresParams{
		OrganizationID: query.OrganizationID, Today: query.Today, DueSoonUntil: query.DueSoonUntil,
		PendingSince: query.PendingSince, NeedsExpiringUntil: query.NeedsExpiring,
	})
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("read dashboard figures of %s: %w", query.OrganizationID, err)
	}
	uncounted, err := queries.ListUncountedPartners(readCtx, sqlc.ListUncountedPartnersParams{
		OrganizationID: query.OrganizationID, TodayStart: query.TodayStart,
	})
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("list uncounted partners of %s: %w", query.OrganizationID, err)
	}
	return domain.Dashboard{
		Figures: figuresOf(figures), Tasks: tasksOf(figures), QuickCount: quickCountOf(uncounted),
	}, nil
}

func figuresOf(row sqlc.GetDashboardFiguresRow) domain.Figures {
	return domain.Figures{
		ActivePartners: int(row.ActivePartners), StockUnits: int(row.StockUnits),
		Unpaid:  domain.Tally{Count: int(row.UnpaidCount), Amount: money.Rupiah(row.UnpaidAmount)},
		Overdue: domain.Tally{Count: int(row.OverdueCount), Amount: money.Rupiah(row.OverdueAmount)},
		DueSoon: domain.Tally{Count: int(row.DueSoonCount), Amount: money.Rupiah(row.DueSoonAmount)},
	}
}

func tasksOf(row sqlc.GetDashboardFiguresRow) domain.Tasks {
	return domain.Tasks{
		DropsToConfirm: int(row.DropsToConfirm), CountsToConfirm: int(row.CountsToConfirm),
		PaymentsToReview: int(row.PaymentsToReview), PaymentsRejected: int(row.PaymentsRejected),
		ProposalsAwaiting: int(row.ProposalsAwaiting), NeedsExpiring: int(row.NeedsExpiring),
	}
}

func quickCountOf(rows []sqlc.ListUncountedPartnersRow) []domain.UncountedPartner {
	partners := make([]domain.UncountedPartner, 0, len(rows))
	for _, row := range rows {
		partners = append(partners, domain.UncountedPartner{
			PartnershipID: row.PartnershipID, CounterpartName: row.CounterpartName, StockUnits: int(row.StockUnits),
		})
	}
	return partners
}

func (r *Dashboards) ListOpenProposalCounterparts(ctx context.Context, organizationID string) ([]string, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	counterparts, err := sqlc.New(r.pool.Querier(ctx)).ListOpenProposalCounterparts(readCtx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list open proposal counterparts of %s: %w", organizationID, err)
	}
	return counterparts, nil
}
