package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	nilID              = "00000000-0000-0000-0000-000000000000"
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

func (r *Proposals) Open(ctx context.Context, opening domain.Opening) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	err := queries.CreateProposal(writeCtx, sqlc.CreateProposalParams{
		ID: opening.IDs.ProposalID, Number: opening.Number, Kind: string(opening.Kind),
		ConsignorOrgID: opening.ConsignorID, StoreOrgID: opening.StoreID, FromOrgID: opening.Sender.OrganizationID,
		AwaitingOrgID: opening.AwaitingID, PartnershipID: opening.PartnershipID, InviteCode: opening.IDs.Code,
		NeedID:    opening.NeedID,
		ExpiresAt: opening.ExpiresAt, CreatedAt: opening.At,
	})
	if database.IsUniqueViolation(err, openPairConstraint) {
		return domain.ErrAlreadyOpen
	}
	if err != nil {
		return fmt.Errorf("create proposal %s: %w", opening.IDs.ProposalID, err)
	}
	return saveRevision(writeCtx, queries, revisionRow{
		id: opening.IDs.RevisionID, proposalID: opening.IDs.ProposalID, no: firstRevision,
		authorOrgID: opening.Sender.OrganizationID, authorUserID: opening.Sender.UserID,
		consignorID: opening.ConsignorID, draft: opening.Draft, at: opening.At,
	})
}

type revisionRow struct {
	id           string
	proposalID   string
	no           int
	authorOrgID  string
	authorUserID string
	consignorID  string
	draft        domain.RevisionDraft
	at           time.Time
}

func saveRevision(ctx context.Context, queries *sqlc.Queries, revision revisionRow) error {
	terms := revision.draft.Terms
	err := queries.CreateProposalRevision(ctx, sqlc.CreateProposalRevisionParams{
		ID: revision.id, ProposalID: revision.proposalID, RevisionNo: int64(revision.no),
		AuthorOrgID: revision.authorOrgID, AuthorUserID: revision.authorUserID,
		SettlementSchedule: string(terms.Schedule), PaymentTermsDays: int64(terms.PaymentTermsDays),
		CountRhythm: string(terms.Rhythm), Notes: terms.Notes, Message: revision.draft.Message,
		CreatedAt: revision.at,
	})
	if err != nil {
		return fmt.Errorf("create revision of %s: %w", revision.proposalID, err)
	}
	return addItems(ctx, queries, revision)
}

func addItems(ctx context.Context, queries *sqlc.Queries, revision revisionRow) error {
	items := revision.draft.Items
	if len(items) == 0 {
		return nil
	}
	params := sqlc.AddProposalItemsParams{
		RevisionID: revision.id, ConsignorOrgID: revision.consignorID,
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
		return fmt.Errorf("add items to %s: %w", revision.id, err)
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

func (r *Proposals) SaveDecision(ctx context.Context, decision domain.Decision) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	saved, err := sqlc.New(r.pool.Querier(ctx)).DecideProposal(writeCtx, sqlc.DecideProposalParams{
		ID: decision.ProposalID, Status: string(decision.Status), DecidedBy: decision.DecidedBy,
		DecidedAt: decision.At, DecisionReason: decision.Reason, PartnershipID: decision.PartnershipID,
		RevisionNo: int64(decision.RevisionNo),
	})
	if err != nil {
		return fmt.Errorf("save decision on %s: %w", decision.ProposalID, err)
	}
	if saved == 0 {
		return domain.ErrAlreadyDecided
	}
	return nil
}

func (r *Proposals) ExpireOpen(ctx context.Context, now time.Time) (int, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	expired, err := sqlc.New(r.pool.Querier(ctx)).ExpireOpenProposals(writeCtx, now)
	if err != nil {
		return 0, fmt.Errorf("expire proposals open at %s: %w", now, err)
	}
	return int(expired), nil
}
