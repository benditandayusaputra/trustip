package domain

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const consignorStaffID = "0192f0c4-0000-7000-8000-000000000502"

func aView(t *testing.T) View {
	t.Helper()
	return View{Count: aCount(t)}
}

func TestEffective_FRCNT08_UnansweredCountIsDeemedConfirmedAfterFortyEightHours(t *testing.T) {
	view := aView(t)
	deadline := view.ReceivedAt.Add(48 * time.Hour)

	before, after := view.Effective(deadline.Add(-time.Second)), view.Effective(deadline)

	if before.Status != Recorded || after.Status != Confirmed || !after.AutoConfirmed ||
		!after.ConfirmedAt.Equal(deadline) {
		t.Fatalf("before = %+v, after = %+v", before.Status, after)
	}
}

func TestConfirm_FRCNT08_ConsignorConfirmsTheCount(t *testing.T) {
	view := aView(t)
	at := view.ReceivedAt.Add(time.Hour)

	decision, err := view.Confirm(Decider{UserID: consignorStaffID, OrganizationID: consignorID}, at)

	if err != nil || decision.Event != "COUNT_CONFIRMED" || decision.By != consignorStaffID || !decision.At.Equal(at) {
		t.Fatalf("decision = %+v, err = %v", decision, err)
	}
	request := view.SealRequestOf(decision)
	if request.Event != "COUNT_CONFIRMED" || request.ActorOrgID != consignorID ||
		request.Content["number"] != view.Number {
		t.Fatalf("request = %+v", request)
	}
}

func TestConfirm_StoreCannotConfirmItsOwnCount(t *testing.T) {
	view := aView(t)

	_, err := view.Confirm(Decider{UserID: cashierID, OrganizationID: storeID}, view.ReceivedAt)

	if codeOf(err) != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestConfirm_FRCNT08_AfterTheDeadlineNothingIsPending(t *testing.T) {
	view := aView(t)

	_, err := view.Confirm(Decider{UserID: consignorStaffID, OrganizationID: consignorID},
		view.ReceivedAt.Add(48*time.Hour))

	if codeOf(err) != apperror.CodeDocumentNotPending {
		t.Fatalf("err = %v", err)
	}
}

func TestNewListQuery_FRCNT10_FiltersByPartnerAndJakartaDays(t *testing.T) {
	query, err := NewListQuery(ListFilter{PartnershipID: partnershipID, From: "2026-11-01", To: "2026-11-10"})

	jakarta := noon().Location()
	if err != nil || !query.From.Equal(time.Date(2026, 11, 1, 0, 0, 0, 0, jakarta)) ||
		!query.To.Equal(time.Date(2026, 11, 11, 0, 0, 0, 0, jakarta)) || query.PartnershipID != partnershipID {
		t.Fatalf("query = %+v, err = %v", query, err)
	}
}

func TestNewListQuery_RefusesUnknownFilters(t *testing.T) {
	for name, filter := range map[string]ListFilter{
		"partnership_id": {PartnershipID: "x"}, "from": {From: "x"}, "to": {From: "2026-11-10", To: "2026-11-01"},
	} {
		_, err := NewListQuery(filter)

		if _, found := fieldCodes(t, err)[name]; !found {
			t.Fatalf("%s: codes = %v", name, fieldCodes(t, err))
		}
	}
}
