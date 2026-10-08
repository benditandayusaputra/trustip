package domain

import (
	"bytes"
	"errors"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/canonical"
)

const (
	storeOwnerID     = "0192f0c4-0000-7000-8000-000000000501"
	consignorOwnerID = "0192f0c4-0000-7000-8000-000000000502"
	proofFileID      = "0192f0c4-0000-7000-8000-000000000601"
)

func aPayable() Payable {
	return Payable{
		SettlementID: "0192f0c4-0000-7000-8000-000000000701", Number: "ST-202611-0001", PartnershipID: partnershipID,
		ConsignorID: consignorID, StoreID: storeID, StoreName: "Warung Bu Sri", Status: Issued, Remaining: 691187,
		Proof: Proof{FileID: proofFileID, SHA256: bytes.Repeat([]byte{7}, 32)},
	}
}

func aTransfer() PaymentInput {
	return PaymentInput{Amount: 691187, Method: "TRANSFER", PaidOn: "2026-11-20", ProofFileID: proofFileID}
}

func codeOf(err error) apperror.Code {
	if appErr, ok := errors.AsType[*apperror.Error](err); ok {
		return appErr.Code()
	}
	return ""
}

func TestCheckPayment_BRSTL05_TransferForTheRemainingAmountPasses(t *testing.T) {
	draft, err := aPayable().CheckPayment(aTransfer(), jakarta("2026-11-20 10:00"))

	if err != nil || draft.Amount != 691187 || draft.Method != Transfer || !draft.PaidOn.Equal(jakarta("2026-11-20")) {
		t.Fatalf("draft = %+v, err = %v", draft, err)
	}
}

