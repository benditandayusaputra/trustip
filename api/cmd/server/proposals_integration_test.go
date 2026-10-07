package main

import (
	"net/http"
	"strings"
	"testing"
)

type invitationScene struct {
	consignor  *browser
	store      *browser
	productID  string
	inviteCode string
	proposalID string
}

func newInvitationScene(t *testing.T) invitationScene {
	t.Helper()
	consignor := consignorBrowser(t)
	product := consignor.postJSON(productsPath, productBody(""))
	scene := invitationScene{consignor: consignor, productID: quotedAfter(string(product.body), `"id":"`)}
	sent := consignor.postJSON(invitationsPath, invitationBody(scene.productID))
	if sent.status != http.StatusCreated {
		t.Fatalf("send = %d %s", sent.status, sent.body)
	}
	loadContract(t).requireConforms(t, sent)
	scene.inviteCode = quotedAfter(string(sent.body), `"invite_code":"`)
	scene.proposalID = quotedAfter(string(sent.body), `"id":"`)
	scene.store = consignor.anotherUserOnTheSameServer("tini@contoh.id")
	scene.store.found(onboarding(semarangCityCode), "LOGO")
	return scene
}

func TestInvitation_FRPRP07_StoreOpensAndClaimsTheConsignorInvitation(t *testing.T) {
	scene := newInvitationScene(t)
	contract := loadContract(t)

	viewed := scene.store.get(invitationsPath + "/" + scene.inviteCode)
	claimed := scene.store.post(invitationsPath + "/" + scene.inviteCode + "/claim")
	read := scene.store.get("/api/v1/proposals/" + scene.proposalID)

	for _, got := range []exchange{viewed, claimed, read} {
		requireCode(t, got, "OK")
		contract.requireConforms(t, got)
	}
	if !strings.Contains(string(viewed.body), `"item_count":1`) ||
		!strings.Contains(string(read.body), `"per_unit":{"gross":2000,"store_share":250,"consignor_share":1750}`) ||
		!strings.Contains(string(read.body), `"number":"AJ-`) {
		t.Fatalf("viewed = %s\nread = %s", viewed.body, read.body)
	}
}

func TestInvitation_BRPRP07_ConsignorOpeningAConsignorInvitationIsToldItIsForStores(t *testing.T) {
	scene := newInvitationScene(t)
	other := scene.consignor.anotherUserOnTheSameServer("ani@contoh.id")
	other.found(strings.Replace(consignorOnboarding(), "Warung Bu Tini", "Penitip Lain", 1), "LOGO")

	got := other.get(invitationsPath + "/" + scene.inviteCode)

	requireCode(t, got, "INVITE_WRONG_ORG_TYPE")
	if !strings.Contains(string(got.body), "Undangan ini untuk toko.") {
		t.Fatalf("got = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}
