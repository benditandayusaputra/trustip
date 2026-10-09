package main

import (
	"net/http"
	"strings"
	"testing"
)

func publicProductPath(scene invitationScene) string {
	product := string(scene.consignor.get(productsPath + "/" + scene.productID).body)
	return "/api/v1/public/products/" + quotedAfter(product, `"consignor_slug":"`) + "/" +
		quotedAfter(product, `"sku":"`)
}

func TestPublicProduct_BRPUB01_StoreAppearsOnlyAfterItChoosesToBeShown(t *testing.T) {
	scene := recordedDrop(t)
	path := publicProductPath(scene.invitationScene)
	visitor := scene.consignor.visitorOnTheSameServer()

	hidden := visitor.get(path)
	scene.store.patch("/api/v1/partnerships/"+scene.partnershipID+"/public-listing", `{"show_on_public":true}`)
	shown := visitor.get(path + "?lat=-7.0456&lng=110.4386")

	requireCode(t, hidden, "OK")
	loadContract(t).requireConforms(t, shown)
	if !strings.Contains(string(hidden.body), `"stores":[]`) ||
		!strings.Contains(string(shown.body), `"composition":"Tepung, gula"`) ||
		!strings.Contains(string(shown.body), `"suggested_retail_price":2000`) ||
		strings.Contains(string(shown.body), `"stores":[]`) {
		t.Fatalf("hidden = %s\nshown = %s", hidden.body, shown.body)
	}
	requireNoneOf(t, shown, "1750", "deposit", "commission", "balance", "whatsapp", "6281234567890")
}

func TestPublicProduct_FRPUB04_UnknownSKUIsNotFound(t *testing.T) {
	scene := recordedDrop(t)
	path := publicProductPath(scene.invitationScene)

	got := scene.consignor.visitorOnTheSameServer().get(path + "X")

	if got.status != http.StatusNotFound {
		t.Fatalf("got = %d %s", got.status, got.body)
	}
}

func TestPublicProduct_FRPUB05_PositionOutOfRangeIsAValidationError(t *testing.T) {
	scene := recordedDrop(t)

	got := scene.consignor.visitorOnTheSameServer().get(publicProductPath(scene.invitationScene) + "?lat=95&lng=1")

	if got.status != http.StatusUnprocessableEntity {
		t.Fatalf("got = %d %s", got.status, got.body)
	}
}
