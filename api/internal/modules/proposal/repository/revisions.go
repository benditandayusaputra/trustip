package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Proposals) SaveCounter(ctx context.Context, offer domain.Counteroffer) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	moved, err := queries.CounterProposal(writeCtx, sqlc.CounterProposalParams{
		ID: offer.ProposalID, RevisionNo: int64(offer.RevisionNo), PreviousNo: int64(offer.PreviousNo),
		AwaitingOrgID: offer.AwaitingOrgID, ExpiresAt: offer.ExpiresAt, CounteredAt: offer.At,
	})
	if err != nil {
		return fmt.Errorf("counter proposal %s: %w", offer.ProposalID, err)
	}
	if moved == 0 {
		return domain.ErrAlreadyDecided
	}
	return saveRevision(writeCtx, queries, revisionRow{
		id: offer.RevisionID, proposalID: offer.ProposalID, no: offer.RevisionNo, authorOrgID: offer.AuthorOrgID,
		authorUserID: offer.AuthorUserID, consignorID: offer.ConsignorID, draft: offer.Draft, at: offer.At,
	})
}

func (r *Proposals) ListRevisions(ctx context.Context, key domain.ProposalKey) ([]domain.Revision, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	rows, err := queries.ListProposalRevisions(readCtx, sqlc.ListProposalRevisionsParams{
		ID: key.ProposalID, OrganizationID: key.OrganizationID,
	})
	if err != nil {
		return nil, fmt.Errorf("list revisions of %s: %w", key.ProposalID, err)
	}
	if len(rows) == 0 {
		return nil, apperror.OutsideScope()
	}
	itemRows, err := queries.ListProposalRevisionItems(readCtx, sqlc.ListProposalRevisionItemsParams{
		ID: key.ProposalID, OrganizationID: key.OrganizationID,
	})
	if err != nil {
		return nil, fmt.Errorf("list revision items of %s: %w", key.ProposalID, err)
	}
	return toRevisions(rows, itemsByRevision(itemRows)), nil
}

func itemsByRevision(rows []sqlc.ListProposalRevisionItemsRow) map[string][]sqlc.ListProposalItemsRow {
	grouped := map[string][]sqlc.ListProposalItemsRow{}
	for _, row := range rows {
		grouped[row.RevisionID] = append(grouped[row.RevisionID], sqlc.ListProposalItemsRow{
			ProductID: row.ProductID, ProductName: row.ProductName, Sku: row.Sku, Unit: row.Unit,
			Action: row.Action, PricingModel: row.PricingModel, RetailPrice: row.RetailPrice,
			CommissionBps: row.CommissionBps, DepositPrice: row.DepositPrice,
		})
	}
	return grouped
}

func toRevisions(rows []sqlc.ListProposalRevisionsRow, items map[string][]sqlc.ListProposalItemsRow) []domain.Revision {
	revisions := make([]domain.Revision, 0, len(rows))
	for _, row := range rows {
		revisions = append(revisions, domain.Revision{
			ID: row.ID, No: int(row.RevisionNo), AuthorOrgID: row.AuthorOrgID, AuthorUserID: row.AuthorUserID,
			Terms: terms.General{
				Schedule: terms.Schedule(row.SettlementSchedule), PaymentTermsDays: int(row.PaymentTermsDays),
				Rhythm: terms.Rhythm(row.CountRhythm), Notes: row.Notes,
			},
			Message: row.Message, Items: toItems(items[row.ID]), At: row.CreatedAt,
		})
	}
	return revisions
}
