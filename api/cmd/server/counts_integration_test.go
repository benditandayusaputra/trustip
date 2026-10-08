package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/claimed"
	countdomain "github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
	countservice "github.com/benditandayusaputra/trustip/api/internal/modules/count/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
)

const (
	countsPath = "/api/v1/counts"
	countKey   = "0192f0c4-0000-7000-8000-00000000c001"
)

func (b *browser) recordCount(body string) exchange {
	b.t.Helper()
	return b.recordCountWithKey(body, countKey)
}

func (b *browser) recordCountWithKey(body, key string) exchange {
	b.t.Helper()
	req := httptest.NewRequestWithContext(b.t.Context(), http.MethodPost, countsPath, strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set("Idempotency-Key", key)
	b.attachCookies(req)
	fromApp(req)
	b.visits++
	edgetest.SignFrom(b.t, req, edgetest.Origin{At: b.now.Now(), ClientIP: fmt.Sprintf("198.51.100.%d", b.visits)})
	return b.send(req)
}

type countLine struct {
	recorded, counted, returned int
	reason                      string
}

func countBody(partnershipID, productID string, line countLine) string {
	reason := "null"
	if line.reason != "" {
		reason = `"` + line.reason + `"`
	}
	return fmt.Sprintf(`{"partnership_id":"%s","agreement_version":1,"occasion":"STORE_CLOSING","lines":[`+
		`{"product_id":"%s","recorded_qty":%d,"counted_qty":%d,"returned_qty":%d,"return_reason":%s}]}`,
		partnershipID, productID, line.recorded, line.counted, line.returned, reason)
}

func TestRecordCount_FRCNT05_TwentyLoavesTwoLeftAndReturnedGiveTheSectionTwelveRecap(t *testing.T) {
	scene := recordedDrop(t)

	counted := scene.store.recordCount(countBody(scene.partnershipID, scene.productID,
		countLine{recorded: 20, counted: 2, returned: 2, reason: "UNSOLD"}))

	requireCreated(t, counted)
	loadContract(t).requireConforms(t, counted)
	body := string(counted.body)
	if !strings.Contains(body, `"number":"HS-202610-0001"`) || !strings.Contains(body, `"sold_qty":18`) ||
		!strings.Contains(body, `"split":{"gross":36000,"store_share":4500,"consignor_share":31500}`) ||
		!strings.Contains(body, `"kept_qty":0`) || !strings.Contains(body, `"is_valid":true`) {
		t.Fatalf("counted = %s", body)
	}
	if got := productStock(t, scene.invitationScene, scene.partnershipID); got != 0 {
		t.Fatalf("stock = %d, want 0", got)
	}
	inbox := scene.consignor.get("/api/v1/notifications")
	if !strings.Contains(string(inbox.body), "menghitung sisa: 18 terjual, 2 dikembalikan (HS-202610-0001)") {
		t.Fatalf("inbox = %s", inbox.body)
	}
	requireCode(t, scene.store.post(scene.path+"/cancel"), "DROP_ALREADY_COUNTED")
}

func TestRecordCount_CNT_ConcurrentCountsOnOnePartnerOnlyOnePasses(t *testing.T) {
	scene := recordedDrop(t)
	deps := buildTradeComponents(scene.store.pool, scene.store.now, testSealKeys{}).counts
	cashier := scene.store.actor(t)
	errs := make([]error, 2)
	var group sync.WaitGroup

	for index, key := range []string{countKey, "0192f0c4-0000-7000-8000-00000000c002"} {
		group.Go(func() {
			_, errs[index] = countservice.NewRecordCount(deps).Execute(t.Context(), cashier,
				countservice.RecordCountCommand{IdempotencyKey: key, Input: countdomain.RecordInput{
					PartnershipID: scene.partnershipID, AgreementVersion: 1, CountedAt: claimed.At(testNow()),
					Occasion: "STORE_CLOSING", Lines: []countdomain.RequestLine{
						{ProductID: scene.productID, RecordedQty: 20, CountedQty: 12},
					},
				}})
		})
	}
	group.Wait()

	passed, refused := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			passed++
		case apperror.From(err).Code() == apperror.CodeStockChanged:
			refused++
		}
	}
	if passed != 1 || refused != 1 {
		t.Fatalf("errs = %v", errs)
	}
	if got := productStock(t, scene.invitationScene, scene.partnershipID); got != 12 {
		t.Fatalf("stock = %d, want 12", got)
	}
}

