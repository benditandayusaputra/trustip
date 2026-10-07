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

type Seals struct {
	pool *database.Pool
}

func NewSeals(pool *database.Pool) *Seals {
	return &Seals{pool: pool}
}

func (r *Seals) LockHead(ctx context.Context, partnershipID string) (domain.Head, error) {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).LockSealHead(writeCtx, partnershipID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Head{}, apperror.NotFound()
	}
	if err != nil {
		return domain.Head{}, fmt.Errorf("lock seal head of %s: %w", partnershipID, err)
	}
	return domain.Head{Seq: int(row.SealSeq), Hash: row.SealHeadHash}, nil
}

func (r *Seals) Append(ctx context.Context, seal domain.Seal) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	entry := seal.Entry
	err := queries.AppendSeal(writeCtx, sqlc.AppendSealParams{
		ID: seal.ID, PartnershipID: entry.PartnershipID, Seq: int64(entry.Seq), FormatVersion: domain.FormatVersion,
		EventType: string(entry.Event), EntityType: entry.Document.Type, EntityID: entry.Document.ID,
		EntityNumber: entry.Document.Number, ContentHash: entry.ContentHash, ActorUserID: entry.ActorUserID,
		ActorOrgID: entry.ActorOrgID, OccurredAt: entry.OccurredAt, PrevHash: orEmpty(entry.PrevHash),
		EntryHash: seal.EntryHash, Signature: seal.Signature.Bytes, KeyID: seal.Signature.KeyID, Code: seal.Code,
	})
	if err != nil {
		return fmt.Errorf("append seal %s: %w", seal.ID, err)
	}
	err = queries.AdvanceSealHead(writeCtx, sqlc.AdvanceSealHeadParams{
		ID: entry.PartnershipID, SealSeq: int64(entry.Seq), SealHeadHash: seal.EntryHash,
	})
	if err != nil {
		return fmt.Errorf("advance seal head of %s: %w", entry.PartnershipID, err)
	}
	return nil
}

func orEmpty(hash []byte) []byte {
	if hash == nil {
		return []byte{}
	}
	return hash
}

func (r *Seals) ListChain(ctx context.Context, partnershipID string) ([]domain.Seal, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListPartnershipSeals(readCtx, partnershipID)
	if err != nil {
		return nil, fmt.Errorf("list seals of %s: %w", partnershipID, err)
	}
	seals := make([]domain.Seal, 0, len(rows))
	for _, row := range rows {
		seals = append(seals, toSeal(row))
	}
	return seals, nil
}

func toSeal(row sqlc.ListPartnershipSealsRow) domain.Seal {
	return domain.Seal{
		ID: row.ID,
		Entry: domain.Entry{
			PartnershipID: row.PartnershipID, Seq: int(row.Seq), Event: domain.Event(row.EventType),
			Document:    domain.Document{Type: row.EntityType, ID: row.EntityID, Number: row.EntityNumber},
			ContentHash: row.ContentHash, ActorUserID: row.ActorUserID, ActorOrgID: row.ActorOrgID,
			OccurredAt: row.OccurredAt, PrevHash: row.PrevHash,
		},
		EntryHash: row.EntryHash,
		Signature: domain.Signature{KeyID: row.KeyID, Bytes: row.Signature},
		Code:      row.Code,
	}
}

func (r *Seals) ListEntitySeals(ctx context.Context, document domain.Document) ([]domain.TrailEntry, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListEntitySeals(readCtx, sqlc.ListEntitySealsParams{
		EntityType: document.Type, EntityID: document.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("list seals of %s %s: %w", document.Type, document.ID, err)
	}
	trail := make([]domain.TrailEntry, 0, len(rows))
	for _, row := range rows {
		seal := toSeal(sqlc.ListPartnershipSealsRow{
			ID: row.ID, PartnershipID: row.PartnershipID, Seq: row.Seq, EventType: row.EventType,
			EntityType: row.EntityType, EntityID: row.EntityID, EntityNumber: row.EntityNumber,
			ContentHash: row.ContentHash, ActorUserID: row.ActorUserID, ActorOrgID: row.ActorOrgID,
			OccurredAt: row.OccurredAt, PrevHash: row.PrevHash, EntryHash: row.EntryHash, Signature: row.Signature,
			KeyID: row.KeyID, Code: row.Code,
		})
		trail = append(trail, domain.TrailEntry{
			Seal: seal, ActorOrgName: row.ActorOrgName, ActorUserName: row.ActorUserName,
		})
	}
	return trail, nil
}