func TestCheckPayment_BRSTL05_AmountMustBeTheRemainingSettlement(t *testing.T) {
	input := aTransfer()
	input.Amount = 691186

	_, err := aPayable().CheckPayment(input, jakarta("2026-11-20 10:00"))

	appErr, _ := errors.AsType[*apperror.Error](err)
	if fieldCodes(t, err)["amount"] != "PAYMENT_AMOUNT_MISMATCH" ||
		appErr.Fields()[0].Message != "Jumlah harus sama dengan sisa Setoran (Rp691.187)." {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckPayment_BRSTL05_ProofIsRequiredForTransferAndQRISButNotCash(t *testing.T) {
	for method, code := range map[string]string{"TRANSFER": "PROOF_REQUIRED", "QRIS": "PROOF_REQUIRED", "CASH": ""} {
		input := aTransfer()
		input.Method, input.ProofFileID = method, ""
		payable := aPayable()
		payable.Proof = Proof{}

		_, err := payable.CheckPayment(input, jakarta("2026-11-20 10:00"))

		if code == "" && err != nil {
			t.Fatalf("%s: err = %v", method, err)
		}
		if code != "" && fieldCodes(t, err)["proof_file_id"] != code {
			t.Fatalf("%s: codes = %v", method, fieldCodes(t, err))
		}
	}
}

func TestCheckPayment_ProofMustBeAnUploadOfTheStore(t *testing.T) {
	payable := aPayable()
	payable.Proof = Proof{}

	_, err := payable.CheckPayment(aTransfer(), jakarta("2026-11-20 10:00"))

	if fieldCodes(t, err)["proof_file_id"] != "UNKNOWN_CHOICE" {
		t.Fatalf("codes = %v", fieldCodes(t, err))
	}
}

func TestCheckPayment_FRSTL05_MethodAndDateAreChecked(t *testing.T) {
	cases := map[string]PaymentInput{
		"method":  {Amount: 691187, Method: "CHEQUE", PaidOn: "2026-11-20"},
		"paid_on": {Amount: 691187, Method: "CASH", PaidOn: "2026-11-21"},
	}
	for name, input := range cases {
		_, err := aPayable().CheckPayment(input, jakarta("2026-11-20 10:00"))
		if _, found := fieldCodes(t, err)[name]; !found {
			t.Fatalf("%s: codes = %v", name, fieldCodes(t, err))
		}
	}
}

func TestCheckPayment_BRSTL06_PaidOrPendingSettlementTakesNoNewPayment(t *testing.T) {
	paid, pending := aPayable(), aPayable()
	paid.Status = Paid
	pending.HasOpenPayment = true

	_, paidErr := paid.CheckPayment(aTransfer(), jakarta("2026-11-20 10:00"))
	_, pendingErr := pending.CheckPayment(aTransfer(), jakarta("2026-11-20 10:00"))

	if codeOf(paidErr) != apperror.CodeSettlementPaid || codeOf(pendingErr) != apperror.CodePaymentPending {
		t.Fatalf("paid = %v, pending = %v", paidErr, pendingErr)
	}
}

func aRecordedPayment(t *testing.T) Payment {
	t.Helper()
	draft, err := aPayable().CheckPayment(aTransfer(), jakarta("2026-11-20 10:00"))
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	return draft.Record(PaymentRecording{
		ID: "0192f0c4-0000-7000-8000-000000000801", RecordedBy: storeOwnerID, IdempotencyKey: "key",
		Fingerprint: []byte{1}, At: jakarta("2026-11-20 10:00"),
	})
}

func TestRecord_RSEC35_SealedContentLocksTheProofHash(t *testing.T) {
	payment := aRecordedPayment(t)
	swapped := payment
	swapped.Proof.SHA256 = bytes.Repeat([]byte{8}, 32)

	request := payment.SealRequest()
	original, originalErr := canonical.Digest(payment.Content())
	changed, changedErr := canonical.Digest(swapped.Content())

	if request.Event != "PAYMENT_RECORDED" || request.DocumentType != "PAYMENT" ||
		request.DocumentNumber != "ST-202611-0001" || request.ActorOrgID != storeID {
		t.Fatalf("request = %+v", request)
	}
	if errors.Join(originalErr, changedErr) != nil || bytes.Equal(original, changed) {
		t.Fatalf("digests = %x %x", original, changed)
	}
}

func TestRecord_FRSTL05_ConsignorOwnersHearOfThePayment(t *testing.T) {
	notice := aRecordedPayment(t).RecordedNotice()

	if notice.OrganizationID != consignorID || !notice.OwnersOnly || notice.Type != "PAYMENT_RECORDED" ||
		notice.Body != "Warung Bu Sri mencatat pembayaran Rp691.187 untuk ST-202611-0001." {
		t.Fatalf("notice = %+v", notice)
	}
}

func TestReview_FRSTL06_ConsignorConfirmsOrRejectsWithAReason(t *testing.T) {
	payment := aRecordedPayment(t)
	consignor := Decider{UserID: consignorOwnerID, OrganizationID: consignorID}

	confirmed, confirmErr := payment.Confirm(consignor, jakarta("2026-11-21 08:00"))
	rejected, rejectErr := payment.Reject(consignor, "Belum masuk rekening", jakarta("2026-11-21 08:00"))
	_, emptyErr := payment.Reject(consignor, "", jakarta("2026-11-21 08:00"))
	_, storeErr := payment.Confirm(Decider{UserID: storeOwnerID, OrganizationID: storeID}, jakarta("2026-11-21 08:00"))

	if confirmErr != nil || confirmed.Status != PaymentConfirmed || confirmed.Event != "PAYMENT_CONFIRMED" {
		t.Fatalf("confirmed = %+v, err = %v", confirmed, confirmErr)
	}
	if rejectErr != nil || rejected.Status != PaymentRejected || rejected.Reason != "Belum masuk rekening" {
		t.Fatalf("rejected = %+v, err = %v", rejected, rejectErr)
	}
	if fieldCodes(t, emptyErr)["reason"] != "REQUIRED" || codeOf(storeErr) != apperror.CodeForbidden {
		t.Fatalf("empty = %v, store = %v", emptyErr, storeErr)
	}
}

func TestReview_ReviewedPaymentIsNoLongerPending(t *testing.T) {
	payment := aRecordedPayment(t)
	payment.Status = PaymentRejected

	_, err := payment.Confirm(Decider{UserID: consignorOwnerID, OrganizationID: consignorID}, jakarta("2026-11-21"))

	if codeOf(err) != apperror.CodeDocumentNotPending {
		t.Fatalf("err = %v", err)
	}
}

func TestReviewNotice_FRSTL06_StoreOwnersHearTheAnswer(t *testing.T) {
	payment := aRecordedPayment(t)
	review, err := payment.Reject(Decider{UserID: consignorOwnerID, OrganizationID: consignorID}, "Belum masuk",
		jakarta("2026-11-21 08:00"))
	if err != nil {
		t.Fatalf("reject: %v", err)
	}

	notice := payment.ReviewNotice(review)

	if notice.OrganizationID != storeID || !notice.OwnersOnly || notice.Type != "PAYMENT_REJECTED" ||
		notice.Body != "Pembayaran Rp691.187 untuk ST-202611-0001 ditolak: Belum masuk." {
		t.Fatalf("notice = %+v", notice)
	}
}

func TestSettleDirectly_FRSTL07_ConsignorMarksCashReceivedAsConfirmed(t *testing.T) {
	payment, err := aPayable().SettleDirectly(Decider{UserID: consignorOwnerID, OrganizationID: consignorID},
		PaymentRecording{ID: "p", IdempotencyKey: "k", Fingerprint: []byte{1}, At: jakarta("2026-11-20 10:00")})

	if err != nil || payment.Status != PaymentConfirmed || payment.Method != Cash || payment.Amount != 691187 ||
		payment.RecordedByOrg != consignorID || !payment.PaidOn.Equal(jakarta("2026-11-20")) {
		t.Fatalf("payment = %+v, err = %v", payment, err)
	}
	if payment.SealRequest().Event != "PAYMENT_CONFIRMED" {
		t.Fatalf("seal = %+v", payment.SealRequest())
	}
}

func TestSettleDirectly_StoreCannotMarkItsOwnSettlementPaid(t *testing.T) {
	_, err := aPayable().SettleDirectly(Decider{UserID: storeOwnerID, OrganizationID: storeID},
		PaymentRecording{ID: "p", At: jakarta("2026-11-20 10:00")})

	if codeOf(err) != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}
