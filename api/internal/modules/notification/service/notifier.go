package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
)

type Notice struct {
	Draft    domain.DraftInput
	Audience domain.Audience
}

type NotifierDeps struct {
	Sink  NotificationSink
	Clock Clock
	IDs   IDGenerator
}

type Notifier struct {
	deps NotifierDeps
}

func NewNotifier(deps NotifierDeps) *Notifier {
	return &Notifier{deps: deps}
}

func (n *Notifier) Notify(ctx context.Context, notice Notice) error {
	draft, err := domain.NewDraft(notice.Draft)
	if err != nil {
		return fmt.Errorf("prepare %s notification: %w", notice.Draft.Type, err)
	}
	deliveries := domain.Deliveries(draft, domain.RecipientsOf(draft, notice.Audience), n.deps.Clock.Now())
	for index := range deliveries {
		id, err := n.deps.IDs.NewID()
		if err != nil {
			return fmt.Errorf("generate notification id: %w", err)
		}
		deliveries[index].ID = id
	}
	if err := n.deps.Sink.Deliver(ctx, deliveries); err != nil {
		return fmt.Errorf("deliver %s notifications: %w", draft.Type, err)
	}
	return nil
}
