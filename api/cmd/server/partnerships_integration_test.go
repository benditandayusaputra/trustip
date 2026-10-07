package main

import (
	"strings"
	"testing"
)

func TestPartnerships_FRPRT01_BothSidesListTheirPartner(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	contract := loadContract(t)

	fromStore := scene.store.get("/api/v1/partnerships?status=ACTIVE")
	fromConsignor := scene.consignor.get("/api/v1/partnerships")

	for _, got := range []exchange{fromStore, fromConsignor} {
		requireCode(t, got, "OK")
		contract.requireConforms(t, got)
		if !strings.Contains(string(got.body), partnershipID) {
			t.Fatalf("got = %s", got.body)
		}
	}
}

func TestPartnerships_BRORG04_ActivePartnerSeesTheConsignorAddress(t *testing.T) {
	scene, partnershipID := acceptedScene(t)

	got := scene.store.get("/api/v1/partnerships/" + partnershipID)

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
	if !strings.Contains(string(got.body), `"address":"Jl. Sirojudin 12"`) ||
		!strings.Contains(string(got.body), `"type":"CONSIGNOR"`) {
		t.Fatalf("got = %s", got.body)
	}
}

func TestPartnerships_FRPRD06_StoreSeesTheCatalogWithAgreedTerms(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	scene.consignor.postJSON(productsPath, strings.Replace(productBody(""), "Roti Sobek Cokelat", "Keripik", 1))

	got := scene.store.get("/api/v1/partnerships/" + partnershipID + "/catalog")

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
	body := string(got.body)
	if strings.Count(body, `"product_id"`) != 2 || !strings.Contains(body, `"consignor_share":1750`) ||
		!strings.Contains(body, `"terms":null`) {
		t.Fatalf("got = %s", got.body)
	}
}
