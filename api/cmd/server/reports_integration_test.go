package main

import (
	"strings"
	"testing"
	"time"
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

func TestDailyRecap_FRRPT01_StoreSeesEachConsignorAndWhatIsAlreadyPaid(t *testing.T) {
	scene := issuedSettlement(t)
	scene.consignor.post(settlementsPath + "/" + scene.settlementID + "/mark-paid")

	recap := scene.store.get("/api/v1/reports/daily")
	fromConsignor := scene.consignor.get("/api/v1/reports/daily")

	requireCode(t, recap, "OK")
	loadContract(t).requireConforms(t, recap)
	if !strings.Contains(string(recap.body), `"date":"2026-10-19"`) ||
		!strings.Contains(string(recap.body), `"received_qty":20`) ||
		!strings.Contains(string(recap.body), `"consignor_share":31500},"paid":31500`) {
		t.Fatalf("recap = %s", recap.body)
	}
	requireCode(t, fromConsignor, "FORBIDDEN")
}

func TestSettlementReport_FRRPT06_CountsIssuedPaidUnpaidAndOverdue(t *testing.T) {
	scene := issuedSettlement(t)
	scene.store.now.Advance(10 * 24 * time.Hour)
	scene.consignor.post("/api/v1/sessions/refresh")

	report := scene.consignor.get("/api/v1/reports/settlements?from=2026-10-01&to=2026-10-31")

	requireCode(t, report, "OK")
	loadContract(t).requireConforms(t, report)
	if !strings.Contains(string(report.body), `"issued":{"count":1,"amount":31500},"paid":{"count":0,"amount":0},`+
		`"unpaid":{"count":1,"amount":31500},"overdue":{"count":1,"amount":31500}`) {
		t.Fatalf("report = %s", report.body)
	}
}

func TestSalesReport_DSH_DailyChartNarrowsToOneProductOrPartner(t *testing.T) {
	scene := issuedSettlement(t)
	byDay := salesReportPath + "&group_by=DAY"
	unknownID := "0192f0c4-0000-7000-8000-00000000ffff"

	ownProduct := scene.consignor.get(byDay + "&product_id=" + scene.productID)
	otherProduct := scene.consignor.get(byDay + "&product_id=" + unknownID)
	otherPartner := scene.consignor.get(byDay + "&partner_id=" + unknownID)
	malformed := scene.consignor.get(byDay + "&partner_id=toko")

	requireCode(t, ownProduct, "OK")
	loadContract(t).requireConforms(t, ownProduct)
	if !strings.Contains(string(ownProduct.body), `"summary":{"sold_qty":18,`) {
		t.Fatalf("own product = %s", ownProduct.body)
	}
	for _, got := range []exchange{otherProduct, otherPartner} {
		if !strings.Contains(string(got.body), `"summary":{"sold_qty":0,"returned_qty":0,"received_qty":0,`) {
			t.Fatalf("got = %s", got.body)
		}
	}
	requireCode(t, malformed, "VALIDATION_FAILED")
}
