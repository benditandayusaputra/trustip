package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Inbox) ListRecipients(ctx context.Context, organizationID string) ([]domain.Recipient, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListOrganizationRecipients(readCtx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list recipients of %s: %w", organizationID, err)
	}
	recipients := make([]domain.Recipient, 0, len(rows))
	for _, row := range rows {
		recipients = append(recipients, domain.Recipient{UserID: row.UserID, Role: actor.Role(row.Role)})
	}
	return recipients, nil
}
