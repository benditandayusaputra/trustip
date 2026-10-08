package main

import (
	"strings"
	"testing"
)

func TestStock_FRSTK01_RecordedDropShowsOneBatchToBothParties(t *testing.T) {
	scene := recordedDrop(t)
	path := "/api/v1/partnerships/" + scene.partnershipID + "/stock"

	fromStore := scene.store.get(path)
	fromConsignor := scene.consignor.get(path)

	requireCode(t, fromStore, "OK")
	loadContract(t).requireConforms(t, fromStore)
	if stripRequestID(fromStore.body) != stripRequestID(fromConsignor.body) ||
		!strings.Contains(string(fromStore.body), `"balance":20,"flag":"SAFE","nearest_expiry":"2026-10-22"`) ||
		!strings.Contains(string(fromStore.body), `"batch_code":"TM-202610-0001/1"`) ||
		!strings.Contains(string(fromStore.body), `"age_days":0`) {
		t.Fatalf("store = %s\nconsignor = %s", fromStore.body, fromConsignor.body)
	}
}

func TestStock_FRSTK02_FRSTK03_EachSideSeesItsCombinedStock(t *testing.T) {
	scene := recordedDrop(t)

	fromStore := scene.store.get("/api/v1/stock")
	fromConsignor := scene.consignor.get("/api/v1/stock")

	contract := loadContract(t)
	for _, got := range []exchange{fromStore, fromConsignor} {
		requireCode(t, got, "OK")
		contract.requireConforms(t, got)
		if !strings.Contains(string(got.body), `"balance":20`) {
			t.Fatalf("got = %s", got.body)
		}
	}
	if !strings.Contains(string(fromStore.body), `"type":"CONSIGNOR"`) ||
		!strings.Contains(string(fromConsignor.body), `"type":"STORE"`) {
		t.Fatalf("store = %s\nconsignor = %s", fromStore.body, fromConsignor.body)
	}
}

func TestStock_FRSTK04_CardLinksEveryMovementToItsDocument(t *testing.T) {
	scene := recordedDrop(t)
	scene.store.post(scene.path + "/cancel")
	dropID := strings.TrimPrefix(scene.path, dropsPath+"/")

	card := scene.consignor.get("/api/v1/partnerships/" + scene.partnershipID + "/stock/" + scene.productID + "/card")

	requireCode(t, card, "OK")
	loadContract(t).requireConforms(t, card)
	body := string(card.body)
	if strings.Count(body, `"source_id":"`+dropID+`"`) != 2 ||
		!strings.Contains(body, `"type":"ADJUSTMENT","qty":-20`) || !strings.Contains(body, `"balance_after":0`) ||
		!strings.Contains(body, `"type":"RECEIVE","qty":20`) || !strings.Contains(body, `"balance_after":20`) {
		t.Fatalf("card = %s", body)
	}
	stock := scene.store.get("/api/v1/partnerships/" + scene.partnershipID + "/stock")
	if !strings.Contains(string(stock.body), `"data":[]`) {
		t.Fatalf("stock = %s", stock.body)
	}
}
