package service_test

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/service"
)

func (run scenario) recorded(t *testing.T) string {
	t.Helper()
	receipt, err := service.NewRecordCount(run.deps).Execute(t.Context(), cashier(), command(20))
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	return receipt.View.ID
}

func TestConfirmCount_FRCNT08_ConsignorConfirmsAndSealsIt(t *testing.T) {
	run := newScenario()
	countID := run.recorded(t)

	confirmed, err := service.NewConfirmCount(run.deps).Execute(t.Context(), consignorStaff(), countID)

	if err != nil || confirmed.Status != domain.Confirmed || len(run.events.seals) != 2 ||
		run.events.seals[1].Event != "COUNT_CONFIRMED" || run.events.audits[1].Action != audit.CountConfirmed {
		t.Fatalf("confirmed = %+v, err = %v, events = %+v", confirmed, err, run.events)
	}
}

func TestConfirmCount_SecondConfirmationIsNotPending(t *testing.T) {
	run := newScenario()
	countID := run.recorded(t)
	confirm := service.NewConfirmCount(run.deps)
	if _, err := confirm.Execute(t.Context(), consignorStaff(), countID); err != nil {
		t.Fatalf("first: %v", err)
	}

	_, err := confirm.Execute(t.Context(), consignorStaff(), countID)

	if codeOf(err) != apperror.CodeDocumentNotPending {
		t.Fatalf("err = %v", err)
	}
}

func TestConfirmCount_RSEC36_StoreMayNotConfirm(t *testing.T) {
	run := newScenario()
	countID := run.recorded(t)

	_, err := service.NewConfirmCount(run.deps).Execute(t.Context(), cashier(), countID)

	if codeOf(err) != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestListCounts_FRCNT10_ScopesToTheActorAndCutsOffFortyEightHoursAgo(t *testing.T) {
	run := newScenario()

	_, err := service.NewListCounts(run.deps).Execute(t.Context(), consignorStaff(), service.ListRequest{
		Filter: domain.ListFilter{PartnershipID: partnershipID},
	})

	query := run.counts.listed[0]
	if err != nil || query.OrganizationID != consignorID || query.PartnershipID != partnershipID ||
		!query.ConfirmCutoff.Equal(fixedClock{}.Now().Add(-48*time.Hour)) {
		t.Fatalf("query = %+v, err = %v", query, err)
	}
}
