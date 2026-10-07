package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

func (r *Inbox) MarkRead(ctx context.Context, read domain.Read) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	updated, err := sqlc.New(r.pool.Querier(ctx)).MarkNotificationRead(writeCtx, sqlc.MarkNotificationReadParams{
		ID: read.NotificationID, UserID: read.UserID, ReadAt: read.At,
	})
	if err != nil {
		return fmt.Errorf("mark notification %s read: %w", read.NotificationID, err)
	}
	if updated == 0 {
		return apperror.NotFound()
	}
	return nil
}

func (r *Inbox) MarkAllRead(ctx context.Context, read domain.Read) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	err := sqlc.New(r.pool.Querier(ctx)).MarkAllNotificationsRead(writeCtx, sqlc.MarkAllNotificationsReadParams{
		UserID: read.UserID, ReadAt: read.At,
	})
	if err != nil {
		return fmt.Errorf("mark all notifications of %s read: %w", read.UserID, err)
	}
	return nil
}
