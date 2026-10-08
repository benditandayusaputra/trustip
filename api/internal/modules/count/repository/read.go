package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Counts) GetCount(ctx context.Context, key domain.CountKey) (domain.View, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.GetCount(readCtx, sqlc.GetCountParams{ID: key.CountID, OrganizationID: key.OrganizationID})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.View{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.View{}, fmt.Errorf("get count %s: %w", key.CountID, err)
	}
	view := viewOf(row)
	if err := withLines(readCtx, queries, &view); err != nil {
		return domain.View{}, err
	}
	settlements, err := queries.ListCountSettlements(readCtx, view.ID)
	if err != nil {
		return domain.View{}, fmt.Errorf("find settlement of count %s: %w", view.ID, err)
	}
	for _, settlement := range settlements {
		view.Settlement = domain.SettlementRef{ID: settlement.ID, Number: settlement.Number}
	}
	return view, nil
}

func viewOf(row sqlc.GetCountRow) domain.View {
	return domain.View{
		Count: domain.Count{
			ID: row.ID, Number: row.Number, PartnershipID: row.PartnershipID, StoreID: row.StoreOrgID,
			ConsignorID: row.ConsignorOrgID, Status: domain.Status(row.Status), Occasion: domain.Occasion(row.Occasion),
			Note: row.Note, CountedAt: row.CountedAt, ClientRecordedAt: row.ClientRecordedAt,
			ReceivedAt: row.ReceivedAt, RecordedOffline: row.RecordedOffline, RecordedBy: row.RecordedBy,
			SoldQty: int(row.SoldQty), ReturnedQty: int(row.ReturnedQty),
			Split: pricing.Split{
				Gross: money.Rupiah(row.GrossAmount), StoreShare: money.Rupiah(row.StoreShareAmount),
				ConsignorShare: money.Rupiah(row.ConsignorShareAmount),
			},
		},
		StoreName: row.StoreName, ConsignorName: row.ConsignorName, RecordedByName: row.RecordedByName,
		ConfirmedByName: row.ConfirmedByName, ConfirmedAt: row.ConfirmedAt, IsConfirmed: row.IsConfirmed,
		AutoConfirmed: row.AutoConfirmed,
		Products:      map[string]domain.LineProduct{},
	}
}

func withLines(ctx context.Context, queries *sqlc.Queries, view *domain.View) error {
	lines, err := queries.ListCountLines(ctx, view.ID)
	if err != nil {
		return fmt.Errorf("list lines of count %s: %w", view.ID, err)
	}
	sales, err := queries.ListCountSales(ctx, view.ID)
	if err != nil {
		return fmt.Errorf("list sales of count %s: %w", view.ID, err)
	}
	returns, err := queries.ListCountReturns(ctx, view.ID)
	if err != nil {
		return fmt.Errorf("list returns of count %s: %w", view.ID, err)
	}
	for _, line := range lines {
		view.Products[line.ProductID] = domain.LineProduct{Name: line.ProductName, SKU: line.Sku, Unit: line.Unit}
		view.Lines = append(view.Lines, lineOf(line, allocations{sales: sales, returns: returns}))
	}
	return nil
}

type allocations struct {
	sales   []sqlc.ListCountSalesRow
	returns []sqlc.ReturnAllocation
}

func lineOf(row sqlc.ListCountLinesRow, found allocations) domain.CountLine {
	line := domain.CountLine{
		ID: row.ID, Ordinal: int(row.LineNo),
		Line: domain.Line{
			ProductID: row.ProductID, RecordedQty: int(row.RecordedQty), CountedQty: int(row.CountedQty),
			SoldQty: int(row.SoldQty), ReturnedQty: int(row.ReturnedQty), KeptQty: int(row.KeptQty),
			ReturnReason: domain.ReturnReason(row.ReturnReason),
		},
	}
	for _, sale := range found.sales {
		if sale.CountLineID == row.ID {
			line.Sold = append(line.Sold, saleOf(sale))
		}
	}
	for _, taken := range found.returns {
		if taken.CountLineID == row.ID {
			line.Taken = append(line.Taken, domain.ReturnRecord{
				ID: taken.ID, Allocation: stock.Allocation{BatchID: taken.DropItemID, Qty: int(taken.Qty)},
			})
		}
	}
	return line
}

func saleOf(row sqlc.ListCountSalesRow) domain.SaleRecord {
	return domain.SaleRecord{ID: row.ID, SaleAllocation: domain.SaleAllocation{
		BatchID: row.DropItemID, Qty: int(row.Qty),
		Split: pricing.Split{
			Gross: money.Rupiah(row.GrossAmount), StoreShare: money.Rupiah(row.StoreShareAmount),
			ConsignorShare: money.Rupiah(row.ConsignorShareAmount),
		},
	}}
}
