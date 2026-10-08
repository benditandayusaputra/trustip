package service_test

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/service"
)

func TestConfirmDrop_FRDRP06_ConsignorConfirmsSealsAndTellsTheRecorder(t *testing.T) {
	scenario := newWorld()
	dropID := scenario.recorded(t)

	confirmed, err := service.NewConfirmDrop(scenario.deps).Execute(t.Context(), consignorOwner(), dropID)

	if err != nil || confirmed.Status != domain.Confirmed {
		t.Fatalf("confirmed = %+v, err = %v", confirmed, err)
	}
	events := scenario.recorder
	if len(events.seals) != 2 || events.seals[1].Event != "DROP_CONFIRMED" || events.audits[1].Action !=
		audit.DropConfirmed || events.notices[1].RecipientUserID != cashierID {
		t.Fatalf("events = %+v", events)
	}
}

func TestConfirmDrop_RSEC36_StoreMayNotConfirm(t *testing.T) {
	scenario := newWorld()
	dropID := scenario.recorded(t)

	_, err := service.NewConfirmDrop(scenario.deps).Execute(t.Context(), cashier(), dropID)

	if codeOf(err) != apperror.CodeForbidden {
		t.Fatalf("err = %v, want FORBIDDEN", err)
	}
}

func TestConfirmDrop_SecondConfirmationIsNotPending(t *testing.T) {
	scenario := newWorld()
	dropID := scenario.recorded(t)
	confirm := service.NewConfirmDrop(scenario.deps)
	if _, err := confirm.Execute(t.Context(), consignorOwner(), dropID); err != nil {
		t.Fatalf("first: %v", err)
	}

	_, err := confirm.Execute(t.Context(), consignorOwner(), dropID)

	if codeOf(err) != apperror.CodeDocumentNotPending {
		t.Fatalf("err = %v, want DOCUMENT_NOT_PENDING", err)
	}
}

func TestGetDrop_BRDRP08_ReadsAsConfirmedAfterFortyEightHours(t *testing.T) {
	scenario := newWorld()
	dropID := scenario.recorded(t)
	*scenario.clock = scenario.now.Add(48 * time.Hour)

	view, err := service.NewGetDrop(scenario.deps).Execute(t.Context(), consignorOwner(), dropID)

	if err != nil || view.Status != domain.Confirmed || !view.AutoConfirmed {
		t.Fatalf("view = %+v, err = %v", view, err)
	}
}

func TestCancelDrop_FRDRP10_StoreCancelsAndTheStockGoesBack(t *testing.T) {
	scenario := newWorld()
	dropID := scenario.recorded(t)
	*scenario.clock = scenario.now.Add(10 * time.Minute)

	cancelled, err := service.NewCancelDrop(scenario.deps).Execute(t.Context(), cashier(), dropID)

	if err != nil || cancelled.Status != domain.Cancelled {
		t.Fatalf("cancelled = %+v, err = %v", cancelled, err)
	}
	if len(scenario.drops.reversals) != 1 || scenario.drops.reversals[0].Qty != -20 ||
		scenario.recorder.seals[1].Event != "DROP_CANCELLED" || len(scenario.recorder.notices) != 1 {
		t.Fatalf("reversals = %+v, events = %+v", scenario.drops.reversals, scenario.recorder)
	}
}

func TestCancelDrop_FRDRP10_CountedStockCannotBeCancelled(t *testing.T) {
	scenario := newWorld()
	dropID := scenario.recorded(t)
	scenario.drops.stockUsed = true

	_, err := service.NewCancelDrop(scenario.deps).Execute(t.Context(), cashier(), dropID)

	if codeOf(err) != apperror.CodeDropAlreadyCounted || len(scenario.drops.reversals) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestCancelDrop_RSEC36_ConsignorMayNotCancel(t *testing.T) {
	scenario := newWorld()
	dropID := scenario.recorded(t)

	_, err := service.NewCancelDrop(scenario.deps).Execute(t.Context(), consignorOwner(), dropID)

	if codeOf(err) != apperror.CodeForbidden {
		t.Fatalf("err = %v, want FORBIDDEN", err)
	}
}

func TestCancelDrop_BRAUTHZ02_OtherOrganizationsDropIsNotFound(t *testing.T) {
	scenario := newWorld()
	dropID := scenario.recorded(t)
	stranger := cashier()
	stranger.Membership.OrganizationID = "0192f0c4-0000-7000-8000-000000000409"

	_, err := service.NewCancelDrop(scenario.deps).Execute(t.Context(), stranger, dropID)

	if codeOf(err) != apperror.CodeNotFound {
		t.Fatalf("err = %v, want NOT_FOUND", err)
	}
}
