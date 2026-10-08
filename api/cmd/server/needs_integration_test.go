package main

import (
	"net/http"
	"strings"
	"testing"
)

const needsPath = "/api/v1/needs"

func needBody(title string) string {
	return `{"title":"` + title + `","category_id":"` + breadCategoryID + `",` +
		`"description":"Roti manis isi cokelat untuk kantin.","est_qty_per_week":50,"max_retail_price":5000,` +
		`"preferred_model":"COMMISSION","offered_commission_bps":2000}`
}

func storeWithNeeds(t *testing.T) *browser {
	t.Helper()
	store := databaseBrowser(t)
	store.found(onboarding(semarangCityCode), "LOGO")
	return store
}

func TestNeeds_FRNED01_FRNED06_StorePostsListsAndClosesANeed(t *testing.T) {
	store := storeWithNeeds(t)
	contract := loadContract(t)

	posted := store.postJSON(needsPath, needBody("Butuh roti untuk sarapan"))
	needID := quotedAfter(string(posted.body), `"id":"`)
	listed := store.get(needsPath + "?status=OPEN")
	detail := store.get(needsPath + "/" + needID)
	extended := store.post(needsPath + "/" + needID + "/extend")
	fulfilled := store.post(needsPath + "/" + needID + "/fulfill")
	again := store.post(needsPath + "/" + needID + "/close")

	if posted.status != http.StatusCreated {
		t.Fatalf("posted = %d %s", posted.status, posted.body)
	}
	for _, got := range []exchange{posted, listed, detail, extended, fulfilled, again} {
		contract.requireConforms(t, got)
	}
	requireCode(t, again, "NEED_NOT_OPEN")
	if !strings.Contains(string(listed.body), needID) || !strings.Contains(string(detail.body), `"responses":[]`) ||
		!strings.Contains(string(fulfilled.body), `"status":"FULFILLED"`) {
		t.Fatalf("listed = %s\ndetail = %s\nfulfilled = %s", listed.body, detail.body, fulfilled.body)
	}
}

func TestNeeds_BRNED01_SixthOpenNeedIsRefused(t *testing.T) {
	store := storeWithNeeds(t)
	for index := range 5 {
		got := store.postJSON(needsPath, needBody("Butuh roti nomor "+string(rune('1'+index))))
		if got.status != http.StatusCreated {
			t.Fatalf("need %d = %d %s", index, got.status, got.body)
		}
	}

	sixth := store.postJSON(needsPath, needBody("Butuh roti keenam"))

	requireCode(t, sixth, "NEED_LIMIT_REACHED")
	loadContract(t).requireConforms(t, sixth)
}

func TestNeeds_BRNED02_PhoneNumberInTheDescriptionIsRefused(t *testing.T) {
	store := storeWithNeeds(t)

	got := store.postJSON(needsPath, strings.Replace(needBody("Butuh roti untuk sarapan"),
		"untuk kantin.", "hubungi 0812 3456 7890.", 1))

	if got.status != http.StatusUnprocessableEntity || !strings.Contains(string(got.body), "CONTACT_NOT_ALLOWED") {
		t.Fatalf("got = %d %s", got.status, got.body)
	}
}

func TestNeeds_BRAUTHZ01_ConsignorCannotPostANeed(t *testing.T) {
	consignor := consignorBrowser(t)

	got := consignor.postJSON(needsPath, needBody("Butuh roti untuk sarapan"))

	requireCode(t, got, "FORBIDDEN")
}
