package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
)

type PageRequest struct {
	Cursor string
	Limit  string
}

type Page struct {
	Items      []domain.Notification
	NextCursor string
	HasMore    bool
}

type ListNotifications struct {
	reader NotificationReader
}

func NewListNotifications(reader NotificationReader) *ListNotifications {
	return &ListNotifications{reader: reader}
}

func (uc *ListNotifications) Execute(ctx context.Context, signedIn actor.Actor, request PageRequest) (Page, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return Page{}, fmt.Errorf("check notification reader: %w", err)
	}
	after, err := pagination.ParseCursor(request.Cursor)
	if err != nil {
		return Page{}, fmt.Errorf("read cursor: %w", err)
	}
	limit := pagination.ParseLimit(request.Limit)
	found, err := uc.reader.ListNotifications(ctx, domain.PageQuery{
		UserID: signedIn.UserID, After: after, Limit: limit + 1,
	})
	if err != nil {
		return Page{}, fmt.Errorf("list notifications: %w", err)
	}
	if len(found) <= limit {
		return Page{Items: found}, nil
	}
	last := found[limit-1]
	return Page{
		Items:      found[:limit],
		NextCursor: pagination.CursorOf(pagination.Position{At: last.CreatedAt, ID: last.ID}),
		HasMore:    true,
	}, nil
}

type CountUnread struct {
	reader NotificationReader
}

func NewCountUnread(reader NotificationReader) *CountUnread {
	return &CountUnread{reader: reader}
}

func (uc *CountUnread) Execute(ctx context.Context, signedIn actor.Actor) (int, error) {
	if err := policy.RequireMember(signedIn); err != nil {
		return 0, fmt.Errorf("check notification reader: %w", err)
	}
	count, err := uc.reader.CountUnread(ctx, signedIn.UserID)
	if err != nil {
		return 0, fmt.Errorf("count unread: %w", err)
	}
	return count, nil
}
