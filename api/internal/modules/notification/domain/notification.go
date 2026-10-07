package domain

import (
	"errors"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
)

const appArea = "/app"

var (
	ErrUnsafeLink = errors.New("notification link must point inside the app area")
	typeFormat    = regexp.MustCompile(`^[A-Z_]+$`)
)

type Recipient struct {
	UserID string
	Role   actor.Role
}

type Draft struct {
	Type         string
	Title        string
	Body         string
	Link         string
	CarriesMoney bool
}

type DraftInput Draft

type Audience struct {
	Candidates []Recipient
	ActorID    string
}

type Notification struct {
	ID        string
	UserID    string
	Type      string
	Title     string
	Body      string
	Link      string
	ReadAt    time.Time
	CreatedAt time.Time
}

func (n Notification) IsRead() bool {
	return !n.ReadAt.IsZero()
}

func NewDraft(input DraftInput) (Draft, error) {
	if !typeFormat.MatchString(input.Type) || !isAppLink(input.Link) {
		return Draft{}, ErrUnsafeLink
	}
	return Draft(input), nil
}

func isAppLink(link string) bool {
	if !strings.HasPrefix(link, appArea) || strings.ContainsAny(link, "\\:") {
		return false
	}
	cleaned := path.Clean(link)
	return cleaned == appArea || strings.HasPrefix(cleaned, appArea+"/")
}

func RecipientsOf(draft Draft, audience Audience) []Recipient {
	seen := map[string]bool{audience.ActorID: true}
	recipients := []Recipient{}
	for _, candidate := range audience.Candidates {
		if seen[candidate.UserID] || (draft.CarriesMoney && candidate.Role != actor.Owner) {
			continue
		}
		seen[candidate.UserID] = true
		recipients = append(recipients, candidate)
	}
	return recipients
}

func Deliveries(draft Draft, recipients []Recipient, at time.Time) []Notification {
	notifications := make([]Notification, 0, len(recipients))
	for _, recipient := range recipients {
		notifications = append(notifications, Notification{
			UserID: recipient.UserID, Type: draft.Type, Title: draft.Title, Body: draft.Body, Link: draft.Link,
			CreatedAt: at,
		})
	}
	return notifications
}

type Read struct {
	UserID         string
	NotificationID string
	At             time.Time
}

type PageQuery struct {
	UserID string
	After  pagination.Position
	Limit  int
}
