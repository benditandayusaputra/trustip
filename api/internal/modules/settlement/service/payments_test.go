package service_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/service"
)

const (
	proofID    = "0192f0c4-0000-7000-8000-000000000601"
	paymentKey = "0192f0c4-0000-7000-8000-000000000999"
)

func (f *fakeSettlements) FindPaymentByKey(
	_ context.Context, key domain.PaymentKeyRef,
) (domain.StoredPayment, bool, error) {
	for _, payment := range f.payments {
		if payment.RecordedByOrg == key.OrganizationID && payment.IdempotencyKey == key.Key {
			return domain.StoredPayment{
				PaymentID: payment.ID, SettlementID: payment.SettlementID, Fingerprint: payment.Fingerprint,
			}, true, nil
		}
	}
	return domain.StoredPayment{}, false, nil
}

func (f *fakeSettlements) LockPayable(ctx context.Context, key domain.PayableKey) (domain.Payable, error) {
	view, err := f.GetSettlement(ctx, domain.SettlementKey{
		OrganizationID: key.OrganizationID, SettlementID: key.SettlementID,
	})
	if err != nil {
		return domain.Payable{}, err
	}
	payable := domain.Payable{
		SettlementID: view.ID, Number: view.Number, PartnershipID: view.PartnershipID, ConsignorID: view.ConsignorID,
		StoreID: view.StoreID, StoreName: "Warung Bu Sri", Status: view.Status,
		Remaining: view.Summary.Split.ConsignorShare, Methods: domain.Methods(),
	}
	for _, payment := range f.payments {
		isOpen := payment.Status != domain.PaymentRejected
		payable.HasOpenPayment = payable.HasOpenPayment || (payment.SettlementID == view.ID && isOpen)
	}
	if key.ProofFileID == proofID && key.OrganizationID == storeID {
		payable.Proof = domain.Proof{FileID: proofID, SHA256: bytes.Repeat([]byte{7}, 32)}
	}
	return payable, nil
}

func (f *fakeSettlements) CreatePayment(_ context.Context, payment domain.Payment) error {
	f.payments[payment.ID] = payment
	return nil
}

func (f *fakeSettlements) LockPayment(_ context.Context, key domain.PaymentKey) (domain.Payment, error) {
	payment, ok := f.payments[key.PaymentID]
	if !ok || (payment.ConsignorID != key.OrganizationID && payment.StoreID != key.OrganizationID) {
		return domain.Payment{}, apperror.OutsideScope()
	}
	return payment, nil
}

func (f *fakeSettlements) SaveReview(_ context.Context, review domain.Review) error {
	payment := f.payments[review.PaymentID]
	if payment.Status != domain.PendingConfirmation {
		return domain.ErrPaymentDecided
	}
	payment.Status = review.Status
	f.payments[review.PaymentID] = payment
	return nil
}

func (f *fakeSettlements) MarkPaid(_ context.Context, settlementID string, _ time.Time) error {
	f.paid = append(f.paid, settlementID)
	for index, settlement := range f.created {
		if settlement.ID == settlementID {
			f.created[index].Status = domain.Paid
		}
	}
	return nil
}

func (f *fakeSettlements) ApproveSources(_ context.Context, approval domain.Approval) error {
	f.approved = append(f.approved, approval)
	return nil
}

func consignorOwner() actor.Actor {
	return actor.Actor{UserID: "0192f0c4-0000-7000-8000-000000000502", Membership: actor.Membership{
		OrganizationID: consignorID, OrganizationType: actor.Consignor, Role: actor.Owner,
	}}
}

func issuedScenario(t *testing.T) (scenario, domain.Settlement) {
	t.Helper()
	run := newScenario(terms.ScheduleOnCount, 18)
	settlement, _, err := service.NewIssuer(run.deps).IssueOnCount(t.Context(), partnershipID, countedAt())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	return run, settlement
}

func transfer(settlementID string) service.RecordPaymentCommand {
	return service.RecordPaymentCommand{
		SettlementID: settlementID, IdempotencyKey: paymentKey,
		Input: domain.PaymentInput{Amount: 31500, Method: "TRANSFER", PaidOn: "2026-11-10", ProofFileID: proofID},
	}
}

func TestRecordPayment_FRSTL05_StoreOwnerRecordsATransferWithProof(t *testing.T) {
	run, settlement := issuedScenario(t)

	receipt, err := service.NewRecordPayment(run.deps).Execute(t.Context(), storeOwner(), transfer(settlement.ID))

	if err != nil || receipt.IsReplay || len(run.settlements.payments) != 1 {
		t.Fatalf("receipt = %+v, err = %v", receipt, err)
	}
	if run.events.seals[1].Event != "PAYMENT_RECORDED" || run.events.notices[2].Type != "PAYMENT_RECORDED" ||
		run.events.audits[2].Action != audit.PaymentRecorded {
		t.Fatalf("events = %+v", run.events)
	}
}

