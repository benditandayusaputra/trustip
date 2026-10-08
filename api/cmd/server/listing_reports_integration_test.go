package main

import (
	"net/http"
	"strings"
	"testing"
)

const listingReportsPath = "/api/v1/listing-reports"

func reportBody(targetType, targetID, reason string) string {
	return `{"target_type":"` + targetType + `","target_id":"` + targetID + `","reason":"` + reason +
		`","note":"Foto produk bukan milik toko ini."}`
}

func TestReportListing_FRTRS04_ConsignorReportsAStoreAndReportingAgainUpdatesIt(t *testing.T) {
	consignor, store := listedStoreScene(t)

	first := consignor.postJSON(listingReportsPath, reportBody("ORGANIZATION", store.id, "MISLEADING"))
	again := consignor.postJSON(listingReportsPath, reportBody("ORGANIZATION", store.id, "SPAM"))

	if first.status != http.StatusCreated || again.status != http.StatusCreated {
		t.Fatalf("first = %d %s\nagain = %d %s", first.status, first.body, again.status, again.body)
	}
	loadContract(t).requireConforms(t, first)
	if quotedAfter(string(first.body), `"id":"`) != quotedAfter(string(again.body), `"id":"`) ||
		!strings.Contains(string(again.body), `"status":"OPEN"`) {
		t.Fatalf("first = %s\nagain = %s", first.body, again.body)
	}
}

func TestReportListing_BRAUTHZ02_OwnOrMissingTargetIsNotFound(t *testing.T) {
	consignor, store := listedStoreScene(t)

	own := store.owner.postJSON(listingReportsPath, reportBody("ORGANIZATION", store.id, "SPAM"))
	missing := consignor.postJSON(listingReportsPath,
		reportBody("NEED", "0192f0c4-0000-7000-8000-00000000ffff", "SPAM"))
	invalid := consignor.postJSON(listingReportsPath, reportBody("USER", store.id, "OTHER"))

	requireCode(t, own, "NOT_FOUND")
	requireCode(t, missing, "NOT_FOUND")
	requireCode(t, invalid, "VALIDATION_FAILED")
	loadContract(t).requireConforms(t, invalid)
}

func TestReportListing_RSEC43_SixthReportInAnHourIsRateLimited(t *testing.T) {
	consignor, store := listedStoreScene(t)

	var last exchange
	for range reportsPerHourPerUser + 1 {
		last = consignor.postJSON(listingReportsPath, reportBody("ORGANIZATION", store.id, "SPAM"))
	}

	requireCode(t, last, "RATE_LIMITED")
}
