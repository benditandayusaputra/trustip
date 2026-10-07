package main

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
)

type memoryInbox struct {
	mu    sync.Mutex
	items []domain.Notification
}

func (m *memoryInbox) ListNotifications(_ context.Context, query domain.PageQuery) ([]domain.Notification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	mine := []domain.Notification{}
	for _, item := range m.items {
		if item.UserID == query.UserID {
			mine = append(mine, item)
		}
	}
	return mine[:min(query.Limit, len(mine))], nil
}

func (m *memoryInbox) CountUnread(ctx context.Context, userID string) (int, error) {
	items, err := m.ListNotifications(ctx, domain.PageQuery{UserID: userID, Limit: len(m.items)})
	unread := 0
	for _, item := range items {
		if !item.IsRead() {
			unread++
		}
	}
	return unread, err
}

func (m *memoryInbox) MarkRead(_ context.Context, read domain.Read) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for index, item := range m.items {
		if item.ID == read.NotificationID && item.UserID == read.UserID {
			m.items[index].ReadAt = read.At
			return nil
		}
	}
	return apperror.NotFound()
}

func (m *memoryInbox) MarkAllRead(_ context.Context, read domain.Read) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for index, item := range m.items {
		if item.UserID == read.UserID {
			m.items[index].ReadAt = read.At
		}
	}
	return nil
}

const seededNotificationID = "0192f0c4-0000-7000-8000-000000000101"

func testNotificationComponents(now *clock.Manual) notificationComponents {
	inbox := &memoryInbox{items: []domain.Notification{{
		ID: seededNotificationID, UserID: firstSequenceUserID, Type: "DROP_RECORDED", Title: "Titipan masuk",
		Body: "Warung Bu Tini mencatat 20 unit titipan masuk.", Link: "/app/titipan", CreatedAt: testNow(),
	}}}
	return notificationComponents{inbox: inbox, clock: now}
}

func TestNotifications_FRNTF01_MemberSeesTheirListAndUnreadCount(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	listed := visitor.get("/api/v1/notifications")
	counted := visitor.get("/api/v1/notifications/unread-count")

	requireCode(t, listed, "OK")
	if !strings.Contains(string(listed.body), `"link":"/app/titipan","is_read":false`) ||
		!strings.Contains(string(counted.body), `"count":1`) {
		t.Fatalf("listed = %s, counted = %s", listed.body, counted.body)
	}
	loadContract(t).requireConforms(t, listed)
	loadContract(t).requireConforms(t, counted)
}

func TestNotifications_FRNTF04_MarkingOneAndAllAsRead(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	one := visitor.post("/api/v1/notifications/" + seededNotificationID + "/read")
	all := visitor.post("/api/v1/notifications/read-all")
	counted := visitor.get("/api/v1/notifications/unread-count")

	requireCode(t, one, "OK")
	requireCode(t, all, "OK")
	if !strings.Contains(string(counted.body), `"count":0`) {
		t.Fatalf("counted = %s", counted.body)
	}
	loadContract(t).requireConforms(t, one)
}

func TestNotifications_RRES04_MalformedCursorIsAFieldError(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.get("/api/v1/notifications?cursor=!!!")

	requireCode(t, got, "VALIDATION_FAILED")
	loadContract(t).requireConforms(t, got)
}
