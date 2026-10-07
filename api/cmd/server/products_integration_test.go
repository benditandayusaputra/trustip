package main

import (
	"net/http"
	"strings"
	"testing"
)

func consignorBrowser(t *testing.T) *browser {
	t.Helper()
	visitor := databaseBrowser(t)
	visitor.postOrganization(consignorOnboarding())
	visitor.post("/api/v1/sessions/refresh")
	return visitor
}

func TestProducts_FRPRD03_CreateEditArchiveAndReactivateAgainstTheDatabase(t *testing.T) {
	visitor := consignorBrowser(t)
	contract := loadContract(t)
	created := visitor.postProduct(productBody(""))
	productPath := productsPath + "/" + quotedAfter(string(created.body), `"id":"`)

	edited := visitor.patch(productPath, strings.Replace(productBody("ROT-0001"), `"pcs"`, `"bungkus"`, 1))
	archived := visitor.post(productPath + "/archive")
	active := visitor.get(productsPath)
	reactivated := visitor.post(productPath + "/reactivate")

	for _, got := range []exchange{created, edited, archived, active, reactivated} {
		contract.requireConforms(t, got)
	}
	if !strings.Contains(string(edited.body), `"unit":"bungkus"`) ||
		!strings.Contains(string(archived.body), `"status":"ARCHIVED"`) ||
		!strings.Contains(string(active.body), `"data":[]`) ||
		!strings.Contains(string(reactivated.body), `"status":"ACTIVE"`) {
		t.Fatalf("edited = %s\narchived = %s\nactive = %s", edited.body, archived.body, active.body)
	}
}

func TestProducts_FRPRD05_SearchFindsByNameAndSKU(t *testing.T) {
	visitor := consignorBrowser(t)
	visitor.postProduct(productBody(""))
	visitor.postProduct(strings.Replace(productBody(""), "Roti Sobek Cokelat", "Keripik Pedas", 1))

	byName := visitor.get(productsPath + "?q=keripik")
	bySKU := visitor.get(productsPath + "?q=rot-0001")
	none := visitor.get(productsPath + "?q=%25")

	if strings.Count(string(byName.body), `"sku":`) != 1 || !strings.Contains(string(byName.body), "KER-0001") ||
		strings.Count(string(bySKU.body), `"sku":`) != 1 || strings.Count(string(none.body), `"sku":`) != 0 {
		t.Fatalf("byName = %s\nbySKU = %s\nnone = %s", byName.body, bySKU.body, none.body)
	}
	loadContract(t).requireConforms(t, byName)
}

func TestProducts_BRPRD01_DuplicateSKUIsAFieldError(t *testing.T) {
	visitor := consignorBrowser(t)
	visitor.postProduct(productBody("ROT-0001"))

	got := visitor.postProduct(productBody("rot-0001"))

	requireCode(t, got, "VALIDATION_FAILED")
	if got.status != http.StatusUnprocessableEntity || !strings.Contains(string(got.body), `"code":"TAKEN"`) {
		t.Fatalf("got = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestProducts_FRPRD01_ProductPhotoUploadedByTheConsignorIsLinked(t *testing.T) {
	visitor := consignorBrowser(t)
	uploaded := visitor.upload(part{name: "purpose", data: []byte("PRODUCT_PHOTO")},
		part{name: "file", filename: "roti.jpg", data: photoWithLocation(t)})
	photoID := quotedAfter(string(uploaded.body), `"id":"`)

	got := visitor.postProduct(strings.Replace(productBody(""), `"photo_file_id":null`,
		`"photo_file_id":"`+photoID+`"`, 1))

	if got.status != http.StatusCreated || !strings.Contains(string(got.body), `"photo":{"id":"`+photoID+`"`) {
		t.Fatalf("got = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestProducts_FRPRD04_SeekingStoresToggleIsSaved(t *testing.T) {
	visitor := consignorBrowser(t)
	created := visitor.postProduct(productBody(""))
	productPath := productsPath + "/" + quotedAfter(string(created.body), `"id":"`)

	got := visitor.patch(productPath+"/seeking", `{"is_seeking_stores":true}`)

	requireCode(t, got, "OK")
	if !strings.Contains(string(got.body), `"is_seeking_stores":true`) ||
		!strings.Contains(string(got.body), `"consignor_slug":"warung-bu-tini"`) {
		t.Fatalf("got = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}
