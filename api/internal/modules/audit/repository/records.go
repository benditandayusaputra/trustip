package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

type Records struct {
	pool *database.Pool
}

func NewRecords(pool *database.Pool) *Records {
	return &Records{pool: pool}
}

func (r *Records) Record(ctx context.Context, entry audit.Entry) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	metadata := entry.Metadata
	if metadata == nil {
		metadata = map[string]string{}
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode audit metadata: %w", err)
	}
	err = sqlc.New(r.pool.Querier(ctx)).RecordAudit(writeCtx, sqlc.RecordAuditParams{
		ID: entry.ID, OrganizationID: entry.OrganizationID, ActorID: entry.ActorID, Action: string(entry.Action),
		EntityType: string(entry.EntityType), EntityID: entry.EntityID, Metadata: encoded, Ip: entry.IP,
		CreatedAt: entry.At,
	})
	if err != nil {
		return fmt.Errorf("record audit %s: %w", entry.Action, err)
	}
	return nil
}
