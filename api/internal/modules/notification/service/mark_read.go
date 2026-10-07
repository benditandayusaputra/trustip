package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
)

type MarkRead struct {
	marker NotificationMarker
	clock  Clock
}

func NewMarkRead(marker NotificationMarker, clock Clock) *MarkRead {
	return &MarkRead{marker: marker, clock: clock}
}

func (uc *MarkRead) Execute(ctx context.Context, signedIn actor.Actor, notificationID string) error {
	if err := policy.RequireMember(signedIn); err != nil {
		return fmt.Errorf("check notification reader: %w", err)
	}
	read := domain.Read{UserID: signedIn.UserID, NotificationID: notificationID, At: uc.clock.Now()}
	if err := uc.marker.MarkRead(ctx, read); err != nil {
		return fmt.Errorf("mark %s read: %w", notificationID, err)
	}
	return nil
}

type MarkAllRead struct {
	marker NotificationMarker
	clock  Clock
}

func NewMarkAllRead(marker NotificationMarker, clock Clock) *MarkAllRead {
	return &MarkAllRead{marker: marker, clock: clock}
}

func (uc *MarkAllRead) Execute(ctx context.Context, signedIn actor.Actor) error {
	if err := policy.RequireMember(signedIn); err != nil {
		return fmt.Errorf("check notification reader: %w", err)
	}
	if err := uc.marker.MarkAllRead(ctx, domain.Read{UserID: signedIn.UserID, At: uc.clock.Now()}); err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}
