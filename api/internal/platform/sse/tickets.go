package sse

import (
	"crypto/rand"
	"errors"
	"sync"
	"time"
)

var ErrTicketInvalid = errors.New("stream ticket is unknown, used, or expired")

type Clock interface {
	Now() time.Time
}

type Ticket struct {
	Value     string
	ExpiresAt time.Time
}

type issuedTicket struct {
	userID    string
	expiresAt time.Time
}

type Tickets struct {
	mu       sync.Mutex
	clock    Clock
	lifetime time.Duration
	issued   map[string]issuedTicket
}

func NewTickets(clock Clock, lifetime time.Duration) *Tickets {
	return &Tickets{clock: clock, lifetime: lifetime, issued: map[string]issuedTicket{}}
}

func (t *Tickets) Issue(userID string) Ticket {
	now := t.clock.Now()
	ticket := Ticket{Value: rand.Text(), ExpiresAt: now.Add(t.lifetime)}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sweepExpired(now)
	t.issued[ticket.Value] = issuedTicket{userID: userID, expiresAt: ticket.ExpiresAt}
	return ticket
}

func (t *Tickets) Redeem(value string) (string, error) {
	now := t.clock.Now()
	t.mu.Lock()
	defer t.mu.Unlock()
	issued, ok := t.issued[value]
	delete(t.issued, value)
	if !ok || !now.Before(issued.expiresAt) {
		return "", ErrTicketInvalid
	}
	return issued.userID, nil
}

func (t *Tickets) sweepExpired(now time.Time) {
	for value, issued := range t.issued {
		if !now.Before(issued.expiresAt) {
			delete(t.issued, value)
		}
	}
}

func (t *Tickets) size() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.issued)
}
