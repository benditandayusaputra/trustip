package main

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	notificationdomain "github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
	notificationrepository "github.com/benditandayusaputra/trustip/api/internal/modules/notification/repository"
	notificationservice "github.com/benditandayusaputra/trustip/api/internal/modules/notification/service"
	proposaldomain "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
	"github.com/benditandayusaputra/trustip/api/internal/platform/sse"
)

type recipientDirectory interface {
	ListRecipients(ctx context.Context, organizationID string) ([]notificationdomain.Recipient, error)
}

type organizationAnnouncer struct {
	directory recipientDirectory
	notifier  *notificationservice.Notifier
}

type foundation struct {
	pool  *database.Pool
	clock serverClock
	feed  liveFeed
}

type liveFeed interface {
	Publish(userID string, event sse.Event)
}

func newOrganizationAnnouncer(base foundation) organizationAnnouncer {
	inbox := notificationrepository.NewInbox(base.pool)
	sink := liveInbox{inbox: inbox, commits: base.pool, feed: base.feed}
	return organizationAnnouncer{directory: inbox, notifier: notificationservice.NewNotifier(
		notificationservice.NotifierDeps{Sink: sink, Clock: base.clock, IDs: idgen.UUIDv7{}},
	)}
}

const liveNotificationEvent = "notification"

type liveNotification struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type liveInbox struct {
	inbox   *notificationrepository.Inbox
	commits *database.Pool
	feed    liveFeed
}

func (i liveInbox) Deliver(ctx context.Context, notifications []notificationdomain.Notification) error {
	if err := i.inbox.Deliver(ctx, notifications); err != nil {
		return fmt.Errorf("store notifications: %w", err)
	}
	events := make([]sse.Event, 0, len(notifications))
	for _, notification := range notifications {
		summary, err := json.Marshal(liveNotification{ID: notification.ID, Type: notification.Type})
		if err != nil {
			return fmt.Errorf("encode live notification: %w", err)
		}
		events = append(events, sse.Event{Name: liveNotificationEvent, Data: string(summary)})
	}
	i.commits.AfterCommit(ctx, func() {
		for index, notification := range notifications {
			i.feed.Publish(notification.UserID, events[index])
		}
	})
	return nil
}

type organizationNotice struct {
	OrganizationID string
	ActorUserID    string
	Type           string
	Title          string
	Body           string
	Link           string
	OwnersOnly     bool
}

type proposalAnnouncer struct {
	organizationAnnouncer
}

func (a proposalAnnouncer) Announce(ctx context.Context, notice proposaldomain.Notice) error {
	return a.announce(ctx, organizationNotice(notice))
}

func (a organizationAnnouncer) announce(ctx context.Context, notice organizationNotice) error {
	return a.announceTo(ctx, notice, "")
}

func (a organizationAnnouncer) announceTo(ctx context.Context, notice organizationNotice, userID string) error {
	recipients, err := a.directory.ListRecipients(ctx, notice.OrganizationID)
	if err != nil {
		return fmt.Errorf("find recipients: %w", err)
	}
	if userID != "" {
		recipients = slices.DeleteFunc(recipients, func(recipient notificationdomain.Recipient) bool {
			return recipient.UserID != userID
		})
	}
	err = a.notifier.Notify(ctx, notificationservice.Notice{
		Draft: notificationdomain.DraftInput{
			Type: notice.Type, Title: notice.Title, Body: notice.Body, Link: notice.Link,
			CarriesMoney: notice.OwnersOnly,
		},
		Audience: notificationdomain.Audience{Candidates: recipients, ActorID: notice.ActorUserID},
	})
	if err != nil {
		return fmt.Errorf("notify %s: %w", notice.Type, err)
	}
	return nil
}
