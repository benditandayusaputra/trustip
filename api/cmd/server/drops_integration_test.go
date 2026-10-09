package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/claimed"
	dropdomain "github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
	dropservice "github.com/benditandayusaputra/trustip/api/internal/modules/drop/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
)

const (
	dropsPath = "/api/v1/drops"
	dropKey   = "0192f0c4-0000-7000-8000-00000000d001"
)

func (b *browser) recordDrop(body, key string) exchange {
	b.t.Helper()
	req := httptest.NewRequestWithContext(b.t.Context(), http.MethodPost, dropsPath, strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set("Idempotency-Key", key)
	b.attachCookies(req)
	fromApp(req)
	b.visits++
	edgetest.SignFrom(b.t, req, edgetest.Origin{At: b.now.Now(), ClientIP: fmt.Sprintf("198.51.100.%d", b.visits)})
	return b.send(req)
}

func dropBody(partnershipID, productID string, qty int) string {
	return fmt.Sprintf(`{"partnership_id":"%s","agreement_version":1,"items":[{"product_id":"%s","qty":%d}]}`,
		partnershipID, productID, qty)
}

func productStock(t *testing.T, scene invitationScene, partnershipID string) int64 {
	t.Helper()
	balance, err := sqlc.New(scene.store.pool.Querier(t.Context())).SumProductStock(t.Context(),
		sqlc.SumProductStockParams{PartnershipID: partnershipID, ProductID: scene.productID})
	if err != nil {
		t.Fatalf("stock: %v", err)
	}
	return balance
}

func TestRecordDrop_FRDRP01_TwentyLoavesRaiseTheStockAndTellTheConsignor(t *testing.T) {
	scene, partnershipID := acceptedScene(t)

	recorded := scene.store.recordDrop(dropBody(partnershipID, scene.productID, 20), dropKey)

	requireCreated(t, recorded)
	loadContract(t).requireConforms(t, recorded)
	body := string(recorded.body)
	if !strings.Contains(body, `"number":"TM-202610-0001"`) || !strings.Contains(body, `"total_qty":20`) ||
		!strings.Contains(body, `"batch_code":"TM-202610-0001/1"`) ||
		!strings.Contains(body, `"expiry_date":"2026-10-22"`) ||
		!strings.Contains(body, `"is_valid":true`) || !strings.Contains(body, `"event_type":"DROP_RECORDED"`) {
		t.Fatalf("recorded = %s", body)
	}
	if got := productStock(t, scene, partnershipID); got != 20 {
		t.Fatalf("stock = %d, want 20", got)
	}
	inbox := scene.consignor.get("/api/v1/notifications")
	if !strings.Contains(string(inbox.body), "mencatat 20 unit titipan masuk (TM-202610-0001)") {
		t.Fatalf("inbox = %s", inbox.body)
	}
}

func requireCreated(t *testing.T, got exchange) {
	t.Helper()
	if got.status != http.StatusCreated {
		t.Fatalf("status = %d %s", got.status, got.body)
	}
}

func TestRecordDrop_BRDRP07_SavingTwiceMakesOneDocument(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	body := dropBody(partnershipID, scene.productID, 20)

	first := scene.store.recordDrop(body, dropKey)
	second := scene.store.recordDrop(body, dropKey)
	changed := scene.store.recordDrop(dropBody(partnershipID, scene.productID, 21), dropKey)

	requireCreated(t, first)
	requireCreated(t, second)
	if second.header.Get("Idempotent-Replay") != "true" ||
		quotedAfter(string(first.body), `"id":"`) != quotedAfter(string(second.body), `"id":"`) {
		t.Fatalf("first = %s\nsecond = %s", first.body, second.body)
	}
	requireCode(t, changed, "IDEMPOTENCY_KEY_REUSED")
	loadContract(t).requireConforms(t, second)
	if got := productStock(t, scene, partnershipID); got != 20 {
		t.Fatalf("stock = %d, want 20", got)
	}
}

func TestRecordDrop_BRDRP07_ConcurrentSavesWithOneKeyMakeOneDocument(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	deps := buildTradeComponents(testFoundation(scene.store.pool, scene.store.now), testSealKeys{}).drops
	cashier := scene.store.actor(t)
	command := dropservice.RecordDropCommand{IdempotencyKey: dropKey, Input: dropdomain.RecordInput{
		PartnershipID: partnershipID, AgreementVersion: 1, OccurredAt: claimed.At(testNow()),
		Items: []dropdomain.ItemInput{{ProductID: scene.productID, Qty: 20}},
	}}
	receipts := make([]dropservice.DropReceipt, 2)
	errs := make([]error, 2)
	var group sync.WaitGroup

	for index := range receipts {
		group.Go(func() {
			receipts[index], errs[index] = dropservice.NewRecordDrop(deps).Execute(t.Context(), cashier, command)
		})
	}
	group.Wait()

	if errors.Join(errs...) != nil || receipts[0].View.ID != receipts[1].View.ID ||
		receipts[0].IsReplay == receipts[1].IsReplay {
		t.Fatalf("receipts = %+v, errs = %v", receipts, errs)
	}
	if got := productStock(t, scene, partnershipID); got != 20 {
		t.Fatalf("stock = %d, want 20", got)
	}
}

func TestGetDrop_FRDRP06_BothPartiesSeeTheSameSealedDrop(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	recorded := scene.store.recordDrop(dropBody(partnershipID, scene.productID, 20), dropKey)
	path := dropsPath + "/" + quotedAfter(string(recorded.body), `"id":"`)

	fromStore := scene.store.get(path)
	fromConsignor := scene.consignor.get(path)

	requireCode(t, fromStore, "OK")
	loadContract(t).requireConforms(t, fromConsignor)
	if stripRequestID(fromStore.body) != stripRequestID(fromConsignor.body) ||
		!strings.Contains(string(fromConsignor.body),
			`"value_if_sold":{"gross":40000,"store_share":5000,"consignor_share":35000}`) {
		t.Fatalf("store = %s\nconsignor = %s", fromStore.body, fromConsignor.body)
	}
}

func TestRecordDrop_RSEC36_ConsignorCannotRecordADrop(t *testing.T) {
	scene, partnershipID := acceptedScene(t)

	got := scene.consignor.recordDrop(dropBody(partnershipID, scene.productID, 20), dropKey)

	requireCode(t, got, "FORBIDDEN")
	loadContract(t).requireConforms(t, got)
}

func TestRecordDrop_BRDRP01_ProductOutsideTheAgreementIsAFieldError(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	other := scene.consignor.postJSON(productsPath, productBody("LAIN-1"))

	got := scene.store.recordDrop(
		dropBody(partnershipID, quotedAfter(string(other.body), `"id":"`), 5), dropKey)

	requireCode(t, got, "VALIDATION_FAILED")
	if !strings.Contains(string(got.body), `"field":"items[0].product_id","code":"PRODUCT_NOT_IN_AGREEMENT"`) {
		t.Fatalf("got = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestRecordDrop_RRES06_MissingIdempotencyKeyIsMalformed(t *testing.T) {
	scene, partnershipID := acceptedScene(t)

	got := scene.store.recordDrop(dropBody(partnershipID, scene.productID, 20), "")

	requireCode(t, got, "MALFORMED_REQUEST")
	loadContract(t).requireConforms(t, got)
}

type dropScene struct {
	invitationScene
	partnershipID string
	path          string
}

func recordedDrop(t *testing.T) dropScene {
	t.Helper()
	return recordedDropWith(t, "WEEKLY")
}

func recordedDropWith(t *testing.T, schedule string) dropScene {
	t.Helper()
	scene, partnershipID := acceptedSceneWith(t, schedule)
	recorded := scene.store.recordDrop(dropBody(partnershipID, scene.productID, 20), dropKey)
	requireCreated(t, recorded)
	return dropScene{
		invitationScene: scene, partnershipID: partnershipID,
		path: dropsPath + "/" + quotedAfter(string(recorded.body), `"id":"`),
	}
}

func TestConfirmDrop_FRDRP06_ConsignorConfirmsAndTheRecorderIsTold(t *testing.T) {
	scene := recordedDrop(t)
	path := scene.path

	confirmed := scene.consignor.post(path + "/confirm")
	again := scene.consignor.post(path + "/confirm")

	requireCode(t, confirmed, "OK")
	loadContract(t).requireConforms(t, confirmed)
	if !strings.Contains(string(confirmed.body), `"status":"CONFIRMED"`) ||
		!strings.Contains(string(confirmed.body), `"event_type":"DROP_CONFIRMED"`) ||
		!strings.Contains(string(confirmed.body), `"is_valid":true`) {
		t.Fatalf("confirmed = %s", confirmed.body)
	}
	requireCode(t, again, "DOCUMENT_NOT_PENDING")
	inbox := scene.store.get("/api/v1/notifications")
	if !strings.Contains(string(inbox.body), "menyatakan TM-202610-0001 sesuai.") {
		t.Fatalf("inbox = %s", inbox.body)
	}
}

func TestCancelDrop_FRDRP10_StoreCancelsWithinFifteenMinutesAndTheStockGoesBack(t *testing.T) {
	scene := recordedDrop(t)
	partnershipID, path := scene.partnershipID, scene.path

	cancelled := scene.store.post(path + "/cancel")

	requireCode(t, cancelled, "OK")
	loadContract(t).requireConforms(t, cancelled)
	if !strings.Contains(string(cancelled.body), `"status":"CANCELLED"`) ||
		!strings.Contains(string(cancelled.body), `"event_type":"DROP_CANCELLED"`) {
		t.Fatalf("cancelled = %s", cancelled.body)
	}
	if got := productStock(t, scene.invitationScene, partnershipID); got != 0 {
		t.Fatalf("stock = %d, want 0", got)
	}
	requireCode(t, scene.consignor.post(path+"/confirm"), "DOCUMENT_NOT_PENDING")
}

func TestCancelDrop_FRDRP10_SixteenMinutesLaterIsTooLate(t *testing.T) {
	scene := recordedDrop(t)
	path := scene.path
	scene.store.now.Advance(16 * time.Minute)
	scene.store.post("/api/v1/sessions/refresh")

	got := scene.store.post(path + "/cancel")

	requireCode(t, got, "CANCEL_WINDOW_PASSED")
	loadContract(t).requireConforms(t, got)
}

func TestGetDrop_BRDRP08_UnansweredDropReadsAsConfirmedAfterFortyEightHours(t *testing.T) {
	scene := recordedDrop(t)
	path := scene.path
	scene.consignor.now.Advance(48 * time.Hour)
	scene.consignor.post("/api/v1/sessions/refresh")

	got := scene.consignor.get(path)

	requireCode(t, got, "OK")
	if !strings.Contains(string(got.body), `"status":"CONFIRMED"`) ||
		!strings.Contains(string(got.body), `"is_automatic":true`) {
		t.Fatalf("got = %s", got.body)
	}
}

func TestListDrops_FRDRP08_HistoryFiltersByStatusAndPartner(t *testing.T) {
	scene := recordedDrop(t)
	second := scene.store.recordDrop(dropBody(scene.partnershipID, scene.productID, 5),
		"0192f0c4-0000-7000-8000-00000000d002")
	requireCreated(t, second)
	scene.consignor.post(scene.path + "/confirm")

	all := scene.store.get(dropsPath + "?partnership_id=" + scene.partnershipID)
	pending := scene.consignor.get(dropsPath + "?status=RECORDED")
	ranged := scene.store.get(dropsPath + "?from=2026-10-20&to=2026-10-31")

	contract := loadContract(t)
	for _, got := range []exchange{all, pending, ranged} {
		requireCode(t, got, "OK")
		contract.requireConforms(t, got)
	}
	if strings.Count(string(all.body), `"number":"TM-`) != 2 ||
		strings.Count(string(pending.body), `"number":"TM-`) != 1 ||
		!strings.Contains(string(pending.body), `"total_qty":5`) ||
		strings.Contains(string(ranged.body), `"number":"TM-`) {
		t.Fatalf("all = %s\npending = %s\nranged = %s", all.body, pending.body, ranged.body)
	}
}
