package domain

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const ownerID = "0192f0c4-0000-7000-8000-000000000502"

func aPendingView(t *testing.T) View {
	t.Helper()
	return View{Drop: aDrop(t), StoreName: "Warung Bu Sri", ConsignorName: "Roti Pak Anton"}
}

func consignorDecider() Decider {
	return Decider{UserID: ownerID, OrganizationID: consignorID}
}

func storeDecider() Decider {
	return Decider{UserID: cashierID, OrganizationID: storeID}
}

func TestEffective_BRDRP08_RecordedDropIsDeemedConfirmedAfterFortyEightHours(t *testing.T) {
	view := aPendingView(t)
	deadline := view.ReceivedAt.Add(48 * time.Hour)

	before := view.Effective(deadline.Add(-time.Second))
	after := view.Effective(deadline)

	if before.Status != Recorded || before.IsConfirmed {
		t.Fatalf("before = %+v", before)
	}
	if after.Status != Confirmed || !after.AutoConfirmed || !after.ConfirmedAt.Equal(deadline) {
		t.Fatalf("after = %+v", after)
	}
}

func TestConfirm_FRDRP06_ConsignorConfirmsAndTheRecorderIsTold(t *testing.T) {
	view := aPendingView(t)
	at := view.ReceivedAt.Add(time.Hour)

	decision, err := view.Confirm(consignorDecider(), at)

	if err != nil || decision.Status != Confirmed || decision.By != ownerID || !decision.At.Equal(at) ||
		decision.Event != "DROP_CONFIRMED" {
		t.Fatalf("decision = %+v, err = %v", decision, err)
	}
	notice := decision.Notice
	if notice.RecipientUserID != cashierID || notice.Type != "DROP_CONFIRMED" ||
		notice.Body != "Roti Pak Anton menyatakan TM-202611-0007 sesuai." {
		t.Fatalf("notice = %+v", notice)
	}
}

func TestConfirm_BRDRP08_AfterTheDeadlineTheDropIsAlreadyConfirmed(t *testing.T) {
	view := aPendingView(t)

	_, err := view.Confirm(consignorDecider(), view.ReceivedAt.Add(48*time.Hour))

	if codeOf(err) != apperror.CodeDocumentNotPending {
		t.Fatalf("err = %v, want DOCUMENT_NOT_PENDING", err)
	}
}

func TestConfirm_StoreCannotConfirmItsOwnDrop(t *testing.T) {
	view := aPendingView(t)

	_, err := view.Confirm(storeDecider(), view.ReceivedAt)

	if codeOf(err) != apperror.CodeForbidden {
		t.Fatalf("err = %v, want FORBIDDEN", err)
	}
}

func TestCancel_FRDRP10_StoreCancelsWithinFifteenMinutes(t *testing.T) {
	view := aPendingView(t)
	at := view.ReceivedAt.Add(15 * time.Minute)

	decision, err := view.Cancel(Cancellation{Decider: storeDecider(), At: at})

	if err != nil || decision.Status != Cancelled || decision.Event != "DROP_CANCELLED" || decision.Notice.Type != "" {
		t.Fatalf("decision = %+v, err = %v", decision, err)
	}
}

func TestCancel_FRDRP10_RefusedOneSecondAfterFifteenMinutes(t *testing.T) {
	view := aPendingView(t)

	_, err := view.Cancel(Cancellation{Decider: storeDecider(), At: view.ReceivedAt.Add(15*time.Minute + time.Second)})

	if codeOf(err) != apperror.CodeCancelWindowPassed {
		t.Fatalf("err = %v, want CANCEL_WINDOW_PASSED", err)
	}
}

func TestCancel_FRDRP10_RefusedOnceTheStockWasCounted(t *testing.T) {
	view := aPendingView(t)

	_, err := view.Cancel(Cancellation{Decider: storeDecider(), At: view.ReceivedAt, IsStockUsed: true})

	if codeOf(err) != apperror.CodeDropAlreadyCounted {
		t.Fatalf("err = %v, want DROP_ALREADY_COUNTED", err)
	}
}

func TestCancel_RefusedAfterTheConsignorConfirmed(t *testing.T) {
	view := aPendingView(t)
	view.Status = Confirmed

	_, err := view.Cancel(Cancellation{Decider: storeDecider(), At: view.ReceivedAt})

	if codeOf(err) != apperror.CodeDocumentNotPending {
		t.Fatalf("err = %v, want DOCUMENT_NOT_PENDING", err)
	}
}

func TestCancel_ConsignorCannotCancel(t *testing.T) {
	view := aPendingView(t)

	_, err := view.Cancel(Cancellation{Decider: consignorDecider(), At: view.ReceivedAt})

	if codeOf(err) != apperror.CodeForbidden {
		t.Fatalf("err = %v, want FORBIDDEN", err)
	}
}

func TestReversal_FRDRP10_TakesEveryBatchBackOut(t *testing.T) {
	view := aPendingView(t)
	decision, err := view.Cancel(Cancellation{Decider: storeDecider(), At: view.ReceivedAt})
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}

	movements := decision.Reversal([]string{"back-1", "back-2"})

	if len(movements) != 2 || movements[0].Qty != -20 || movements[0].Type != "ADJUSTMENT" ||
		movements[1].ID != "back-2" || movements[1].DropItemID != "item-2" || movements[1].CreatedBy != cashierID {
		t.Fatalf("movements = %+v", movements)
	}
}

func TestSealRequestOf_DecisionsSealTheSameContent(t *testing.T) {
	view := aPendingView(t)
	decision, err := view.Confirm(consignorDecider(), view.ReceivedAt)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}

	request := view.SealRequestOf(decision)

	if request.Event != "DROP_CONFIRMED" || request.ActorOrgID != consignorID || request.ActorUserID != ownerID ||
		request.Content["number"] != "TM-202611-0007" {
		t.Fatalf("request = %+v", request)
	}
}
