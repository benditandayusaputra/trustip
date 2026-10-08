package main

import (
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
)

func acceptedScene(t *testing.T) (invitationScene, string) {
	t.Helper()
	return acceptedSceneWith(t, "WEEKLY")
}

func acceptedSceneWith(t *testing.T, schedule string) (invitationScene, string) {
	t.Helper()
	scene := newInvitationSceneWith(t, schedule)
	scene.claimAsStore(t)
	accepted := scene.store.post("/api/v1/proposals/" + scene.proposalID + "/accept")
	partnershipID := quotedAfter(string(accepted.body), `"partnership_id":"`)
	if partnershipID == "" {
		t.Fatalf("accepted = %s", accepted.body)
	}
	return scene, partnershipID
}

func TestAgreement_FRAGR01_BothPartiesSeeTheSameSealedAgreement(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	path := "/api/v1/partnerships/" + partnershipID + "/agreement"

	fromStore := scene.store.get(path)
	fromConsignor := scene.consignor.get(path)

	requireCode(t, fromStore, "OK")
	loadContract(t).requireConforms(t, fromStore)
	if stripRequestID(fromStore.body) != stripRequestID(fromConsignor.body) ||
		!strings.Contains(string(fromStore.body), `"is_valid":true`) ||
		!strings.Contains(string(fromStore.body), `"event_type":"AGREEMENT_ACCEPTED"`) ||
		!strings.Contains(string(fromStore.body), `"store_share":250,"consignor_share":1750`) {
		t.Fatalf("store = %s\nconsignor = %s", fromStore.body, fromConsignor.body)
	}
}

func TestAgreement_RSEC33_PriceChangedInTheDatabaseShowsAnInvalidSeal(t *testing.T) {
	scene, partnershipID := acceptedScene(t)
	path := "/api/v1/partnerships/" + partnershipID + "/agreement"
	agreementID := quotedAfter(string(scene.store.get(path).body), `"id":"`)
	tampered, err := sqlc.New(scene.store.pool.Querier(t.Context())).TamperAgreementItemPrice(t.Context(),
		sqlc.TamperAgreementItemPriceParams{AgreementID: agreementID, ProductID: scene.productID, RetailPrice: 2500})
	if err != nil || tampered != 1 {
		t.Fatalf("tamper = %d, err = %v", tampered, err)
	}

	fromStore := scene.store.get(path)
	fromConsignor := scene.consignor.get(path)

	for _, got := range []exchange{fromStore, fromConsignor} {
		if !strings.Contains(string(got.body), `"is_valid":false`) {
			t.Fatalf("got = %s", got.body)
		}
	}
	loadContract(t).requireConforms(t, fromStore)
}
