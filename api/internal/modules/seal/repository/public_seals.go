package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Seals) GetPublicSeal(ctx context.Context, code string) (domain.PublicSeal, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetPublicSeal(readCtx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PublicSeal{}, apperror.NotFound()
	}
	if err != nil {
		return domain.PublicSeal{}, fmt.Errorf("get seal %s: %w", code, err)
	}
	return domain.PublicSeal{
		Seal: toSeal(sqlc.ListPartnershipSealsRow{
			ID: row.ID, PartnershipID: row.PartnershipID, Seq: row.Seq, EventType: row.EventType,
			EntityType: row.EntityType, EntityID: row.EntityID, EntityNumber: row.EntityNumber,
			ContentHash: row.ContentHash, ActorUserID: row.ActorUserID, ActorOrgID: row.ActorOrgID,
			OccurredAt: row.OccurredAt, PrevHash: row.PrevHash, EntryHash: row.EntryHash, Signature: row.Signature,
			KeyID: row.KeyID, Code: row.Code,
		}),
		Parties:           domain.Parties{ConsignorName: row.ConsignorName, StoreName: row.StoreName},
		PreviousEntryHash: row.PreviousEntryHash,
	}, nil
}

func (r *Seals) ListSealedPartnerships(ctx context.Context) ([]string, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	ids, err := sqlc.New(r.pool.Querier(ctx)).ListSealedPartnerships(readCtx)
	if err != nil {
		return nil, fmt.Errorf("list sealed partnerships: %w", err)
	}
	return ids, nil
}
