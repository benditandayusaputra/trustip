package service

import (
	"context"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
)

type NotificationSink interface {
	Deliver(ctx context.Context, notifications []domain.Notification) error
}

type NotificationReader interface {
	ListNotifications(ctx context.Context, query domain.PageQuery) ([]domain.Notification, error)
	CountUnread(ctx context.Context, userID string) (int, error)
}

type NotificationMarker interface {
	MarkRead(ctx context.Context, read domain.Read) error
	MarkAllRead(ctx context.Context, read domain.Read) error
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() (string, error)
}
