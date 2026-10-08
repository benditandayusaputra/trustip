package main

import (
	"strings"
	"testing"
)

const salesReportPath = "/api/v1/reports/sales?from=2026-10-19&to=2026-10-19"

func TestSalesReport_BRRPT03_TotalsMatchTheSettlementOfTheSameRange(t *testing.T) {
	scene := issuedSettlement(t)

	byPartner := scene.store.get(salesReportPath)
	byProduct := scene.consignor.get(salesReportPath + "&group_by=PRODUCT")
	settlement := scene.store.get(settlementsPath + "/" + scene.settlementID)

	requireCode(t, byPartner, "OK")
	loadContract(t).requireConforms(t, byPartner)
	split := `"split":{"gross":36000,"store_share":4500,"consignor_share":31500}`
	if !strings.Contains(string(settlement.body), split) {
		t.Fatalf("settlement = %s", settlement.body)
	}
	if !strings.Contains(string(byPartner.body), `"summary":{"sold_qty":18,"returned_qty":2,"received_qty":20,`+
		`"returned_bps":1000,`+split) || !strings.Contains(string(byPartner.body), `"label":"Warung Bu Tini"`) {
		t.Fatalf("by partner = %s", byPartner.body)
	}
	if !strings.Contains(string(byProduct.body), split) ||
		!strings.Contains(string(byProduct.body), `"group_by":"PRODUCT"`) {
		t.Fatalf("by product = %s", byProduct.body)
	}
}

func TestSalesReport_BRRPT04_RangeLongerThan366DaysIsRefused(t *testing.T) {
	owner := consignorBrowser(t)

	got := owner.get("/api/v1/reports/sales?from=2026-01-01&to=2027-01-02")

	requireCode(t, got, "VALIDATION_FAILED")
	loadContract(t).requireConforms(t, got)
}
