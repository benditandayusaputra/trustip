package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/service"
)

type fakeInbox struct {
	delivered []domain.Notification
	stored    []domain.Notification
	asked     []domain.PageQuery
	reads     []domain.Read
}

func (f *fakeInbox) Deliver(_ context.Context, notifications []domain.Notification) error {
	f.delivered = append(f.delivered, notifications...)
	return nil
}

func (f *fakeInbox) ListNotifications(_ context.Context, query domain.PageQuery) ([]domain.Notification, error) {
	f.asked = append(f.asked, query)
	return f.stored[:min(query.Limit, len(f.stored))], nil
}

func (f *fakeInbox) CountUnread(context.Context, string) (int, error) { return 3, nil }

func (f *fakeInbox) MarkRead(_ context.Context, read domain.Read) error {
	if read.NotificationID == "elsewhere" {
		return apperror.NotFound()
	}
	f.reads = append(f.reads, read)
	return nil
}

func (f *fakeInbox) MarkAllRead(_ context.Context, read domain.Read) error {
	f.reads = append(f.reads, read)
	return nil
}

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC) }

type sequenceIDs struct{ next int }

func (s *sequenceIDs) NewID() (string, error) {
	s.next++
	return fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", s.next), nil
}

func member() actor.Actor {
	return actor.Actor{UserID: "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e", Membership: actor.Membership{
		OrganizationID: "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c", OrganizationType: actor.Store, Role: actor.Staff,
	}}
}

func TestNotify_BRNTF01_DeliversOneRowPerRecipientExceptTheActor(t *testing.T) {
	inbox := &fakeInbox{}
	notifier := service.NewNotifier(service.NotifierDeps{Sink: inbox, Clock: fixedClock{}, IDs: &sequenceIDs{}})

	err := notifier.Notify(t.Context(), service.Notice{
		Draft: domain.DraftInput{Type: "DROP_RECORDED", Title: "Titipan masuk", Body: "x", Link: "/app/titipan"},
		Audience: domain.Audience{
			ActorID:    "store-owner",
			Candidates: []domain.Recipient{{UserID: "store-owner"}, {UserID: "consignor-owner"}},
		},
	})

	if err != nil || len(inbox.delivered) != 1 || inbox.delivered[0].UserID != "consignor-owner" ||
		inbox.delivered[0].ID == "" {
		t.Fatalf("delivered = %+v, err = %v", inbox.delivered, err)
	}
}

func TestNotify_FRNTF05_UnsafeLinkIsRefused(t *testing.T) {
	inbox := &fakeInbox{}
	notifier := service.NewNotifier(service.NotifierDeps{Sink: inbox, Clock: fixedClock{}, IDs: &sequenceIDs{}})

	err := notifier.Notify(t.Context(), service.Notice{
		Draft: domain.DraftInput{Type: "DROP_RECORDED", Link: "https://evil.example"},
	})

	if err == nil || len(inbox.delivered) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestListNotifications_RRES04_AsksOneExtraToKnowIfMoreExist(t *testing.T) {
	inbox := &fakeInbox{}
	for index := range 3 {
		inbox.stored = append(inbox.stored, domain.Notification{
			ID: fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", index), CreatedAt: fixedClock{}.Now(),
		})
	}

	page, err := service.NewListNotifications(inbox).Execute(t.Context(), member(), service.PageRequest{Limit: "2"})

	if err != nil || len(page.Items) != 2 || !page.HasMore || page.NextCursor == "" || inbox.asked[0].Limit != 3 {
		t.Fatalf("page = %+v, err = %v", page, err)
	}
}

func TestListNotifications_FRNTF01_OnlyTheSignedInUsersOwn(t *testing.T) {
	inbox := &fakeInbox{}

	if _, err := service.NewListNotifications(inbox).Execute(t.Context(), member(), service.PageRequest{}); err != nil {
		t.Fatalf("err = %v", err)
	}

	if inbox.asked[0].UserID != member().UserID || inbox.asked[0].Limit != 21 {
		t.Fatalf("asked = %+v", inbox.asked)
	}
}

func TestMarkRead_BRAUTHZ02_NotificationOfAnotherUserIsNotFound(t *testing.T) {
	err := service.NewMarkRead(&fakeInbox{}, fixedClock{}).Execute(t.Context(), member(), "elsewhere")

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestMarkAllRead_FRNTF04_MarksForTheSignedInUser(t *testing.T) {
	inbox := &fakeInbox{}

	err := service.NewMarkAllRead(inbox, fixedClock{}).Execute(t.Context(), member())

	if err != nil || len(inbox.reads) != 1 || inbox.reads[0].UserID != member().UserID {
		t.Fatalf("reads = %+v, err = %v", inbox.reads, err)
	}
}

func TestCountUnread_FRNTF01_NewcomerIsForbidden(t *testing.T) {
	_, err := service.NewCountUnread(&fakeInbox{}).Execute(t.Context(), actor.Actor{UserID: "u"})

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}