func TestRecordCount_BRCNT07_SavingTwiceMakesOneCount(t *testing.T) {
	scene := recordedDrop(t)
	body := countBody(scene.partnershipID, scene.productID, countLine{recorded: 20, counted: 12})

	first := scene.store.recordCount(body)
	second := scene.store.recordCount(body)

	requireCreated(t, first)
	requireCreated(t, second)
	if second.header.Get("Idempotent-Replay") != "true" ||
		quotedAfter(string(first.body), `"id":"`) != quotedAfter(string(second.body), `"id":"`) {
		t.Fatalf("first = %s\nsecond = %s", first.body, second.body)
	}
	if got := productStock(t, scene.invitationScene, scene.partnershipID); got != 12 {
		t.Fatalf("stock = %d, want 12", got)
	}
}

func TestCountSheet_FRCNT01_StoreSeesTheRecordedStockAndTheConsignorCannotCount(t *testing.T) {
	scene := recordedDrop(t)
	path := "/api/v1/partnerships/" + scene.partnershipID + "/count-sheet"

	fromStore := scene.store.get(path)
	fromConsignor := scene.consignor.get(path)

	requireCode(t, fromStore, "OK")
	loadContract(t).requireConforms(t, fromStore)
	if !strings.Contains(string(fromStore.body), `"recorded_qty":20`) ||
		!strings.Contains(string(fromStore.body), `"deposit_price":1750`) {
		t.Fatalf("sheet = %s", fromStore.body)
	}
	requireCode(t, fromConsignor, "FORBIDDEN")
}

func TestRecordCount_BRCNT01_CountAboveTheStockNamesTheRecordedStock(t *testing.T) {
	scene := recordedDrop(t)

	got := scene.store.recordCount(countBody(scene.partnershipID, scene.productID,
		countLine{recorded: 20, counted: 21}))

	requireCode(t, got, "VALIDATION_FAILED")
	loadContract(t).requireConforms(t, got)
	if !strings.Contains(string(got.body), "Sisa melebihi stok tercatat (20).") {
		t.Fatalf("got = %s", got.body)
	}
}

func TestGetCount_BothPartiesSeeTheSameSealedRecap(t *testing.T) {
	scene := recordedDrop(t)
	counted := scene.store.recordCount(countBody(scene.partnershipID, scene.productID,
		countLine{recorded: 20, counted: 2, returned: 2, reason: "UNSOLD"}))
	path := countsPath + "/" + quotedAfter(string(counted.body), `"id":"`)

	fromStore := scene.store.get(path)
	fromConsignor := scene.consignor.get(path)

	requireCode(t, fromConsignor, "OK")
	loadContract(t).requireConforms(t, fromConsignor)
	if stripRequestID(fromStore.body) != stripRequestID(fromConsignor.body) ||
		!strings.Contains(string(fromConsignor.body), `"is_valid":true`) {
		t.Fatalf("store = %s\nconsignor = %s", fromStore.body, fromConsignor.body)
	}
}

func TestConfirmCount_FRCNT08_ConsignorConfirmsOnceAndBothSeeTheHistory(t *testing.T) {
	scene := recordedDrop(t)
	counted := scene.store.recordCount(countBody(scene.partnershipID, scene.productID,
		countLine{recorded: 20, counted: 12}))
	path := countsPath + "/" + quotedAfter(string(counted.body), `"id":"`)

	confirmed := scene.consignor.post(path + "/confirm")
	again := scene.consignor.post(path + "/confirm")
	history := scene.store.get(countsPath + "?partnership_id=" + scene.partnershipID)

	requireCode(t, confirmed, "OK")
	loadContract(t).requireConforms(t, confirmed)
	if !strings.Contains(string(confirmed.body), `"status":"CONFIRMED"`) ||
		!strings.Contains(string(confirmed.body), `"event_type":"COUNT_CONFIRMED"`) ||
		!strings.Contains(string(confirmed.body), `"is_valid":true`) {
		t.Fatalf("confirmed = %s", confirmed.body)
	}
	requireCode(t, again, "DOCUMENT_NOT_PENDING")
	requireCode(t, history, "OK")
	loadContract(t).requireConforms(t, history)
	if !strings.Contains(string(history.body), `"sold_qty":8`) ||
		!strings.Contains(string(history.body), `"consignor_share":14000`) {
		t.Fatalf("history = %s", history.body)
	}
}

func TestGetCount_FRCNT08_UnansweredCountReadsAsConfirmedAfterFortyEightHours(t *testing.T) {
	scene := recordedDrop(t)
	counted := scene.store.recordCount(countBody(scene.partnershipID, scene.productID,
		countLine{recorded: 20, counted: 12}))
	scene.consignor.now.Advance(48 * time.Hour)
	scene.consignor.post("/api/v1/sessions/refresh")

	got := scene.consignor.get(countsPath + "/" + quotedAfter(string(counted.body), `"id":"`))

	if !strings.Contains(string(got.body), `"status":"CONFIRMED"`) ||
		!strings.Contains(string(got.body), `"is_automatic":true`) {
		t.Fatalf("got = %s", got.body)
	}
}
