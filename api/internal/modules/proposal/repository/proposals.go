package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const (
	firstRevision      = 1
	openPairConstraint = "proposals_open_pair_key"
)

type Proposals struct {
	pool *database.Pool
}

func NewProposals(pool *database.Pool) *Proposals {
	return &Proposals{pool: pool}
}

func (r *Proposals) NextNumber(ctx context.Context, key domain.SequenceKey) (int, error) {
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

func (r *Proposals) CreateInvitation(ctx context.Context, invitation domain.Invitation) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	err := queries.CreateProposal(writeCtx, sqlc.CreateProposalParams{
		ID: invitation.IDs.ProposalID, Number: invitation.Number, Kind: string(domain.New),
		ConsignorOrgID: invitation.ConsignorID(), StoreOrgID: invitation.StoreID(),
		FromOrgID: invitation.Sender.OrganizationID, InviteCode: invitation.IDs.Code,
		ExpiresAt: invitation.ExpiresAt, CreatedAt: invitation.At,
	})
	if err != nil {
		return fmt.Errorf("create proposal %s: %w", invitation.IDs.ProposalID, err)
	}
	if err := createRevision(writeCtx, queries, invitation); err != nil {
		return err
	}
	return addItems(writeCtx, queries, invitation)
}

func createRevision(ctx context.Context, queries *sqlc.Queries, invitation domain.Invitation) error {
	terms := invitation.Draft.Terms
	err := queries.CreateProposalRevision(ctx, sqlc.CreateProposalRevisionParams{
		ID: invitation.IDs.RevisionID, ProposalID: invitation.IDs.ProposalID, RevisionNo: firstRevision,
		AuthorOrgID: invitation.Sender.OrganizationID, AuthorUserID: invitation.Sender.UserID,
		SettlementSchedule: string(terms.Schedule), PaymentTermsDays: int64(terms.PaymentTermsDays),
		CountRhythm: string(terms.Rhythm), Notes: terms.Notes, Message: invitation.Draft.Message,
		CreatedAt: invitation.At,
	})
	if err != nil {
		return fmt.Errorf("create revision of %s: %w", invitation.IDs.ProposalID, err)
	}
	return nil
}

func addItems(ctx context.Context, queries *sqlc.Queries, invitation domain.Invitation) error {
	items := invitation.Draft.Items
	if len(items) == 0 {
		return nil
	}
	params := sqlc.AddProposalItemsParams{
		RevisionID: invitation.IDs.RevisionID, ConsignorOrgID: invitation.ConsignorID(),
	}
	for _, item := range items {
		params.ProductIds = append(params.ProductIds, item.ProductID)
		params.Actions = append(params.Actions, string(item.Action))
		params.PricingModels = append(params.PricingModels, string(item.Terms.Model))
		params.RetailPrices = append(params.RetailPrices, item.Terms.RetailPrice.Int64())
		params.CommissionBps = append(params.CommissionBps, int64(item.Terms.CommissionBPS))
		params.DepositPrices = append(params.DepositPrices, item.Terms.DepositPrice.Int64())
	}
	added, err := queries.AddProposalItems(ctx, params)
	if err != nil {
		return fmt.Errorf("add items to %s: %w", invitation.IDs.RevisionID, err)
	}
	if added != int64(len(items)) {
		return domain.ErrUnknownProduct
	}
	return nil
}

func (r *Proposals) GetInvitation(ctx context.Context, code string) (domain.InvitationView, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetInvitation(readCtx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.InvitationView{}, apperror.NotFound()
	}
	if err != nil {
		return domain.InvitationView{}, fmt.Errorf("get invitation: %w", err)
	}
	return domain.InvitationView{
		ProposalID: row.ID, Status: domain.Status(row.Status), ExpiresAt: row.ExpiresAt,
		Sender: domain.SenderView{
			OrganizationID: row.FromOrgID, Name: row.SenderName, Type: actor.OrganizationType(row.SenderType),
		},
		ConsignorID: row.ConsignorOrgID, StoreID: row.StoreOrgID, ItemCount: int(row.ItemCount),
	}, nil
}

func (r *Proposals) ClaimInvitation(ctx context.Context, claim domain.Claim) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	claimed, err := sqlc.New(r.pool.Querier(ctx)).ClaimInvitation(writeCtx, sqlc.ClaimInvitationParams{
		ID: claim.ProposalID, ConsignorOrgID: claim.ConsignorID, StoreOrgID: claim.StoreID,
		ClaimerOrgID: claim.ClaimerID, ClaimedAt: claim.At,
	})
	if database.IsUniqueViolation(err, openPairConstraint) {
		return domain.ErrAlreadyOpen
	}
	if err != nil {
		return fmt.Errorf("claim invitation %s: %w", claim.ProposalID, err)
	}
	if claimed == 0 {
		return domain.ErrClaimRaced
	}
	return nil
}

func (r *Proposals) HasActivePartnership(ctx context.Context, pair domain.Pair) (bool, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	found, err := sqlc.New(r.pool.Querier(ctx)).HasActivePartnership(readCtx, sqlc.HasActivePartnershipParams{
		ConsignorOrgID: pair.ConsignorID, StoreOrgID: pair.StoreID,
	})
	if err != nil {
		return false, fmt.Errorf("check partnership of %s and %s: %w", pair.ConsignorID, pair.StoreID, err)
	}
	return found, nil
}
