package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Drops) GetDrop(ctx context.Context, key domain.DropKey) (domain.View, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.GetDrop(readCtx, sqlc.GetDropParams{ID: key.DropID, OrganizationID: key.OrganizationID})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.View{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.View{}, fmt.Errorf("get drop %s: %w", key.DropID, err)
	}
	rows, err := queries.ListDropItems(readCtx, key.DropID)
	if err != nil {
		return domain.View{}, fmt.Errorf("list items of drop %s: %w", key.DropID, err)
	}
	view := viewOf(row)
	for _, itemRow := range rows {
		item := itemOf(itemRow)
		view.Items = append(view.Items, item)
		view.ItemViews = append(view.ItemViews, domain.ItemView{
			Item: item, ProductName: itemRow.ProductName, SKU: itemRow.Sku, Unit: itemRow.Unit,
		})
	}
	return view, nil
}

func viewOf(row sqlc.GetDropRow) domain.View {
	return domain.View{
		Drop: domain.Drop{
			ID: row.ID, Number: row.Number, PartnershipID: row.PartnershipID, AgreementID: row.AgreementID,
			AgreementVersion: int(row.AgreementVersion), StoreID: row.StoreOrgID, ConsignorID: row.ConsignorOrgID,
			Status: domain.Status(row.Status), OccurredAt: row.OccurredAt, ClientRecordedAt: row.ClientRecordedAt,
			ReceivedAt: row.ReceivedAt, RecordedOffline: row.RecordedOffline, RecordedBy: row.RecordedBy,
			Items: []domain.Item{},
		},
		StoreName: row.StoreName, ConsignorName: row.ConsignorName, RecordedByName: row.RecordedByName,
		ConfirmedByName: row.ConfirmedByName, ConfirmedAt: row.ConfirmedAt, IsConfirmed: row.IsConfirmed,
		AutoConfirmed: row.AutoConfirmed, CancelledByName: row.CancelledByName, CancelledAt: row.CancelledAt,
		ItemViews: []domain.ItemView{},
	}
}

func itemOf(row sqlc.ListDropItemsRow) domain.Item {
	expiry := stock.NoExpiry()
	if date, err := time.Parse(time.DateOnly, row.ExpiryDate); err == nil {
		expiry = stock.ExpiresOn(date)
	}
	return domain.Item{
		ID: row.ID, ProductID: row.ProductID, BatchCode: row.BatchCode, Expiry: expiry, Qty: int(row.QtyReceived),
		Terms: termsOf(termsColumns{
			model: row.PricingModel, retail: row.RetailPrice, commission: row.CommissionBps, deposit: row.DepositPrice,
		}),
	}
}

func moneyOf(amount int64) money.Rupiah {
	return money.Rupiah(amount)
}
