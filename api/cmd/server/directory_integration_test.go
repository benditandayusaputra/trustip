package main

import (
	"net/http"
	"strings"
	"testing"
)

const directoryStoresPath = "/api/v1/directory/stores"

func listedStoreScene(t *testing.T) (*browser, organizationFixture) {
	t.Helper()
	consignor := consignorBrowser(t)
	consignor.postJSON(productsPath, productBody(""))
	store := consignor.anotherUserOnTheSameServer("tini@contoh.id")
	located := strings.Replace(onboarding(semarangCityCode), `"location":null`,
		`"location":{"latitude":-7.0546,"longitude":110.4386}`, 1)
	fixture := store.found(located, "LOGO")
	store.patch(organizationsPath+"/"+fixture.id+"/store-profile", storeProfileBody(""))
	if listed := store.patch(organizationsPath+"/"+fixture.id+"/listing", `{"is_listed":true}`); listed.status != 200 {
		t.Fatalf("listing = %d %s", listed.status, listed.body)
	}
	return consignor, fixture
}

func TestDirectoryStores_FRDIR01_ConsignorFindsTheListedStoreWithDistanceAndScore(t *testing.T) {
	consignor, store := listedStoreScene(t)

	got := consignor.get(directoryStoresPath + "?origin=HERE&lat=-7.0456&lng=110.4386&radius_km=2")

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
	body := string(got.body)
	if !strings.Contains(body, `"id":"`+store.id+`"`) || !strings.Contains(body, `"distance_m":100`) ||
		!strings.Contains(body, `"code":"ACCEPTS_CATEGORY"`) || !strings.Contains(body, `"is_expanded":true`) ||
		!strings.Contains(body, `"effective_radius_km":null`) {
		t.Fatalf("got = %s", got.body)
	}
}

func TestDirectoryStores_FRDIR06_StoreTypeFilterLeavesOtherTypesOut(t *testing.T) {
	consignor, _ := listedStoreScene(t)

	got := consignor.get(directoryStoresPath + "?origin=CITY&city_code=33.74&store_type=CANTEEN")

	requireCode(t, got, "OK")
	if !strings.Contains(string(got.body), `"data":[]`) {
		t.Fatalf("got = %s", got.body)
	}
}

func TestDirectoryStores_FRDIR05_UnknownRadiusIsAValidationError(t *testing.T) {
	consignor, _ := listedStoreScene(t)

	got := consignor.get(directoryStoresPath + "?radius_km=3")

	if got.status != http.StatusUnprocessableEntity {
		t.Fatalf("got = %d %s", got.status, got.body)
	}
	loadContract(t).requireConforms(t, got)
}

const locatedBody = `"location":{"latitude":-7.0546,"longitude":110.4386}`

type consignorDirectoryScene struct {
	store       *browser
	consignor   *browser
	consignorID string
	productID   string
}

func listedConsignorScene(t *testing.T) consignorDirectoryScene {
	t.Helper()
	consignor := databaseBrowser(t)
	founded := consignor.postOrganization(strings.Replace(consignorOnboarding(), `"location":null`, locatedBody, 1))
	consignor.post("/api/v1/sessions/refresh")
	scene := consignorDirectoryScene{
		consignor: consignor, consignorID: quotedAfter(string(founded.body), `"id":"`),
	}
	product := consignor.postJSON(productsPath, productBody(""))
	scene.productID = quotedAfter(string(product.body), `"id":"`)
	consignor.patch(productsPath+"/"+scene.productID+"/seeking", `{"is_seeking_stores":true}`)
	listed := consignor.patch(organizationsPath+"/"+scene.consignorID+"/listing", `{"is_listed":true}`)
	if listed.status != http.StatusOK {
		t.Fatalf("listing = %d %s", listed.status, listed.body)
	}
	scene.store = consignor.anotherUserOnTheSameServer("tini@contoh.id")
	store := scene.store.found(strings.Replace(onboarding(semarangCityCode), `"location":null`, locatedBody, 1), "LOGO")
	scene.store.patch(organizationsPath+"/"+store.id+"/store-profile", storeProfileBody(""))
	return scene
}

func TestDirectoryProducts_FRDIR02_StoreFindsTheProductSeekingStores(t *testing.T) {
	scene := listedConsignorScene(t)

	got := scene.store.get("/api/v1/directory/products?origin=CITY&city_code=33.74&radius_km=all")

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
	body := string(got.body)
	if !strings.Contains(body, `"id":"`+scene.productID+`"`) || !strings.Contains(body, `"score":`) ||
		strings.Contains(body, "deposit") {
		t.Fatalf("got = %s", got.body)
	}
}

func TestDirectoryConsignors_FRDIR07_HalalFilterLeavesTheConsignorOut(t *testing.T) {
	scene := listedConsignorScene(t)

	all := scene.store.get("/api/v1/directory/consignors?radius_km=2")
	halal := scene.store.get("/api/v1/directory/consignors?has_halal=true")

	requireCode(t, all, "OK")
	loadContract(t).requireConforms(t, all)
	if !strings.Contains(string(all.body), scene.consignorID) || !strings.Contains(string(halal.body), `"data":[]`) {
		t.Fatalf("all = %s\nhalal = %s", all.body, halal.body)
	}
}
