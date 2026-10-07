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

func (r *Agreements) LockActiveAgreement(ctx context.Context, partnershipID string) (domain.Agreement, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.LockActiveAgreement(writeCtx, partnershipID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Agreement{}, apperror.NotFound()
	}
	if err != nil {
		return domain.Agreement{}, fmt.Errorf("lock agreement of %s: %w", partnershipID, err)
	}
	items, err := queries.ListAgreementItems(writeCtx, row.ID)
	if err != nil {
		return domain.Agreement{}, fmt.Errorf("list items of agreement %s: %w", row.ID, err)
	}
	return domain.Agreement{
		ID: row.ID, PartnershipID: row.PartnershipID, Version: int(row.Version), Status: domain.Active,
		Terms: terms.General{
			Schedule: terms.Schedule(row.SettlementSchedule), PaymentTermsDays: int(row.PaymentTermsDays),
			Rhythm: terms.Rhythm(row.CountRhythm), Notes: row.Notes,
		},
		Items: toItems(items), SourceProposalID: row.SourceProposalID, ProposalNumber: row.ProposalNumber,
		ProposedBy: row.ProposedBy, AcceptedBy: row.AcceptedBy, AcceptedAt: row.AcceptedAt,
	}, nil
}

func (r *Agreements) Supersede(ctx context.Context, agreementID string) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	superseded, err := sqlc.New(r.pool.Querier(ctx)).SupersedeAgreement(writeCtx, agreementID)
	if err != nil {
		return fmt.Errorf("supersede agreement %s: %w", agreementID, err)
	}
	if superseded == 0 {
		return apperror.New(apperror.CodeProposalDecided)
	}
	return nil
}

func (r *Agreements) ListVersions(ctx context.Context, key domain.PartnershipKey) ([]domain.Agreement, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	rows, err := queries.ListAgreementVersions(readCtx, sqlc.ListAgreementVersionsParams{
		PartnershipID: key.PartnershipID, OrganizationID: key.OrganizationID,
	})
	if err != nil {
		return nil, fmt.Errorf("list versions of %s: %w", key.PartnershipID, err)
	}
	if len(rows) == 0 {
		return nil, apperror.OutsideScope()
	}
	items, err := queries.ListPartnershipAgreementItems(readCtx, key.PartnershipID)
	if err != nil {
		return nil, fmt.Errorf("list version items of %s: %w", key.PartnershipID, err)
	}
	return toVersions(rows, items), nil
}

func toVersions(
	rows []sqlc.ListAgreementVersionsRow, items []sqlc.ListPartnershipAgreementItemsRow,
) []domain.Agreement {
	itemsByAgreement := map[string][]domain.Item{}
	for _, row := range items {
		itemsByAgreement[row.AgreementID] = append(itemsByAgreement[row.AgreementID], domain.Item{
			ProductID: row.ProductID, ProductName: row.ProductName, SKU: row.Sku, Unit: row.Unit,
			IsActive: row.IsActive,
			Terms: pricing.Terms{
				Model: pricing.Model(row.PricingModel), RetailPrice: money.Rupiah(row.RetailPrice),
				CommissionBPS: int(row.CommissionBps), DepositPrice: money.Rupiah(row.DepositPrice),
			},
		})
	}
	versions := make([]domain.Agreement, 0, len(rows))
	for _, row := range rows {
		versions = append(versions, domain.Agreement{
			ID: row.ID, Version: int(row.Version), Status: domain.Status(row.Status),
			ProposalNumber: row.ProposalNumber,
			AcceptedAt:     row.AcceptedAt, Items: itemsByAgreement[row.ID],
			Terms: terms.General{
				Schedule: terms.Schedule(row.SettlementSchedule), PaymentTermsDays: int(row.PaymentTermsDays),
				Rhythm: terms.Rhythm(row.CountRhythm), Notes: row.Notes,
			},
		})
	}
	return versions
}
