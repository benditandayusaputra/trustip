package main

import (
	"net/http"
	"strings"
	"testing"
)

func amendmentBody(productID string) string {
	return `{"terms":{"settlement_schedule":"WEEKLY","payment_terms_days":14,"count_rhythm":"DAILY","notes":""},` +
		`"items":[{"product_id":"` + productID + `","action":"ADD","pricing_model":"COMMISSION",` +
		`"retail_price":15000,"commission_bps":2000,"deposit_price":0}],"message":"Tambah keripik."}`
}

func TestAmendment_FRAGR06_AddedProductBecomesVersionTwoAfterTheStoreAccepts(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	contract := loadContract(t)
	chipsBody := strings.Replace(productBody(""), "Roti Sobek Cokelat", "Keripik", 1)
	chips := scene.consignor.postJSON(productsPath, chipsBody)
	chipsID := quotedAfter(string(chips.body), `"id":"`)

	sent := scene.consignor.postJSON("/api/v1/partnerships/"+partnershipID+"/amendments", amendmentBody(chipsID))
	accepted := scene.store.post("/api/v1/proposals/" + quotedAfter(string(sent.body), `"id":"`) + "/accept")
	current := scene.consignor.get("/api/v1/partnerships/" + partnershipID + "/agreement")
	first := scene.consignor.get("/api/v1/partnerships/" + partnershipID + "/agreements/1")
	history := scene.store.get("/api/v1/partnerships/" + partnershipID + "/agreements")

	if sent.status != http.StatusCreated || accepted.status != http.StatusOK {
		t.Fatalf("sent = %s\naccepted = %s", sent.body, accepted.body)
	}
	for _, got := range []exchange{sent, accepted, current, first, history} {
		contract.requireConforms(t, got)
	}
	if !strings.Contains(string(current.body), `"version":2`) ||
		!strings.Contains(string(current.body), `"is_valid":true`) ||
		!strings.Contains(string(first.body), `"status":"SUPERSEDED"`) ||
		!strings.Contains(string(first.body), `"is_valid":true`) ||
		!strings.Contains(string(history.body), `"kind":"ADDED","before":null`) ||
		!strings.Contains(string(history.body), `{"field":"payment_terms_days","before":"7","after":"14"}`) {
		t.Fatalf("current = %s\nfirst = %s\nhistory = %s", current.body, first.body, history.body)
	}
}

func TestAmendment_BRPRP04_SecondOpenAmendmentIsRefused(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	path := "/api/v1/partnerships/" + partnershipID + "/amendments"
	stop := strings.Replace(amendmentBody(scene.productID), `"action":"ADD"`, `"action":"STOP"`, 1)
	scene.consignor.postJSON(path, stop)

	again := scene.store.postJSON(path, stop)

	requireCode(t, again, "PROPOSAL_ALREADY_OPEN")
	loadContract(t).requireConforms(t, again)
}

func TestAmendment_FRPRP08_ProductAlreadyInTheAgreementCannotBeAddedAgain(t *testing.T) {
	scene, partnershipID := acceptedScene(t)

	got := scene.consignor.postJSON("/api/v1/partnerships/"+partnershipID+"/amendments",
		amendmentBody(scene.productID))

	requireCode(t, got, "VALIDATION_FAILED")
	if !strings.Contains(string(got.body), `"field":"items[0].product_id"`) {
		t.Fatalf("got = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}
