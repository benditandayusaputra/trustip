package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestBargaining_FRPRP04_StoreCountersTheDepositPriceAndTheConsignorAccepts(t *testing.T) {
	consignor, store := listedStoreScene(t)
	product := consignor.postJSON(productsPath, strings.Replace(productBody(""), "Roti Sobek Cokelat", "Bolu", 1))
	productID := quotedAfter(string(product.body), `"id":"`)
	sent := consignor.postJSON("/api/v1/proposals", directBody(store.id, productID))
	proposalID := quotedAfter(string(sent.body), `"id":"`)
	countered := strings.Replace(invitationBody(productID), `"deposit_price":1750`, `"deposit_price":1700`, 1)
	contract := loadContract(t)

	counter := store.owner.postJSON("/api/v1/proposals/"+proposalID+"/counter", countered)
	history := consignor.get("/api/v1/proposals/" + proposalID + "/revisions")
	accepted := consignor.post("/api/v1/proposals/" + proposalID + "/accept")

	for _, got := range []exchange{counter, history, accepted} {
		requireCode(t, got, "OK")
		contract.requireConforms(t, got)
	}
	if !strings.Contains(string(history.body), `"change":"CHANGED"`) ||
		!strings.Contains(string(history.body), `"revision_no":2`) ||
		!strings.Contains(string(accepted.body), `"status":"ACCEPTED"`) {
		t.Fatalf("history = %s\naccepted = %s", history.body, accepted.body)
	}
	partnershipID := quotedAfter(string(accepted.body), `"partnership_id":"`)
	agreement := store.owner.get("/api/v1/partnerships/" + partnershipID + "/agreement")
	if !strings.Contains(string(agreement.body), `"deposit_price":1700`) {
		t.Fatalf("agreement = %s", agreement.body)
	}
}

func TestBargaining_BRPRP02_AuthorOfTheCounterWaitsForTheOtherSide(t *testing.T) {
	consignor, store := listedStoreScene(t)
	sent := consignor.postJSON("/api/v1/proposals", directBody(store.id, ""))
	proposalID := quotedAfter(string(sent.body), `"id":"`)

	again := consignor.postJSON("/api/v1/proposals/"+proposalID+"/counter", invitationBody(""))

	if again.status != http.StatusConflict {
		t.Fatalf("again = %d %s", again.status, again.body)
	}
	requireCode(t, again, "NOT_YOUR_TURN")
}
