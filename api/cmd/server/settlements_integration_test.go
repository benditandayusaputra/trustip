package main

import (
	"strings"
	"testing"
)

const settlementsPath = "/api/v1/settlements"

func TestRecordCount_FRSTL01_OnCountPartnershipIssuesTheSettlementWithTheCount(t *testing.T) {
	scene := recordedDropWith(t, "ON_COUNT")

	counted := scene.store.recordCount(countBody(scene.partnershipID, scene.productID,
		countLine{recorded: 20, counted: 2, returned: 2, reason: "UNSOLD"}))

	requireCreated(t, counted)
	loadContract(t).requireConforms(t, counted)
	settlementID := quotedAfter(string(counted.body), `"settlement":{"id":"`)
	path := settlementsPath + "/" + settlementID
	fromStore := scene.store.get(path)
	fromConsignor := scene.consignor.get(path)
	requireCode(t, fromConsignor, "OK")
	loadContract(t).requireConforms(t, fromConsignor)
	body := string(fromConsignor.body)
	if stripRequestID(fromStore.body) != stripRequestID(fromConsignor.body) ||
		!strings.Contains(body, `"number":"ST-202610-0001"`) || !strings.Contains(body, `"status":"ISSUED"`) ||
		!strings.Contains(body, `"split":{"gross":36000,"store_share":4500,"consignor_share":31500}`) ||
		!strings.Contains(body, `"due_date":"2026-10-26"`) || !strings.Contains(body, `"number":"HS-202610-0001"`) ||
		!strings.Contains(body, `"event_type":"SETTLEMENT_ISSUED"`) || !strings.Contains(body, `"is_valid":true`) {
		t.Fatalf("store = %s\nconsignor = %s", fromStore.body, body)
	}
	inbox := scene.store.get("/api/v1/notifications")
	if !strings.Contains(string(inbox.body), "Setoran ST-202610-0001 dari Warung Bu Tini terbit") {
		t.Fatalf("inbox = %s", inbox.body)
	}
}

func TestRecordCount_FRSTL02_WeeklyPartnershipKeepsTheSaleRunning(t *testing.T) {
	scene := recordedDrop(t)

	counted := scene.store.recordCount(countBody(scene.partnershipID, scene.productID,
		countLine{recorded: 20, counted: 2, returned: 2, reason: "UNSOLD"}))
	running := scene.consignor.get("/api/v1/partnerships/" + scene.partnershipID + "/running-sales")
	listed := scene.store.get(settlementsPath + "?partnership_id=" + scene.partnershipID)

	requireCreated(t, counted)
	if !strings.Contains(string(counted.body), `"settlement":null`) {
		t.Fatalf("counted = %s", counted.body)
	}
	requireCode(t, running, "OK")
	loadContract(t).requireConforms(t, running)
	if !strings.Contains(string(running.body), `"has_sales":true`) ||
		!strings.Contains(string(running.body), `"sold_qty":18`) ||
		!strings.Contains(string(running.body), `"consignor_share":31500`) {
		t.Fatalf("running = %s", running.body)
	}
	requireCode(t, listed, "OK")
	if !strings.Contains(string(listed.body), `"data":[]`) {
		t.Fatalf("listed = %s", listed.body)
	}
}

func TestListSettlements_FRSTL09_FiltersByStatusAndAge(t *testing.T) {
	scene := recordedDropWith(t, "ON_COUNT")
	scene.store.recordCount(countBody(scene.partnershipID, scene.productID, countLine{recorded: 20, counted: 12}))

	issued := scene.consignor.get(settlementsPath + "?status=ISSUED&age=NOT_DUE")
	late := scene.consignor.get(settlementsPath + "?age=LATE_1_7")
	unknown := scene.consignor.get(settlementsPath + "?age=ANCIENT")

	requireCode(t, issued, "OK")
	loadContract(t).requireConforms(t, issued)
	if !strings.Contains(string(issued.body), `"consignor_share":14000`) ||
		!strings.Contains(string(issued.body), `"age":"NOT_DUE"`) ||
		!strings.Contains(string(issued.body), `"name":"Warung Bu Tini"`) {
		t.Fatalf("issued = %s", issued.body)
	}
	if !strings.Contains(string(late.body), `"data":[]`) {
		t.Fatalf("late = %s", late.body)
	}
	requireCode(t, unknown, "VALIDATION_FAILED")
}

func TestRecordCount_BRCNT06_CountInsideAnIssuedSettlementIsRefused(t *testing.T) {
	scene := recordedDropWith(t, "ON_COUNT")
	scene.store.recordCount(countBody(scene.partnershipID, scene.productID, countLine{recorded: 20, counted: 12}))

	again := scene.store.recordCountWithKey(countBody(scene.partnershipID, scene.productID,
		countLine{recorded: 12, counted: 10}), "0192f0c4-0000-7000-8000-00000000c003")

	requireCode(t, again, "VALIDATION_FAILED")
	if !strings.Contains(string(again.body), `"code":"PERIOD_ALREADY_SETTLED"`) ||
		!strings.Contains(string(again.body), "Tanggal ini sudah masuk Setoran ST-202610-0001.") {
		t.Fatalf("again = %s", again.body)
	}
}
