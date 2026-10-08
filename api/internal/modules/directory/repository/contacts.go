package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/directory/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Directory) GetContact(ctx context.Context, organizationID string) (domain.Contact, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	row, err := sqlc.New(r.pool.Querier(ctx)).GetDirectoryContact(readCtx, organizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Contact{}, apperror.OutsideScope()
	}
	if err != nil {
		return domain.Contact{}, fmt.Errorf("get directory contact %s: %w", organizationID, err)
	}
	return domain.Contact{
		OrganizationID: row.ID, Name: row.Name, Type: actor.OrganizationType(row.Type), CityName: row.CityName,
		WhatsApp: row.Whatsapp,
	}, nil
}
