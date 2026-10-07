package repository_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/notification/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	readerID   = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"
	strangerID = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0f"
	storedHash = "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"
)

func base() time.Time {
	return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
}

func inboxWithUsers(t *testing.T) (*database.Pool, *repository.Inbox) {
	t.Helper()
	pool := dbtest.MigratedPool(t)
	queries := sqlc.New(pool.Querier(t.Context()))
	for id, email := range map[string]string{readerID: "sari@contoh.id", strangerID: "tini@contoh.id"} {
		if err := queries.CreateUser(t.Context(), sqlc.CreateUserParams{
			ID: id, Email: email, PasswordHash: storedHash, PepperVersion: "v1", FullName: "Pengguna",
			Whatsapp: "6281234567890", PrivacyAcceptedAt: base(),
		}); err != nil {
			t.Fatalf("user: %v", err)
		}
	}
	return pool, repository.NewInbox(pool)
}

func notification(index int, userID string) domain.Notification {
	return domain.Notification{
		ID: fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", index), UserID: userID, Type: "DROP_RECORDED",
		Title: "Titipan masuk", Body: "Warung mencatat titipan.", Link: "/app/titipan",
		CreatedAt: base().Add(time.Duration(index) * time.Minute),
	}
}

func deliver(t *testing.T, inbox *repository.Inbox, notifications ...domain.Notification) {
	t.Helper()
	for _, item := range notifications {
		if err := inbox.Deliver(t.Context(), []domain.Notification{item}); err != nil {
			t.Fatalf("deliver: %v", err)
		}
	}
}

func TestListNotifications_RRES04_NewestFirstAcrossCursorPages(t *testing.T) {
	_, inbox := inboxWithUsers(t)
	deliver(t, inbox, notification(1, readerID), notification(2, readerID), notification(3, readerID),
		notification(4, strangerID))

	first, err := inbox.ListNotifications(t.Context(), domain.PageQuery{UserID: readerID, Limit: 2})
	if err != nil || len(first) != 2 || first[0].ID != notification(3, readerID).ID {
		t.Fatalf("first = %+v, err = %v", first, err)
	}
	after := pagination.Position{At: first[1].CreatedAt, ID: first[1].ID}
	second, err := inbox.ListNotifications(t.Context(), domain.PageQuery{UserID: readerID, After: after, Limit: 2})

	if err != nil || len(second) != 1 || second[0].ID != notification(1, readerID).ID || second[0].IsRead() {
		t.Fatalf("second = %+v, err = %v", second, err)
	}
}

func TestMarkRead_FRNTF04_OwnNotificationIsReadAndCountDrops(t *testing.T) {
	_, inbox := inboxWithUsers(t)
	deliver(t, inbox, notification(1, readerID), notification(2, readerID))

	err := inbox.MarkRead(t.Context(), domain.Read{
		UserID: readerID, NotificationID: notification(1, readerID).ID,
		At: base().Add(time.Hour),
	})
	count, countErr := inbox.CountUnread(t.Context(), readerID)

	if err != nil || countErr != nil || count != 1 {
		t.Fatalf("count = %d, err = %v, %v", count, err, countErr)
	}
}

func TestMarkRead_RSEC38_SomeoneElsesNotificationIsNotFound(t *testing.T) {
	_, inbox := inboxWithUsers(t)
	deliver(t, inbox, notification(1, readerID))

	err := inbox.MarkRead(t.Context(), domain.Read{
		UserID: strangerID, NotificationID: notification(1, readerID).ID,
		At: base(),
	})

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestMarkAllRead_FRNTF04_LeavesOtherUsersAlone(t *testing.T) {
	_, inbox := inboxWithUsers(t)
	deliver(t, inbox, notification(1, readerID), notification(2, strangerID))

	if err := inbox.MarkAllRead(t.Context(), domain.Read{UserID: readerID, At: base()}); err != nil {
		t.Fatalf("mark all: %v", err)
	}

	mine, mineErr := inbox.CountUnread(t.Context(), readerID)
	theirs, theirsErr := inbox.CountUnread(t.Context(), strangerID)
	if mineErr != nil || theirsErr != nil || mine != 0 || theirs != 1 {
		t.Fatalf("mine = %d, theirs = %d", mine, theirs)
	}
}

func TestDeliver_BRNTF01_RolledBackWithTheEventTransaction(t *testing.T) {
	pool, inbox := inboxWithUsers(t)
	eventFailed := errors.New("event failed")

	err := pool.WithinTx(t.Context(), func(ctx context.Context) error {
		if err := inbox.Deliver(ctx, []domain.Notification{notification(1, readerID)}); err != nil {
			return fmt.Errorf("deliver: %w", err)
		}
		return eventFailed
	})

	count, countErr := inbox.CountUnread(t.Context(), readerID)
	if !errors.Is(err, eventFailed) || countErr != nil || count != 0 {
		t.Fatalf("err = %v, count = %d", err, count)
	}
}
