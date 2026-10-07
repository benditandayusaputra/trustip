package repository

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const nilID = "00000000-0000-0000-0000-000000000000"

type Inbox struct {
	pool *database.Pool
}

func NewInbox(pool *database.Pool) *Inbox {
	return &Inbox{pool: pool}
}

func (r *Inbox) Deliver(ctx context.Context, notifications []domain.Notification) error {
	if len(notifications) == 0 {
		return nil
	}
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	first := notifications[0]
	ids := make([]string, 0, len(notifications))
	userIDs := make([]string, 0, len(notifications))
	for _, notification := range notifications {
		ids = append(ids, notification.ID)
		userIDs = append(userIDs, notification.UserID)
	}
	err := sqlc.New(r.pool.Querier(ctx)).DeliverNotifications(writeCtx, sqlc.DeliverNotificationsParams{
		Ids: ids, UserIds: userIDs, Type: first.Type, Title: first.Title, Body: first.Body, Link: first.Link,
		CreatedAt: first.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("deliver %s notifications: %w", first.Type, err)
	}
	return nil
}

func (r *Inbox) ListNotifications(ctx context.Context, query domain.PageQuery) ([]domain.Notification, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	afterID := query.After.ID
	if query.After.IsStart() {
		afterID = nilID
	}
	rows, err := sqlc.New(r.pool.Querier(ctx)).ListNotifications(readCtx, sqlc.ListNotificationsParams{
		UserID: query.UserID, FromStart: query.After.IsStart(), AfterAt: query.After.At, AfterID: afterID,
		RowLimit: int64(query.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list notifications of %s: %w", query.UserID, err)
	}
	notifications := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		notification := domain.Notification{
			ID: row.ID, UserID: query.UserID, Type: row.Type, Title: row.Title, Body: row.Body, Link: row.Link,
			CreatedAt: row.CreatedAt,
		}
		if row.IsRead {
			notification.ReadAt = row.ReadAt
		}
		notifications = append(notifications, notification)
	}
	return notifications, nil
}

func (r *Inbox) CountUnread(ctx context.Context, userID string) (int, error) {
	readCtx, cancel := database.ReadContext(ctx)
	defer cancel()
	count, err := sqlc.New(r.pool.Querier(ctx)).CountUnreadNotifications(readCtx, userID)
	if err != nil {
		return 0, fmt.Errorf("count unread of %s: %w", userID, err)
	}
	return int(count), nil
}