func TestRecordPayment_RSEC36_ConsignorAndStaffCannotRecordPayments(t *testing.T) {
	run, settlement := issuedScenario(t)
	staff := storeOwner()
	staff.Membership.Role = actor.Staff

	record := service.NewRecordPayment(run.deps)
	_, consignorErr := record.Execute(t.Context(), consignorOwner(), transfer(settlement.ID))
	_, staffErr := record.Execute(t.Context(), staff, transfer(settlement.ID))

	if codeOf(consignorErr) != apperror.CodeForbidden || codeOf(staffErr) != apperror.CodeForbidden {
		t.Fatalf("consignor = %v, staff = %v", consignorErr, staffErr)
	}
}

func TestRecordPayment_RRES06_SameKeyReplaysAndAnotherBodyIsRefused(t *testing.T) {
	run, settlement := issuedScenario(t)
	record := service.NewRecordPayment(run.deps)
	if _, err := record.Execute(t.Context(), storeOwner(), transfer(settlement.ID)); err != nil {
		t.Fatalf("first: %v", err)
	}
	changed := transfer(settlement.ID)
	changed.Input.Method, changed.Input.ProofFileID = "CASH", ""

	replayed, replayErr := record.Execute(t.Context(), storeOwner(), transfer(settlement.ID))
	_, changedErr := record.Execute(t.Context(), storeOwner(), changed)

	if replayErr != nil || !replayed.IsReplay || codeOf(changedErr) != apperror.CodeIdempotencyKeyReused {
		t.Fatalf("replayed = %+v, err = %v, changed = %v", replayed, replayErr, changedErr)
	}
}

func recordedPaymentID(t *testing.T, run scenario, settlementID string) string {
	t.Helper()
	_, err := service.NewRecordPayment(run.deps).Execute(t.Context(), storeOwner(), transfer(settlementID))
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	for id := range run.settlements.payments {
		return id
	}
	return ""
}

func TestConfirmPayment_BRSTL06_BRSTL08_SettlementIsPaidAndItsCountsApproved(t *testing.T) {
	run, settlement := issuedScenario(t)
	paymentID := recordedPaymentID(t, run, settlement.ID)

	view, err := service.NewConfirmPayment(run.deps).Execute(t.Context(), consignorOwner(), paymentID)
	_, again := service.NewConfirmPayment(run.deps).Execute(t.Context(), consignorOwner(), paymentID)

	if err != nil || view.Status != domain.Paid || len(run.settlements.approved) != 1 ||
		run.settlements.approved[0].SettlementID != settlement.ID {
		t.Fatalf("view = %+v, err = %v", view.Status, err)
	}
	if run.events.seals[2].Event != "PAYMENT_CONFIRMED" || codeOf(again) != apperror.CodeDocumentNotPending {
		t.Fatalf("seals = %+v, again = %v", run.events.seals, again)
	}
}

func TestRejectPayment_BRSTL06_RejectionKeepsTheSettlementOpenForANewPayment(t *testing.T) {
	run, settlement := issuedScenario(t)
	paymentID := recordedPaymentID(t, run, settlement.ID)

	_, err := service.NewRejectPayment(run.deps).Execute(t.Context(), consignorOwner(), service.RejectPaymentCommand{
		PaymentID: paymentID, Reason: "Belum masuk rekening",
	})
	retry := transfer(settlement.ID)
	retry.IdempotencyKey = "0192f0c4-0000-7000-8000-000000000998"
	_, retryErr := service.NewRecordPayment(run.deps).Execute(t.Context(), storeOwner(), retry)

	if err != nil || len(run.settlements.paid) != 0 || retryErr != nil {
		t.Fatalf("reject = %v, paid = %v, retry = %v", err, run.settlements.paid, retryErr)
	}
}

func TestMarkPaid_FRSTL07_ConsignorOwnerMarksCashReceived(t *testing.T) {
	run, settlement := issuedScenario(t)

	view, err := service.NewMarkPaid(run.deps).Execute(t.Context(), consignorOwner(), settlement.ID)
	_, storeErr := service.NewMarkPaid(run.deps).Execute(t.Context(), storeOwner(), settlement.ID)

	if err != nil || view.Status != domain.Paid || len(run.settlements.payments) != 1 ||
		run.events.seals[1].Event != "PAYMENT_CONFIRMED" || codeOf(storeErr) != apperror.CodeForbidden {
		t.Fatalf("view = %+v, err = %v, store = %v", view.Status, err, storeErr)
	}
}
