package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const nilID = "00000000-0000-0000-0000-000000000000"

func (r *Settlements) GetSettlement(ctx context.Context, key domain.SettlementKey) (domain.View, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.GetSettlement(readCtx, sqlc.GetSettlementParams{
		ID: key.SettlementID, OrganizationID: key.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.View{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.View{}, fmt.Errorf("get settlement %s: %w", key.SettlementID, err)
	}
	view, err := viewOf(row)
	if err != nil {
		return domain.View{}, err
	}
	sales, err := queries.ListSettlementAllocations(readCtx, row.ID)
	if err != nil {
		return domain.View{}, fmt.Errorf("list sales of settlement %s: %w", row.ID, err)
	}
	view = withSales(view, sales)
	view.Summary.Qty = int(row.SoldQty)
	view.Summary.Split = splitOf(row.GrossAmount, row.StoreShareAmount, row.ConsignorShareAmount)
	if row.ShowsPayee {
		view.Payee = domain.Payee{
			IsShown: true, Info: row.PaymentInfo, QRFileID: row.PaymentQrFileID, Methods: row.PaymentMethods,
			Bank: domain.BankAccount{
				Name: row.BankName, Number: row.BankAccountNumber, Holder: row.BankAccountHolder,
			},
		}
	}
	return withPayments(readCtx, queries, view)
}

func viewOf(row sqlc.GetSettlementRow) (domain.View, error) {
	dueDate, err := domain.ParseDueDate(row.DueDate)
	if err != nil {
		return domain.View{}, fmt.Errorf("settlement %s: %w", row.ID, err)
	}
	view := domain.View{
		Settlement: domain.Settlement{
			ID: row.ID, Number: row.Number, PartnershipID: row.PartnershipID, ConsignorID: row.ConsignorOrgID,
			StoreID: row.StoreOrgID, Schedule: terms.Schedule(row.Schedule),
			Period: domain.Period{Start: row.PeriodStart, End: row.PeriodEnd}, Status: domain.Status(row.Status),
			DueDate: dueDate, IssuedAt: row.IssuedAt,
		},
		ConsignorName: row.ConsignorName, StoreName: row.StoreName, Products: map[string]domain.ProductInfo{},
	}
	if view.Status == domain.Paid {
		view.PaidAt = row.PaidAt
	}
	return view, nil
}

func withSales(view domain.View, rows []sqlc.ListSettlementAllocationsRow) domain.View {
	seenCounts := map[string]bool{}
	for _, row := range rows {
		view.Allocations = append(view.Allocations, domain.Allocation{
			ID: row.ID, ProductID: row.ProductID, OccurredAt: row.OccurredAt, Qty: int(row.Qty),
			Terms: termsOf(pricing.TermsInput{
				Model: row.PricingModel, RetailPrice: row.RetailPrice, CommissionBPS: int(row.CommissionBps),
				DepositPrice: row.DepositPrice,
			}),
			Split: splitOf(row.GrossAmount, row.StoreShareAmount, row.ConsignorShareAmount),
		})
		view.Products[row.ProductID] = domain.ProductInfo{Name: row.ProductName, SKU: row.Sku, Unit: row.Unit}
		if !seenCounts[row.CountID] {
			seenCounts[row.CountID] = true
			view.Counts = append(view.Counts, domain.SourceCount{
				ID: row.CountID, Number: row.CountNumber, CountedAt: row.CountedAt,
			})
		}
	}
	view.Summary = domain.Summarize(view.Allocations)
	return view
}

func (r *Settlements) ListSettlements(ctx context.Context, query domain.ListQuery) ([]domain.ListItem, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	afterID := query.After.ID
	if query.After.IsStart() {
		afterID = nilID
	}
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListSettlements(readCtx, sqlc.ListSettlementsParams{
		OrganizationID: query.OrganizationID, PartnershipID: query.PartnershipID, Status: query.Status,
		DueFrom: query.DueFrom, DueTo: query.DueTo, FromStart: query.After.IsStart(), AfterAt: query.After.At,
		AfterID: afterID, RowLimit: int64(query.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list settlements of %s: %w", query.OrganizationID, err)
	}
	items := make([]domain.ListItem, 0, len(rows))
	for _, row := range rows {
		dueDate, err := domain.ParseDueDate(row.DueDate)
		if err != nil {
			return nil, fmt.Errorf("settlement %s: %w", row.ID, err)
		}
		items = append(items, domain.ListItem{
			ID: row.ID, Number: row.Number, PartnershipID: row.PartnershipID, Status: domain.Status(row.Status),
			Schedule: row.Schedule, Period: domain.Period{Start: row.PeriodStart, End: row.PeriodEnd},
			DueDate: dueDate, IssuedAt: row.IssuedAt, SoldQty: int(row.SoldQty),
			ConsignorShare: row.ConsignorShareAmount, CounterpartID: row.CounterpartID,
			CounterpartName: row.CounterpartName, CounterpartType: row.CounterpartType,
		})
	}
	return items, nil
}

func (r *Settlements) GetRunning(ctx context.Context, key domain.RunningKey) (domain.Running, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetRunningSales(readCtx, sqlc.GetRunningSalesParams{
		PartnershipID: key.PartnershipID, OrganizationID: key.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Running{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Running{}, fmt.Errorf("get running sales of %s: %w", key.PartnershipID, err)
	}
	running := domain.Running{
		PartnershipID: key.PartnershipID, HasSales: row.AllocationCount > 0, SoldQty: int(row.SoldQty),
		Split: splitOf(row.GrossAmount, row.StoreShareAmount, row.ConsignorShareAmount),
	}
	if running.HasSales {
		running.Since = row.Since
	}
	return running, nil
}
