package main

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	auditrepository "github.com/benditandayusaputra/trustip/api/internal/modules/audit/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

type auditRecorder interface {
	Record(ctx context.Context, entry audit.Entry) error
}

type auditTrail struct {
	records auditRecorder
	ids     interface{ NewID() (string, error) }
}

func newAuditTrail(records *auditrepository.Records) auditTrail {
	return auditTrail{records: records, ids: idgen.UUIDv7{}}
}

func (a auditTrail) Record(ctx context.Context, entry audit.Entry) error {
	id, err := a.ids.NewID()
	if err != nil {
		return fmt.Errorf("generate audit id: %w", err)
	}
	entry.ID = id
	entry.IP = httpserver.ClientIPFrom(ctx)
	if err := a.records.Record(ctx, entry); err != nil {
		return fmt.Errorf("record audit entry: %w", err)
	}
	return nil
}
