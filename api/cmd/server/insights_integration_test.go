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
