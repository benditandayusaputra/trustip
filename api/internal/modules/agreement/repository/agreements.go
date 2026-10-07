package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const (
	activePairConstraint = "partnerships_active_pair_key"
	partnershipActive    = "ACTIVE"
)

type Agreements struct {
	pool *database.Pool
}

func NewAgreements(pool *database.Pool) *Agreements {
	return &Agreements{pool: pool}
}

func (r *Agreements) CreatePartnership(ctx context.Context, partnership domain.Partnership) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	err := sqlc.New(r.pool.Querier(ctx)).CreatePartnership(writeCtx, sqlc.CreatePartnershipParams{
		ID: partnership.ID, ConsignorOrgID: partnership.ConsignorID, StoreOrgID: partnership.StoreID,
		StartedAt: partnership.StartedAt,
	})
	if database.IsUniqueViolation(err, activePairConstraint) {
		return domain.ErrAlreadyPartners
	}
	if err != nil {
		return fmt.Errorf("create partnership %s: %w", partnership.ID, err)
	}
	return nil
}

func (r *Agreements) CreateAgreement(ctx context.Context, agreement domain.Agreement) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	general := agreement.Terms
	err := queries.CreateAgreement(writeCtx, sqlc.CreateAgreementParams{
		ID: agreement.ID, PartnershipID: agreement.PartnershipID, Version: int64(agreement.Version),
		SettlementSchedule: string(general.Schedule), PaymentTermsDays: int64(general.PaymentTermsDays),
		CountRhythm: string(general.Rhythm), Notes: general.Notes, SourceProposalID: agreement.SourceProposalID,
		ProposedBy: agreement.ProposedBy, AcceptedBy: agreement.AcceptedBy, AcceptedAt: agreement.AcceptedAt,
	})
	if err != nil {
		return fmt.Errorf("create agreement %s: %w", agreement.ID, err)
	}
	if len(agreement.Items) == 0 {
		return nil
	}
	if err := queries.AddAgreementItems(writeCtx, itemParams(agreement)); err != nil {
		return fmt.Errorf("add items to agreement %s: %w", agreement.ID, err)
	}
	return nil
}

func itemParams(agreement domain.Agreement) sqlc.AddAgreementItemsParams {
	params := sqlc.AddAgreementItemsParams{AgreementID: agreement.ID}
	for _, item := range agreement.Items {
		params.ProductIds = append(params.ProductIds, item.ProductID)
		params.PricingModels = append(params.PricingModels, string(item.Terms.Model))
		params.RetailPrices = append(params.RetailPrices, item.Terms.RetailPrice.Int64())
		params.CommissionBps = append(params.CommissionBps, int64(item.Terms.CommissionBPS))
		params.DepositPrices = append(params.DepositPrices, item.Terms.DepositPrice.Int64())
		params.ActiveFlags = append(params.ActiveFlags, item.IsActive)
	}
	return params
}
