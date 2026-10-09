package sse

import (
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
)

func newTestTickets() (*Tickets, *clock.Manual) {
	manual := clock.NewManual(time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC))
	return NewTickets(manual, 30*time.Second), manual
}

func TestIssue_RSEC26_TicketLivesThirtySeconds(t *testing.T) {
	tickets, manual := newTestTickets()

	ticket := tickets.Issue("user-a")

	if !ticket.ExpiresAt.Equal(manual.Now().Add(30 * time.Second)) {
		t.Fatalf("expires at %v", ticket.ExpiresAt)
	}
	if len(ticket.Value) < 26 {
		t.Fatalf("ticket %q is too short to be unguessable", ticket.Value)
	}
}

func TestRedeem_ValidTicket_ReturnsItsUser(t *testing.T) {
	tickets, _ := newTestTickets()
	ticket := tickets.Issue("user-a")

	userID, err := tickets.Redeem(ticket.Value)
	if err != nil {
		t.Fatalf("redeem: %v", err)
	}

	if userID != "user-a" {
		t.Fatalf("user = %q", userID)
	}
}

func TestRedeem_RSEC26_TicketWorksOnce(t *testing.T) {
	tickets, _ := newTestTickets()
	ticket := tickets.Issue("user-a")
	if _, err := tickets.Redeem(ticket.Value); err != nil {
		t.Fatalf("first redeem: %v", err)
	}

	_, err := tickets.Redeem(ticket.Value)

	if !errors.Is(err, ErrTicketInvalid) {
		t.Fatalf("err = %v, want ErrTicketInvalid", err)
	}
}

func TestRedeem_RSEC26_AtExpiry_IsRefused(t *testing.T) {
	tickets, manual := newTestTickets()
	ticket := tickets.Issue("user-a")

	manual.Advance(30 * time.Second)
	_, err := tickets.Redeem(ticket.Value)

	if !errors.Is(err, ErrTicketInvalid) {
		t.Fatalf("err = %v, want ErrTicketInvalid", err)
	}
}

func TestRedeem_OneSecondBeforeExpiry_IsAccepted(t *testing.T) {
	tickets, manual := newTestTickets()
	ticket := tickets.Issue("user-a")

	manual.Advance(29 * time.Second)

	if _, err := tickets.Redeem(ticket.Value); err != nil {
		t.Fatalf("redeem: %v", err)
	}
}

func TestRedeem_UnknownTicket_IsRefused(t *testing.T) {
	tickets, _ := newTestTickets()

	_, err := tickets.Redeem("not-a-ticket")

	if !errors.Is(err, ErrTicketInvalid) {
		t.Fatalf("err = %v, want ErrTicketInvalid", err)
	}
}

func TestIssue_SweepsExpiredTickets(t *testing.T) {
	tickets, manual := newTestTickets()
	tickets.Issue("user-a")
	manual.Advance(time.Minute)

	tickets.Issue("user-b")

	if size := tickets.size(); size != 1 {
		t.Fatalf("tickets kept = %d, want 1", size)
	}
}
