package main

import (
	"net/http"
	"strings"
	"testing"
)

func directBody(targetID, productID string) string {
	return strings.Replace(invitationBody(productID), `{"terms"`,
		`{"target_organization_id":"`+targetID+`","terms"`, 1)
}

func TestSendProposal_FRPRP01_ConsignorProposesToAStoreFoundInTheDirectory(t *testing.T) {
	consignor, store := listedStoreScene(t)
	product := consignor.postJSON(productsPath, strings.Replace(productBody(""), "Roti Sobek Cokelat", "Bolu", 1))
	productID := quotedAfter(string(product.body), `"id":"`)

	sent := consignor.postJSON("/api/v1/proposals", directBody(store.id, productID))
	received := store.owner.get("/api/v1/proposals?box=INCOMING")

	if sent.status != http.StatusCreated {
		t.Fatalf("sent = %d %s", sent.status, sent.body)
	}
	loadContract(t).requireConforms(t, sent)
	if !strings.Contains(string(received.body), quotedAfter(string(sent.body), `"id":"`)) {
		t.Fatalf("received = %s", received.body)
	}
}

func TestSendProposal_BRPRP06_SenderWaitsSevenDaysAfterADecline(t *testing.T) {
	consignor, store := listedStoreScene(t)
	sent := consignor.postJSON("/api/v1/proposals", directBody(store.id, ""))
	proposalID := quotedAfter(string(sent.body), `"id":"`)
	store.owner.postJSON("/api/v1/proposals/"+proposalID+"/decline", `{"reason":"Rak sudah penuh."}`)

	again := consignor.postJSON("/api/v1/proposals", directBody(store.id, ""))

	requireCode(t, again, "PROPOSAL_COOLDOWN")
	loadContract(t).requireConforms(t, again)
}

func TestSendProposal_BRAUTHZ02_UnlistedTargetIsNotFound(t *testing.T) {
	consignor, store := listedStoreScene(t)
	store.owner.patch(organizationsPath+"/"+store.id+"/listing", `{"is_listed":false}`)

	got := consignor.postJSON("/api/v1/proposals", directBody(store.id, ""))

	requireCode(t, got, "NOT_FOUND")
}

func TestDirectoryContact_FRDIR11_SignedInConsignorReadsTheStoreWhatsApp(t *testing.T) {
	consignor, store := listedStoreScene(t)

	got := consignor.get("/api/v1/directory/organizations/" + store.id)

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
	if !strings.Contains(string(got.body), `"whatsapp":"6281234567890"`) {
		t.Fatalf("got = %s", got.body)
	}
}
