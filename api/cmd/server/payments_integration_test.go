package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
)

const paymentKey = "0192f0c4-0000-7000-8000-00000000d001"

func (b *browser) recordPayment(settlementID, body, key string) exchange {
	b.t.Helper()
	req := httptest.NewRequestWithContext(b.t.Context(), http.MethodPost,
		settlementsPath+"/"+settlementID+"/payments", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set("Idempotency-Key", key)
	b.attachCookies(req)
	fromApp(req)
	b.visits++
	edgetest.SignFrom(b.t, req, edgetest.Origin{At: b.now.Now(), ClientIP: fmt.Sprintf("198.51.100.%d", b.visits)})
	return b.send(req)
}

func paymentInfoBody(info, qrFileID string, methods ...string) string {
	qrJSON := "null"
	if qrFileID != "" {
		qrJSON = `"` + qrFileID + `"`
	}
	quoted := make([]string, 0, len(methods))
	for _, method := range methods {
		quoted = append(quoted, `"`+method+`"`)
	}
	return fmt.Sprintf(`{"payment_info":%q,"payment_qr_file_id":%s,"payment_methods":[%s],`+
		`"bank_name":"BCA","bank_account_number":"1234567890","bank_account_holder":"Darmi"}`,
		info, qrJSON, strings.Join(quoted, ","))
}

func paymentBody(amount int, method, proofFileID string) string {
	proof := "null"
	if proofFileID != "" {
		proof = `"` + proofFileID + `"`
	}
	return fmt.Sprintf(`{"amount":%d,"method":"%s","paid_on":"2026-10-19","proof_file_id":%s}`, amount, method, proof)
}

type paymentScene struct {
	dropScene
	settlementID string
	countID      string
}

func issuedSettlement(t *testing.T) paymentScene {
	t.Helper()
	scene := recordedDropWith(t, "ON_COUNT")
	counted := scene.store.recordCount(countBody(scene.partnershipID, scene.productID,
		countLine{recorded: 20, counted: 2, returned: 2, reason: "UNSOLD"}))
	settlementID := quotedAfter(string(counted.body), `"settlement":{"id":"`)
	if settlementID == "" {
		t.Fatalf("counted = %s", counted.body)
	}
	return paymentScene{
		dropScene: scene, settlementID: settlementID, countID: quotedAfter(string(counted.body), `"id":"`),
	}
}

func (b *browser) uploadProof() string {
	b.t.Helper()
	uploaded := b.upload(part{name: "purpose", data: []byte("PAYMENT_PROOF")},
		part{name: "file", filename: "image", data: photoWithLocation(b.t)})
	fileID := quotedAfter(string(uploaded.body), `"id":"`)
	if fileID == "" {
		b.t.Fatalf("uploaded = %s", uploaded.body)
	}
	return fileID
}

func TestRecordPayment_BRSTL05_EveryPaymentNeedsProofAnAcceptedMethodAndTheRemainingAmount(t *testing.T) {
	scene := issuedSettlement(t)
	proofID := scene.store.uploadProof()

	withoutProof := scene.store.recordPayment(scene.settlementID, paymentBody(31500, "CASH", ""), paymentKey)
	tooLittle := scene.store.recordPayment(scene.settlementID, paymentBody(31000, "CASH", proofID),
		"0192f0c4-0000-7000-8000-00000000d002")
	notAccepted := scene.store.recordPayment(scene.settlementID, paymentBody(31500, "TRANSFER", proofID),
		"0192f0c4-0000-7000-8000-00000000d003")

	requireCode(t, withoutProof, "VALIDATION_FAILED")
	loadContract(t).requireConforms(t, withoutProof)
	if !strings.Contains(string(withoutProof.body), "Unggah bukti pembayaran berupa foto atau PDF.") {
		t.Fatalf("without proof = %s", withoutProof.body)
	}
	if !strings.Contains(string(tooLittle.body), "Jumlah harus sama dengan sisa Setoran (Rp31.500).") {
		t.Fatalf("too little = %s", tooLittle.body)
	}
	if !strings.Contains(string(notAccepted.body), `"field":"method"`) {
		t.Fatalf("not accepted = %s", notAccepted.body)
	}
}

func TestRecordPayment_PDFReceiptIsKeptAndTheConsignorCanOpenIt(t *testing.T) {
	scene := issuedSettlement(t)
	receipt := []byte("%PDF-1.4\n1 0 obj << /Type /Catalog >> endobj\ntrailer << /Root 1 0 R >>\n%%EOF\n")
	scripted := []byte("%PDF-1.4\n1 0 obj << /S /JavaScript /JS (app.alert(1)) >> endobj\n%%EOF\n")

	uploaded := scene.store.upload(part{name: "purpose", data: []byte("PAYMENT_PROOF")},
		part{name: "file", filename: "bukti.pdf", data: receipt})
	refused := scene.store.upload(part{name: "purpose", data: []byte("PAYMENT_PROOF")},
		part{name: "file", filename: "bukti.pdf", data: scripted})
	proofID := quotedAfter(string(uploaded.body), `"id":"`)
	recorded := scene.store.recordPayment(scene.settlementID, paymentBody(31500, "CASH", proofID), paymentKey)
	opened := scene.consignor.get("/api/v1/files/" + proofID)

	requireCreated(t, uploaded)
	loadContract(t).requireConforms(t, uploaded)
	if !strings.Contains(string(uploaded.body), `"content_type":"application/pdf"`) ||
		!strings.Contains(string(refused.body), `"code":"INVALID_PROOF"`) {
		t.Fatalf("uploaded = %s, refused = %s", uploaded.body, refused.body)
	}
	requireCreated(t, recorded)
	if opened.status != http.StatusOK || opened.header.Get(fiber.HeaderContentType) != "application/pdf" ||
		string(opened.body) != string(receipt) {
		t.Fatalf("opened = %d %s", opened.status, opened.header.Get(fiber.HeaderContentType))
	}
}

func TestConfirmPayment_FRSTL06_TransferWithProofIsAcceptedAndTheSettlementIsPaid(t *testing.T) {
	scene := issuedSettlement(t)
	consignorID := scene.consignor.actor(t).Membership.OrganizationID
	requireCode(t, scene.consignor.patch(organizationsPath+"/"+consignorID+"/payment-info",
		paymentInfoBody("", "", "CASH", "TRANSFER")), "OK")
	proofID := scene.store.uploadProof()

	recorded := scene.store.recordPayment(scene.settlementID, paymentBody(31500, "TRANSFER", proofID), paymentKey)
	replayed := scene.store.recordPayment(scene.settlementID, paymentBody(31500, "TRANSFER", proofID), paymentKey)
	proofForConsignor := scene.consignor.get("/api/v1/files/" + proofID)
	paymentID := quotedAfter(string(recorded.body), `"payments":[{"id":"`)
	confirmed := scene.consignor.post("/api/v1/payments/" + paymentID + "/confirm")
	count := scene.store.get(countsPath + "/" + scene.countID)

	requireCreated(t, recorded)
	loadContract(t).requireConforms(t, recorded)
	if replayed.header.Get("Idempotent-Replay") != "true" || proofForConsignor.status != http.StatusOK {
		t.Fatalf("replayed = %s, proof = %d", replayed.body, proofForConsignor.status)
	}
	requireCode(t, confirmed, "OK")
	loadContract(t).requireConforms(t, confirmed)
	if !strings.Contains(string(confirmed.body), `"status":"PAID"`) ||
		!strings.Contains(string(confirmed.body), `"status":"CONFIRMED"`) ||
		!strings.Contains(string(confirmed.body), `"event_type":"PAYMENT_CONFIRMED"`) {
		t.Fatalf("confirmed = %s", confirmed.body)
	}
	if !strings.Contains(string(count.body), `"status":"CONFIRMED"`) {
		t.Fatalf("BR-STL-08 count = %s", count.body)
	}
	inbox := scene.store.get("/api/v1/notifications")
	if !strings.Contains(string(inbox.body), "Pembayaran Rp31.500 untuk ST-202610-0001 diterima.") {
		t.Fatalf("inbox = %s", inbox.body)
	}
}

func TestRejectPayment_BRSTL06_RejectedCashCanBeRecordedAgain(t *testing.T) {
	scene := issuedSettlement(t)
	proofID := scene.store.uploadProof()
	recorded := scene.store.recordPayment(scene.settlementID, paymentBody(31500, "CASH", proofID), paymentKey)
	paymentID := quotedAfter(string(recorded.body), `"payments":[{"id":"`)

	pending := scene.store.recordPayment(scene.settlementID, paymentBody(31500, "CASH", proofID),
		"0192f0c4-0000-7000-8000-00000000d002")
	rejected := scene.consignor.postJSON("/api/v1/payments/"+paymentID+"/reject", `{"reason":"Uang belum diterima"}`)
	again := scene.store.recordPayment(scene.settlementID, paymentBody(31500, "CASH", proofID),
		"0192f0c4-0000-7000-8000-00000000d003")

	requireCode(t, pending, "PAYMENT_PENDING")
	requireCode(t, rejected, "OK")
	if !strings.Contains(string(rejected.body), `"rejection_reason":"Uang belum diterima"`) ||
		!strings.Contains(string(rejected.body), `"status":"ISSUED"`) {
		t.Fatalf("rejected = %s", rejected.body)
	}
	requireCreated(t, again)
}

func TestMarkPaid_FRSTL07_ConsignorMarksCashReceivedWithoutTheStore(t *testing.T) {
	scene := issuedSettlement(t)
	path := settlementsPath + "/" + scene.settlementID + "/mark-paid"

	byStore := scene.store.post(path)
	byConsignor := scene.consignor.post(path)
	again := scene.consignor.post(path)

	requireCode(t, byStore, "FORBIDDEN")
	requireCode(t, byConsignor, "OK")
	loadContract(t).requireConforms(t, byConsignor)
	if !strings.Contains(string(byConsignor.body), `"status":"PAID"`) ||
		!strings.Contains(string(byConsignor.body), `"method":"CASH"`) {
		t.Fatalf("by consignor = %s", byConsignor.body)
	}
	requireCode(t, again, "SETTLEMENT_ALREADY_PAID")
}

func TestGetSettlement_FRSTL05_StoreSeesThePaymentInfoOfAnActivePartner(t *testing.T) {
	scene := issuedSettlement(t)
	consignorID := scene.consignor.actor(t).Membership.OrganizationID
	saved := scene.consignor.patch(organizationsPath+"/"+consignorID+"/payment-info",
		paymentInfoBody("Transfer sebelum jam 5 sore", "", "CASH", "TRANSFER"))

	got := scene.store.get(settlementsPath + "/" + scene.settlementID)

	requireCode(t, saved, "OK")
	loadContract(t).requireConforms(t, saved)
	if !strings.Contains(string(got.body), `"payee":{"payment_info":"Transfer sebelum jam 5 sore"`) ||
		!strings.Contains(string(got.body), `"methods":["CASH","TRANSFER"],"bank_name":"BCA",`+
			`"bank_account_number":"1234567890","bank_account_holder":"Darmi"`) {
		t.Fatalf("got = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}
