package main

import (
	"context"
	"fmt"

	notificationdomain "github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
	notificationrepository "github.com/benditandayusaputra/trustip/api/internal/modules/notification/repository"
	notificationservice "github.com/benditandayusaputra/trustip/api/internal/modules/notification/service"
	proposaldomain "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

type recipientDirectory interface {
	ListRecipients(ctx context.Context, organizationID string) ([]notificationdomain.Recipient, error)
}

type organizationAnnouncer struct {
	directory recipientDirectory
	notifier  *notificationservice.Notifier
}

func newOrganizationAnnouncer(pool *database.Pool, clock notificationservice.Clock) organizationAnnouncer {
	inbox := notificationrepository.NewInbox(pool)
	return organizationAnnouncer{directory: inbox, notifier: notificationservice.NewNotifier(
		notificationservice.NotifierDeps{Sink: inbox, Clock: clock, IDs: idgen.UUIDv7{}},
	)}
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
	recipients, err := a.directory.ListRecipients(ctx, notice.OrganizationID)
	if err != nil {
		return fmt.Errorf("find recipients: %w", err)
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
