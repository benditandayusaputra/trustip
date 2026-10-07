package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
)

type organizationEvent struct {
	signedIn       actor.Actor
	organizationID string
	action         audit.Action
	metadata       map[string]string
	at             time.Time
}

func recordOrganizationEvent(ctx context.Context, trail AuditTrail, event organizationEvent) error {
	err := trail.Record(ctx, audit.Entry{
		OrganizationID: event.organizationID, ActorID: event.signedIn.UserID, Action: event.action,
		EntityType: audit.OrganizationEntity, EntityID: event.organizationID, Metadata: event.metadata,
		At: event.at,
	})
	if err != nil {
		return fmt.Errorf("audit %s: %w", event.action, err)
	}
	return nil
}
