package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
)

type needEvent struct {
	signedIn actor.Actor
	needID   string
	action   audit.Action
	at       time.Time
}

func finishNeedEvent(ctx context.Context, deps Deps, event needEvent) (domain.Need, error) {
	err := deps.Audit.Record(ctx, audit.Entry{
		OrganizationID: event.signedIn.Membership.OrganizationID, ActorID: event.signedIn.UserID,
		Action: event.action, EntityType: audit.NeedEntity, EntityID: event.needID, At: event.at,
	})
	if err != nil {
		return domain.Need{}, fmt.Errorf("audit %s: %w", event.action, err)
	}
	need, err := deps.Reader.GetNeed(ctx, keyOf(event.signedIn, event.needID))
	if err != nil {
		return domain.Need{}, fmt.Errorf("load need %s: %w", event.needID, err)
	}
	return need.AsOf(event.at), nil
}

func keyOf(signedIn actor.Actor, needID string) domain.NeedKey {
	return domain.NeedKey{StoreID: signedIn.Membership.OrganizationID, NeedID: needID}
}
