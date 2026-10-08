package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	MaxOpenPerStore      = 5
	ShowingPeriod        = 30 * 24 * time.Hour
	ExpiryReminderBefore = 3 * 24 * time.Hour
)

type Status string

const (
	Open      Status = "OPEN"
	Fulfilled Status = "FULFILLED"
	Closed    Status = "CLOSED"
	Expired   Status = "EXPIRED"
)

type Posting struct {
	ID        string
	StoreID   string
	CreatedBy string
	Draft     Draft
	At        time.Time
	ExpiresAt time.Time
}

type Author struct {
	NeedID  string
	StoreID string
	UserID  string
}

func (d Draft) Post(author Author, at time.Time) Posting {
	return Posting{
		ID: author.NeedID, StoreID: author.StoreID, CreatedBy: author.UserID, Draft: d, At: at,
		ExpiresAt: at.Add(ShowingPeriod),
	}
}

type Need struct {
	ID            string
	StoreID       string
	StoreName     string
	CategoryName  string
	Draft         Draft
	Status        Status
	ExpiresAt     time.Time
	CreatedAt     time.Time
	ResponseCount int
}

func (n Need) StatusAt(now time.Time) Status {
	if n.Status == Open && !now.Before(n.ExpiresAt) {
		return Expired
	}
	return n.Status
}

type Change struct {
	NeedID    string
	StoreID   string
	Status    Status
	ExpiresAt time.Time
	At        time.Time
	Previous  Status
}

func (n Need) Extend(now time.Time) (Change, error) {
	current := n.StatusAt(now)
	if current != Open && current != Expired {
		return Change{}, apperror.New(apperror.CodeNeedNotOpen)
	}
	return n.change(Open, now.Add(ShowingPeriod), now), nil
}

func (n Need) Fulfill(now time.Time) (Change, error) {
	return n.finish(Fulfilled, now)
}

func (n Need) Close(now time.Time) (Change, error) {
	return n.finish(Closed, now)
}

func (n Need) finish(status Status, now time.Time) (Change, error) {
	if n.StatusAt(now) != Open {
		return Change{}, apperror.New(apperror.CodeNeedNotOpen)
	}
	return n.change(status, n.ExpiresAt, now), nil
}

func (n Need) change(status Status, expiresAt, at time.Time) Change {
	return Change{NeedID: n.ID, StoreID: n.StoreID, Status: status, ExpiresAt: expiresAt, At: at, Previous: n.Status}
}

type NeedKey struct {
	StoreID string
	NeedID  string
}

func (n Need) AsOf(now time.Time) Need {
	n.Status = n.StatusAt(now)
	return n
}
