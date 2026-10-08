package main

import (
	"strings"
	"testing"
	"time"
)

const week = 7 * 24 * time.Hour

func (s dropScene) passWeek() {
	s.store.now.Advance(week)
	s.store.post("/api/v1/sessions/refresh")
	s.consignor.post("/api/v1/sessions/refresh")
}

func TestDropSuggestions_INS09_TwoSameWeekdaysOfSeventeenSuggestNineteen(t *testing.T) {
	scene := recordedDrop(t)
	path := "/api/v1/partnerships/" + scene.partnershipID + "/drop-suggestions"
	before := scene.store.get(path)
	requireCreated(t, scene.store.recordCountWithKey(countBody(scene.partnershipID, scene.productID,
		countLine{recorded: 20, counted: 3}), "0192f0c4-0000-7000-8000-00000000c101"))
	scene.passWeek()
	requireCreated(t, scene.store.recordDrop(dropBody(scene.partnershipID, scene.productID, 20),
		"0192f0c4-0000-7000-8000-00000000d102"))
	requireCreated(t, scene.store.recordCountWithKey(countBody(scene.partnershipID, scene.productID,
		countLine{recorded: 23, counted: 6}), "0192f0c4-0000-7000-8000-00000000c102"))
	scene.passWeek()

	fromStore := scene.store.get(path)
	fromConsignor := scene.consignor.get(path)

	requireCode(t, before, "OK")
	if !strings.Contains(string(before.body), `"suggested_qty":0,"has_enough_data":false`) {
		t.Fatalf("before = %s", before.body)
	}
	requireCode(t, fromStore, "OK")
	loadContract(t).requireConforms(t, fromStore)
	want := `{"product_id":"` + scene.productID + `","suggested_qty":19,"has_enough_data":true}`
	if !strings.Contains(string(fromStore.body), want) ||
		stripRequestID(fromStore.body) != stripRequestID(fromConsignor.body) {
		t.Fatalf("store = %s\nconsignor = %s", fromStore.body, fromConsignor.body)
	}
}

func TestDashboard_DSH_RecordedDropShowsOnBothDashboards(t *testing.T) {
	scene := recordedDrop(t)

	fromConsignor := scene.consignor.get("/api/v1/dashboard")
	fromStore := scene.store.get("/api/v1/dashboard")

	contract := loadContract(t)
	for _, got := range []exchange{fromConsignor, fromStore} {
		requireCode(t, got, "OK")
		contract.requireConforms(t, got)
		if !strings.Contains(string(got.body), `"active_partners":1,"stock_units":20`) {
			t.Fatalf("got = %s", got.body)
		}
	}
	if !strings.Contains(string(fromConsignor.body), `"drops_to_confirm":1`) ||
		!strings.Contains(string(fromConsignor.body), `"quick_count":[]`) {
		t.Fatalf("consignor = %s", fromConsignor.body)
	}
	uncounted := `"quick_count":[{"partnership_id":"` + scene.partnershipID + `"`
	if !strings.Contains(string(fromStore.body), uncounted) ||
		!strings.Contains(string(fromStore.body), `"drops_to_confirm":0`) {
		t.Fatalf("store = %s", fromStore.body)
	}
}

func TestDashboard_FRCNT11_CountedPartnerLeavesQuickCount(t *testing.T) {
	scene := recordedDrop(t)
	requireCreated(t, scene.store.recordCount(countBody(scene.partnershipID, scene.productID,
		countLine{recorded: 20, counted: 5})))

	got := scene.store.get("/api/v1/dashboard")

	requireCode(t, got, "OK")
	if !strings.Contains(string(got.body), `"quick_count":[]`) {
		t.Fatalf("got = %s", got.body)
	}
}

func TestInsights_INS01_INS07_ExpiredUncountedStockIsFlaggedForEachSide(t *testing.T) {
	scene := recordedDrop(t)
	scene.store.now.Advance(4 * 24 * time.Hour)
	scene.store.post("/api/v1/sessions/refresh")
	scene.consignor.post("/api/v1/sessions/refresh")

	fromStore := scene.store.get("/api/v1/insights")
	fromConsignor := scene.consignor.get("/api/v1/insights")

	requireCode(t, fromStore, "OK")
	loadContract(t).requireConforms(t, fromStore)
	store := string(fromStore.body)
	expired := `"kind":"EXPIRED","level":"HIGH","action":"COUNT_WITH_RETURNS"`
	figures := `"units":20,"days":0,"amount":40000`
	if !strings.Contains(store, `"data":[{`+expired) || !strings.Contains(store, figures) ||
		!strings.Contains(store, `"kind":"NOT_COUNTED","level":"LOW","action":"COUNT"`) {
		t.Fatalf("store = %s", store)
	}
	consignor := string(fromConsignor.body)
	if !strings.Contains(consignor, `"kind":"EXPIRED","level":"HIGH","action":"REMIND_STORE"`) ||
		!strings.Contains(consignor, `"action":"REMIND_COUNT"`) {
		t.Fatalf("consignor = %s", consignor)
	}
}

func TestInsights_INS10_ListedStoreBelowSeventyFiveIsNotSuggested(t *testing.T) {
	consignor, store := listedStoreScene(t)

	listed := consignor.get(directoryStoresPath)
	got := consignor.get("/api/v1/insights")

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
	listedBody := string(listed.body)
	if !strings.Contains(listedBody, `"id":"`+store.id+`"`) || strings.Contains(listedBody, `"score":7`) ||
		strings.Contains(string(got.body), `"kind":"MATCHING_PARTNER"`) {
		t.Fatalf("listed = %s\ninsights = %s", listed.body, got.body)
	}
}

func TestDashboard_BRNED03_StoreSeesNeedsAboutToExpire(t *testing.T) {
	store := storeWithNeeds(t)
	requireCreated(t, store.postJSON(needsPath, needBody("Butuh roti untuk sarapan")))
	early := store.get("/api/v1/dashboard")
	store.now.Advance(27*24*time.Hour + time.Hour)
	store.post("/api/v1/sessions/refresh")

	late := store.get("/api/v1/dashboard")

	requireCode(t, late, "OK")
	loadContract(t).requireConforms(t, late)
	if !strings.Contains(string(early.body), `"needs_expiring":0`) ||
		!strings.Contains(string(late.body), `"needs_expiring":1`) {
		t.Fatalf("early = %s\nlate = %s", early.body, late.body)
	}
}
