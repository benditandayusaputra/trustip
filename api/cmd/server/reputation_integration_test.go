package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	trustrepository "github.com/benditandayusaputra/trustip/api/internal/modules/trust/repository"
	trustservice "github.com/benditandayusaputra/trustip/api/internal/modules/trust/service"
)

func (s paymentScene) issueAnotherSettlement(t *testing.T, round int) string {
	t.Helper()
	s.store.now.Advance(time.Minute)
	requireCreated(t, s.store.recordDrop(dropBody(s.partnershipID, s.productID, 20),
		fmt.Sprintf("0192f0c4-0000-7000-8000-0000000d%04d", round)))
	counted := s.store.recordCountWithKey(countBody(s.partnershipID, s.productID, countLine{recorded: 20}),
		fmt.Sprintf("0192f0c4-0000-7000-8000-0000000c%04d", round))
	requireCreated(t, counted)
	return quotedAfter(string(counted.body), `"settlement":{"id":"`)
}

func TestReputation_BRTRS01_BRTRS02_DailySummaryShowsTheStoreOnTimeShare(t *testing.T) {
	scene := issuedSettlement(t)
	second := scene.issueAnotherSettlement(t, 1)
	scene.issueAnotherSettlement(t, 2)
	requireCode(t, scene.consignor.post(settlementsPath+"/"+scene.settlementID+"/mark-paid"), "OK")
	requireCode(t, scene.consignor.post(settlementsPath+"/"+second+"/mark-paid"), "OK")
	storeID := scene.store.actor(t).Membership.OrganizationID
	requireCode(t, scene.store.patch(organizationsPath+"/"+storeID,
		strings.Replace(profileUpdate(""), `"location":null`, locatedBody, 1)), "OK")
	scene.store.patch(organizationsPath+"/"+storeID+"/store-profile", storeProfileBody(""))
	requireCode(t, scene.store.patch(organizationsPath+"/"+storeID+"/listing", `{"is_listed":true}`), "OK")
	before := scene.consignor.get(directoryStoresPath + "?radius_km=all")
	scene.store.now.Advance(8 * 24 * time.Hour)
	scene.consignor.post("/api/v1/sessions/refresh")

	refreshed, err := trustservice.NewRefreshReputation(trustrepository.NewReputation(scene.store.pool)).
		Execute(t.Context(), scene.store.now.Now())

	if err != nil || refreshed == 0 {
		t.Fatalf("refreshed = %d, err = %v", refreshed, err)
	}
	after := scene.consignor.get(directoryStoresPath + "?radius_km=all")
	requireCode(t, after, "OK")
	loadContract(t).requireConforms(t, after)
	if !strings.Contains(string(before.body), `"reputation":{"is_new":true,"on_time_bps":null,"active_partners":0}`) ||
		!strings.Contains(string(after.body), `"reputation":{"is_new":false,"on_time_bps":6667,"active_partners":1}`) {
		t.Fatalf("before = %s\nafter = %s", before.body, after.body)
	}
}
