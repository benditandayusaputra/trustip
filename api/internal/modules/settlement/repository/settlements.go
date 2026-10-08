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
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const dateLayout = "2006-01-02"

type Settlements struct {
	pool *database.Pool
}

func NewSettlements(pool *database.Pool) *Settlements {
	return &Settlements{pool: pool}
}

func (r *Settlements) GetLedger(ctx context.Context, partnershipID string) (domain.Ledger, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.GetSettlementLedger(readCtx, partnershipID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Ledger{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Ledger{}, fmt.Errorf("get ledger of %s: %w", partnershipID, err)
	}
	rows, err := queries.ListUnsettledAllocations(readCtx, partnershipID)
	if err != nil {
		return domain.Ledger{}, fmt.Errorf("list unsettled sales of %s: %w", partnershipID, err)
	}
	ledger := domain.Ledger{
		PartnershipID: row.ID, ConsignorID: row.ConsignorOrgID, StoreID: row.StoreOrgID, StoreName: row.StoreName,
		Schedule: terms.Schedule(row.SettlementSchedule), PaymentTermsDays: int(row.PaymentTermsDays),
		Unsettled: make([]domain.Allocation, 0, len(rows)),
	}
	if row.SettledUntil.Year() > 1 {
		ledger.SettledUntil = row.SettledUntil
	}
	for _, sale := range rows {
		ledger.Unsettled = append(ledger.Unsettled, domain.Allocation{
			ID: sale.ID, ProductID: sale.ProductID, OccurredAt: sale.OccurredAt, Qty: int(sale.Qty),
			Terms: termsOf(pricing.TermsInput{
				Model: sale.PricingModel, RetailPrice: sale.RetailPrice, CommissionBPS: int(sale.CommissionBps),
				DepositPrice: sale.DepositPrice,
			}),
			Split: splitOf(sale.GrossAmount, sale.StoreShareAmount, sale.ConsignorShareAmount),
		})
	}
	return ledger, nil
}

func termsOf(snapshot pricing.TermsInput) pricing.Terms {
	return pricing.Terms{
		Model: pricing.Model(snapshot.Model), RetailPrice: money.Rupiah(snapshot.RetailPrice),
		CommissionBPS: snapshot.CommissionBPS, DepositPrice: money.Rupiah(snapshot.DepositPrice),
	}
}

func splitOf(gross, storeShare, consignorShare int64) pricing.Split {
	return pricing.Split{
		Gross: money.Rupiah(gross), StoreShare: money.Rupiah(storeShare), ConsignorShare: money.Rupiah(consignorShare),
	}
}

func (r *Settlements) NextNumber(ctx context.Context, key domain.SequenceKey) (int, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	next, err := sqlc.New(r.pool.Querier(ctx)).NextDocumentNumber(writeCtx, sqlc.NextDocumentNumberParams{
		OrganizationID: key.OrganizationID, Prefix: key.Prefix, Period: key.Period,
	})
	if err != nil {
		return 0, fmt.Errorf("next %s number of %s: %w", key.Prefix, key.OrganizationID, err)
	}
	return int(next), nil
}

func (r *Settlements) CreateSettlement(ctx context.Context, settlement domain.Settlement) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	split := settlement.Summary.Split
	err := queries.CreateSettlement(writeCtx, sqlc.CreateSettlementParams{
		ID: settlement.ID, Number: settlement.Number, PartnershipID: settlement.PartnershipID,
		ConsignorOrgID: settlement.ConsignorID, StoreOrgID: settlement.StoreID, Schedule: string(settlement.Schedule),
		PeriodStart: settlement.Period.Start, PeriodEnd: settlement.Period.End, SoldQty: int64(settlement.Summary.Qty),
		GrossAmount: split.Gross.Int64(), StoreShareAmount: split.StoreShare.Int64(),
		ConsignorShareAmount: split.ConsignorShare.Int64(), DueDate: settlement.DueDate.Format(dateLayout),
		IssuedAt: settlement.IssuedAt,
	})
	if err != nil {
		return fmt.Errorf("create settlement %s: %w", settlement.ID, err)
	}
	return r.link(writeCtx, queries, settlement)
}

func (r *Settlements) link(ctx context.Context, queries *sqlc.Queries, settlement domain.Settlement) error {
	ids := make([]string, 0, len(settlement.Allocations))
	for _, allocation := range settlement.Allocations {
		ids = append(ids, allocation.ID)
	}
	linked, err := queries.LinkSettlementAllocations(ctx, sqlc.LinkSettlementAllocationsParams{
		SettlementID: settlement.ID, Ids: ids, PartnershipID: settlement.PartnershipID,
	})
	if err != nil {
		return fmt.Errorf("link sales to settlement %s: %w", settlement.ID, err)
	}
	if int(linked) != len(ids) {
		return fmt.Errorf("link sales to settlement %s: %w", settlement.ID, domain.ErrAlreadyLinked)
	}
	return nil
}
