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
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Proposals) GetProposal(ctx context.Context, key domain.ProposalKey) (domain.Proposal, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.GetProposal(readCtx, sqlc.GetProposalParams{
		ID: key.ProposalID, OrganizationID: key.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Proposal{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("get proposal %s: %w", key.ProposalID, err)
	}
	items, err := queries.ListProposalItems(readCtx, row.RevisionID)
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("list items of %s: %w", row.RevisionID, err)
	}
	proposal := toProposal(row)
	proposal.Revision.Items = toItems(items)
	return proposal, nil
}

func toProposal(row sqlc.GetProposalRow) domain.Proposal {
	return domain.Proposal{
		ID: row.ID, Number: row.Number, Kind: domain.Kind(row.Kind), Status: domain.Status(row.Status),
		Consignor:      domain.Party{OrganizationID: row.ConsignorOrgID, Name: row.ConsignorName},
		Store:          domain.Party{OrganizationID: row.StoreOrgID, Name: row.StoreName},
		FromOrgID:      row.FromOrgID,
		AwaitingOrgID:  row.AwaitingOrgID,
		InviteCode:     row.InviteCode,
		PartnershipID:  row.PartnershipID,
		DecisionReason: row.DecisionReason,
		ExpiresAt:      row.ExpiresAt,
		CreatedAt:      row.CreatedAt,
		Revision: domain.Revision{
			ID: row.RevisionID, No: int(row.CurrentRevisionNo), AuthorOrgID: row.AuthorOrgID,
			AuthorUserID: row.AuthorUserID,
			Terms: terms.General{
				Schedule: terms.Schedule(row.SettlementSchedule), PaymentTermsDays: int(row.PaymentTermsDays),
				Rhythm: terms.Rhythm(row.CountRhythm), Notes: row.Notes,
			},
			Message: row.Message, At: row.RevisedAt,
		},
	}
}

func toItems(rows []sqlc.ListProposalItemsRow) []domain.Item {
	items := make([]domain.Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.Item{
			ProductID: row.ProductID, ProductName: row.ProductName, SKU: row.Sku, Unit: row.Unit,
			Action: domain.Action(row.Action),
			Terms: pricing.Terms{
				Model: pricing.Model(row.PricingModel), RetailPrice: money.Rupiah(row.RetailPrice),
				CommissionBPS: int(row.CommissionBps), DepositPrice: money.Rupiah(row.DepositPrice),
			},
		})
	}
	return items
}

func (r *Proposals) LockProposal(ctx context.Context, key domain.ProposalKey) (domain.Proposal, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	row, err := queries.LockProposal(writeCtx, sqlc.LockProposalParams{
		ID: key.ProposalID, OrganizationID: key.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Proposal{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("lock proposal %s: %w", key.ProposalID, err)
	}
	items, err := queries.ListProposalItems(writeCtx, row.RevisionID)
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("list items of %s: %w", row.RevisionID, err)
	}
	proposal := toProposal(sqlc.GetProposalRow(row))
	proposal.Revision.Items = toItems(items)
	return proposal, nil
}

func (r *Proposals) ListProposals(ctx context.Context, query domain.ListQuery) ([]domain.Summary, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	afterID := query.After.ID
	if query.After.IsStart() {
		afterID = nilID
	}
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListProposals(readCtx, sqlc.ListProposalsParams{
		OrganizationID: query.OrganizationID, Outgoing: query.Box == domain.Outgoing, FromStart: query.After.IsStart(),
		AfterAt: query.After.At, AfterID: afterID, RowLimit: int64(query.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list proposals of %s: %w", query.OrganizationID, err)
	}
	summaries := make([]domain.Summary, 0, len(rows))
	for _, row := range rows {
		summaries = append(summaries, domain.Summary{
			ID: row.ID, Number: row.Number, Kind: domain.Kind(row.Kind), Status: domain.Status(row.Status),
			FromOrgID: row.FromOrgID, AwaitingOrgID: row.AwaitingOrgID, CounterpartName: row.CounterpartName,
			ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt,
		})
	}
	return summaries, nil
}
