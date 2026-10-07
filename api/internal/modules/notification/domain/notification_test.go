package domain

import (
	"slices"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
)

func candidates() []Recipient {
	return []Recipient{
		{UserID: "owner-1", Role: actor.Owner},
		{UserID: "staff-1", Role: actor.Staff},
		{UserID: "owner-2", Role: actor.Owner},
	}
}

func userIDs(recipients []Recipient) []string {
	ids := make([]string, 0, len(recipients))
	for _, recipient := range recipients {
		ids = append(ids, recipient.UserID)
	}
	return ids
}

func TestRecipientsOf_BRNTF02_ActorDoesNotHearAboutTheirOwnAction(t *testing.T) {
	got := RecipientsOf(Draft{Type: "DROP_RECORDED"}, Audience{Candidates: candidates(), ActorID: "owner-1"})

	if !slices.Equal(userIDs(got), []string{"staff-1", "owner-2"}) {
		t.Fatalf("recipients = %v", userIDs(got))
	}
}

func TestRecipientsOf_BRNTF03_MoneyTotalsReachOwnersOnly(t *testing.T) {
	draft := Draft{Type: "SETTLEMENT_ISSUED", CarriesMoney: true}

	got := RecipientsOf(draft, Audience{Candidates: candidates(), ActorID: "someone-else"})

	if !slices.Equal(userIDs(got), []string{"owner-1", "owner-2"}) {
		t.Fatalf("recipients = %v", userIDs(got))
	}
}

func TestRecipientsOf_EachUserIsNotifiedOnce(t *testing.T) {
	repeated := append(candidates(), Recipient{UserID: "staff-1", Role: actor.Staff})

	got := RecipientsOf(Draft{Type: "DROP_RECORDED"}, Audience{Candidates: repeated})

	if !slices.Equal(userIDs(got), []string{"owner-1", "staff-1", "owner-2"}) {
		t.Fatalf("recipients = %v", userIDs(got))
	}
}

func TestNewDraft_FRNTF05_LinkMustStayInsideTheApp(t *testing.T) {
	for link, valid := range map[string]bool{
		"/app/titipan/0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e": true,
		"/app/ajuan":           true,
		"https://evil.example": false,
		"//evil.example/app":   false,
		"/masuk":               false,
		"/app/../masuk":        false,
		"javascript:alert(1)":  false,
	} {
		_, err := NewDraft(DraftInput{Type: "PROPOSAL_RECEIVED", Title: "Ajuan baru", Body: "x", Link: link})

		if (err == nil) != valid {
			t.Errorf("%q: err = %v", link, err)
		}
	}
}

func TestDeliveries_CopyTheDraftForEveryRecipient(t *testing.T) {
	draft := Draft{
		Type: "PROPOSAL_RECEIVED", Title: "Ajuan baru", Body: "Warung Bu Tini mengajukan kerja sama.",
		Link: "/app/ajuan",
	}
	at := time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)

	got := Deliveries(draft, []Recipient{{UserID: "u1"}, {UserID: "u2"}}, at)

	if len(got) != 2 || got[1].UserID != "u2" || got[1].Title != "Ajuan baru" || !got[0].CreatedAt.Equal(at) {
		t.Fatalf("deliveries = %+v", got)
	}
}
