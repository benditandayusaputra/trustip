package main

import (
	"net/http"
	"regexp"
	"strings"
	"testing"
)

var sealCodeFormat = regexp.MustCompile(`"code":"([A-Z2-7]{4}(?:-[A-Z2-7]{4}){3})"`)

func sealCodeOf(t *testing.T, body []byte) string {
	t.Helper()
	found := sealCodeFormat.FindSubmatch(body)
	if found == nil {
		t.Fatalf("no seal code in %s", body)
	}
	return string(found[1])
}

func TestVerifySeal_FRSIG03_RSEC34_VisitorSeesTheSealWithoutAmounts(t *testing.T) {
	scene := recordedDrop(t)
	code := sealCodeOf(t, scene.store.get(scene.path).body)
	visitor := scene.consignor.visitorOnTheSameServer()

	got := visitor.get("/api/v1/public/seals/" + strings.ToLower(strings.ReplaceAll(code, "-", "")))

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
	if !strings.Contains(string(got.body), `"status":"VALID"`) ||
		!strings.Contains(string(got.body), `"event_type":"DROP_RECORDED"`) ||
		!strings.Contains(string(got.body), `"code":"`+code+`"`) {
		t.Fatalf("got = %s", got.body)
	}
	requireNoneOf(t, got, "35000", "1750", `"qty"`, "deposit", "commission", "terms")
}

func TestVerifySeal_FRSIG03_UnknownCodeIsNotFound(t *testing.T) {
	scene := recordedDrop(t)

	got := scene.consignor.visitorOnTheSameServer().get("/api/v1/public/seals/AAAA-AAAA-AAAA-AAAA")

	if got.status != http.StatusNotFound {
		t.Fatalf("got = %d %s", got.status, got.body)
	}
}

func TestSealKeys_FRSIG04_PublicKeyIsPublished(t *testing.T) {
	scene := recordedDrop(t)

	got := scene.consignor.visitorOnTheSameServer().get("/api/v1/public/seal-keys")

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
	if !strings.Contains(string(got.body), `"key_id":"v1","algorithm":"Ed25519"`) {
		t.Fatalf("got = %s", got.body)
	}
}

func TestCheckChains_FRSIG05_RecordedPartnershipChainIsIntact(t *testing.T) {
	scene := recordedDrop(t)
	chains := buildTradeComponents(testFoundation(scene.store.pool, scene.store.now), testSealKeys{}).seals.chains

	report, err := chains.Execute(t.Context())
	if err != nil {
		t.Fatalf("check chains: %v", err)
	}

	if report.Checked == 0 || len(report.Broken) != 0 {
		t.Fatalf("report = %+v", report)
	}
}
