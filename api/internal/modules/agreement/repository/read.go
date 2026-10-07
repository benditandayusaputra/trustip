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
	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Agreements) GetAgreement(ctx context.Context, key domain.VersionKey) (domain.View, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.GetAgreementVersion(readCtx, sqlc.GetAgreementVersionParams{
		PartnershipID: key.PartnershipID, OrganizationID: key.OrganizationID, Version: int64(key.Version),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.View{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.View{}, fmt.Errorf("get agreement of %s: %w", key.PartnershipID, err)
	}
	items, err := queries.ListAgreementItems(readCtx, row.ID)
	if err != nil {
		return domain.View{}, fmt.Errorf("list items of agreement %s: %w", row.ID, err)
	}
	view := toView(row)
	view.Agreement.Items = toItems(items)
	return view, nil
}

func toView(row sqlc.GetAgreementVersionRow) domain.View {
	return domain.View{
		IsPartnershipActive: row.PartnershipStatus == partnershipActive,
		Agreement: domain.Agreement{
			ID: row.ID, PartnershipID: row.PartnershipID, Version: int(row.Version), Status: domain.Status(row.Status),
			Terms: terms.General{
				Schedule: terms.Schedule(row.SettlementSchedule), PaymentTermsDays: int(row.PaymentTermsDays),
				Rhythm: terms.Rhythm(row.CountRhythm), Notes: row.Notes,
			},
			SourceProposalID: row.SourceProposalID, ProposalNumber: row.ProposalNumber,
			ProposedBy: row.ProposedBy, AcceptedBy: row.AcceptedBy, AcceptedAt: row.AcceptedAt,
		},
		Consignor: domain.Party{OrganizationID: row.ConsignorOrgID, Name: row.ConsignorName},
		Store:     domain.Party{OrganizationID: row.StoreOrgID, Name: row.StoreName},
		Proposer:  domain.Person{Name: row.ProposerName, OrganizationName: row.ProposerOrgName},
		Acceptor:  domain.Person{Name: row.AcceptorName, OrganizationName: row.AcceptorOrgName},
	}
}

func toItems(rows []sqlc.ListAgreementItemsRow) []domain.Item {
	items := make([]domain.Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.Item{
			ProductID: row.ProductID, ProductName: row.ProductName, SKU: row.Sku, Unit: row.Unit,
			IsActive: row.IsActive,
			Terms: pricing.Terms{
				Model: pricing.Model(row.PricingModel), RetailPrice: money.Rupiah(row.RetailPrice),
				CommissionBPS: int(row.CommissionBps), DepositPrice: money.Rupiah(row.DepositPrice),
			},
		})
	}
	return items
}
