package main

import (
	"strings"
	"testing"
)

func TestEndPartnership_BRPRT02_BRPRT03_SalesAreSettledAndTheBlockersListed(t *testing.T) {
	scene := recordedDrop(t)
	scene.store.recordCount(countBody(scene.partnershipID, scene.productID, countLine{recorded: 20, counted: 12}))
	path := "/api/v1/partnerships/" + scene.partnershipID

	refused := scene.store.postJSON(path+"/end", `{"reason":"Toko tutup"}`)
	blockers := scene.store.get(path + "/blockers")
	listed := scene.consignor.get("/api/v1/partnerships")

	requireCode(t, refused, "PARTNERSHIP_HAS_BLOCKERS")
	loadContract(t).requireConforms(t, refused)
	if !strings.Contains(string(refused.body), "stok 12 unit masih di toko") ||
		!strings.Contains(string(refused.body), "Setoran ST-202610-0001 belum lunas") {
		t.Fatalf("refused = %s", refused.body)
	}
	requireCode(t, blockers, "OK")
	loadContract(t).requireConforms(t, blockers)
	if !strings.Contains(string(blockers.body), `"kind":"PENDING_COUNT"`) ||
		!strings.Contains(string(blockers.body), `"kind":"PENDING_DROP"`) {
		t.Fatalf("blockers = %s", blockers.body)
	}
	loadContract(t).requireConforms(t, listed)
	if !strings.Contains(string(listed.body), `"stock_qty":12,"unpaid_amount":14000`) {
		t.Fatalf("FR-PRT-01 listed = %s", listed.body)
	}
}

func TestEndPartnership_FRPRT04_PartnershipWithoutBlockersEndsAndStaysReadable(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	path := "/api/v1/partnerships/" + partnershipID

	ended := scene.consignor.postJSON(path+"/end", `{"reason":"Berhenti produksi"}`)
	again := scene.consignor.postJSON(path+"/end", `{"reason":""}`)
	fromStore := scene.store.get(path)

	requireCode(t, ended, "OK")
	loadContract(t).requireConforms(t, ended)
	if !strings.Contains(string(ended.body), `"status":"ENDED"`) ||
		!strings.Contains(string(ended.body), `"end_reason":"Berhenti produksi"`) {
		t.Fatalf("ended = %s", ended.body)
	}
	requireCode(t, again, "PARTNERSHIP_NOT_ACTIVE")
	if !strings.Contains(string(fromStore.body), `"status":"ENDED"`) {
		t.Fatalf("FR-PRT-05 store = %s", fromStore.body)
	}
	inbox := scene.store.get("/api/v1/notifications")
	if !strings.Contains(string(inbox.body), "mengakhiri kemitraan.") {
		t.Fatalf("inbox = %s", inbox.body)
	}
}

func TestSetPublicListing_FRPRT03_StoreChoosesToBeShownAsAPlaceToBuy(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	path := "/api/v1/partnerships/" + partnershipID + "/public-listing"

	byStore := scene.store.patch(path, `{"show_on_public":true}`)
	byConsignor := scene.consignor.patch(path, `{"show_on_public":true}`)

	requireCode(t, byStore, "OK")
	loadContract(t).requireConforms(t, byStore)
	if !strings.Contains(string(byStore.body), `"show_on_public":true`) {
		t.Fatalf("by store = %s", byStore.body)
	}
	requireCode(t, byConsignor, "FORBIDDEN")
}
