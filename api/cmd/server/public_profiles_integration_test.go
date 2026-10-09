package main

import (
	"net/http"
	"strings"
	"testing"
)

const (
	publicStoresPath     = "/api/v1/public/stores/"
	publicConsignorsPath = "/api/v1/public/consignors/"
)

func (b *browser) visitorOnTheSameServer() *browser {
	return &browser{t: b.t, server: b.server, now: b.now, pool: b.pool}
}

func (f organizationFixture) slug() string {
	return quotedAfter(string(f.owner.get(organizationsPath+"/"+f.id).body), `"slug":"`)
}

func requireNoneOf(t *testing.T, got exchange, private ...string) {
	t.Helper()
	for _, value := range private {
		if strings.Contains(string(got.body), value) {
			t.Errorf("public answer carries %q: %s", value, got.body)
		}
	}
}

func TestPublicStore_FRPUB02_BRAUTHZ04_VisitorSeesOnlyWhitelistedFields(t *testing.T) {
	consignor, store := listedStoreScene(t)

	got := consignor.visitorOnTheSameServer().get(publicStoresPath + store.slug())

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
	if !strings.Contains(string(got.body), `"opening_hours":"07.00-21.00"`) ||
		!strings.Contains(string(got.body), `"commission_max_bps":2000`) ||
		!strings.Contains(string(got.body), `"address":"Jl. Sirojudin 12"`) {
		t.Fatalf("got = %s", got.body)
	}
	requireNoneOf(t, got, "whatsapp", "6281234567890", "payment", "sales")
}

func TestPublicStore_BRPUB02_UnlistedStoreIsNotFound(t *testing.T) {
	visitor := databaseBrowser(t)
	store := visitor.found(onboarding(semarangCityCode), "LOGO")

	got := visitor.visitorOnTheSameServer().get(publicStoresPath + store.slug())

	if got.status != http.StatusNotFound {
		t.Fatalf("got = %d %s", got.status, got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestPublicConsignor_FRPUB03_BRAUTHZ04_VisitorSeesProductsWithoutPrivateDetails(t *testing.T) {
	scene := listedConsignorScene(t)
	scene.consignor.patch(organizationsPath+"/"+scene.consignorID+"/consignor-profile", consignorProfileBody)

	consignor := organizationFixture{owner: scene.consignor, id: scene.consignorID}

	got := scene.consignor.visitorOnTheSameServer().get(publicConsignorsPath + consignor.slug())

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
	if !strings.Contains(string(got.body), `"business_id_no":"1234567890123"`) ||
		!strings.Contains(string(got.body), `"district":"Tembalang"`) ||
		!strings.Contains(string(got.body), `"suggested_retail_price":`) {
		t.Fatalf("got = %s", got.body)
	}
	requireNoneOf(t, got, "Jl. Sirojudin", "whatsapp", "6281234567890", "deposit", "payment", "110.4386")
}

func TestPublicConsignor_BRPUB02_StoreSlugIsNotAConsignor(t *testing.T) {
	consignor, store := listedStoreScene(t)

	got := consignor.visitorOnTheSameServer().get(publicConsignorsPath + store.slug())

	if got.status != http.StatusNotFound {
		t.Fatalf("got = %d %s", got.status, got.body)
	}
}
