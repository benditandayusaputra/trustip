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
	dailyLimitConstraint = "proposal_quotas_daily_limit"
	needOpenStatus       = "OPEN"
)

func (r *Proposals) ConsumeQuota(ctx context.Context, key domain.QuotaKey) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	err := sqlc.New(r.pool.Querier(ctx)).ConsumeProposalQuota(writeCtx, sqlc.ConsumeProposalQuotaParams{
		OrganizationID: key.OrganizationID, Day: key.Day,
	})
	if database.IsCheckViolation(err, dailyLimitConstraint) {
		return domain.ErrQuotaExhausted
	}
	if err != nil {
		return fmt.Errorf("consume proposal quota of %s: %w", key.OrganizationID, err)
	}
	return nil
}

func (r *Proposals) GetDirectParties(ctx context.Context, pair domain.DirectPair) (domain.DirectParties, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetDirectParties(readCtx, sqlc.GetDirectPartiesParams{
		SenderID: pair.SenderID, TargetID: pair.TargetID, NeedID: pair.NeedID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DirectParties{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.DirectParties{}, fmt.Errorf("get parties %s and %s: %w", pair.SenderID, pair.TargetID, err)
	}
	return domain.DirectParties{
		SenderName: row.SenderName,
		Target: domain.Target{
			OrganizationID: row.TargetID, Name: row.TargetName, Type: actor.OrganizationType(row.TargetType),
		},
		Need: domain.NeedRef{
			ID: pair.NeedID, StoreID: row.NeedStoreOrgID, IsOpen: row.NeedStatus == needOpenStatus,
			ExpiresAt: row.NeedExpiresAt,
		},
		LastDeclinedAt: row.LastDeclinedAt, ArePartners: row.ArePartners,
	}, nil
}
