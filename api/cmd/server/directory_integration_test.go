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
