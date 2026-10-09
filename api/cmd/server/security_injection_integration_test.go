package main

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"
)

func injectionPayloads() []string {
	return []string{
		`' OR '1'='1`, `'; DROP TABLE organizations; --`, `%`, `_`, `\`, `1) OR (1=1`,
		`00000000-0000-0000-0000-000000000000' --`, `<script>alert(1)</script>`, `2026-01-01'::date --`,
	}
}

func searchParameters() map[string][]string {
	directory := []string{
		"origin", "lat", "lng", "city_code", "radius_km", "sort", "category_id", "q", "cursor", "limit",
	}
	return map[string][]string{
		"/api/v1/cities":                {"province_code"},
		"/api/v1/notifications":         {"cursor", "limit"},
		"/api/v1/directory/stores":      slices.Concat(directory, []string{"store_type", "max_commission_bps"}),
		"/api/v1/directory/consignors":  slices.Concat(directory, []string{"min_price", "min_shelf_life_days"}),
		"/api/v1/directory/products":    slices.Concat(directory, []string{"consignor_id", "min_price", "has_halal"}),
		"/api/v1/directory/needs":       directory,
		"/api/v1/products":              {"q", "category_id", "status", "cursor", "limit"},
		"/api/v1/proposals":             {"box", "cursor", "limit"},
		"/api/v1/partnerships":          {"status", "cursor", "limit"},
		"/api/v1/drops":                 {"partnership_id", "status", "from", "to", "cursor", "limit"},
		"/api/v1/counts":                {"partnership_id", "from", "to", "cursor", "limit"},
		"/api/v1/settlements":           {"partnership_id", "status", "age", "cursor", "limit"},
		"/api/v1/reports/sales":         {"from", "to", "group_by", "partner_id", "product_id"},
		"/api/v1/reports/daily":         {"date"},
		"/api/v1/reports/settlements":   {"from", "to"},
		"/api/v1/public/products/x/y":   {"lat", "lng"},
		"/api/v1/directory/needs?q=abc": {"category_id"},
	}
}

func (b *browser) probeWithInjections(path, name string) {
	b.t.Helper()
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	for index, payload := range injectionPayloads() {
		b.visits = 10 + index
		got := b.get(path + separator + name + "=" + url.QueryEscape(payload))
		if got.status >= http.StatusInternalServerError {
			b.t.Fatalf("GET %s %s=%q = %d %s", path, name, payload, got.status, got.body)
		}
	}
}

func TestSecurity_RSEC61_InjectionPayloadsInSearchAndFiltersAreHarmless(t *testing.T) {
	consignor, store := listedStoreScene(t)
	for path, names := range searchParameters() {
		for _, name := range names {
			consignor.probeWithInjections(path, name)
			consignor.now.Advance(time.Second)
		}
	}

	nearby := consignor.get(directoryStoresPath + "?origin=HERE&lat=-7.0456&lng=110.4386&radius_km=2")
	wildcard := consignor.get(directoryStoresPath + "?origin=HERE&lat=-7.0456&lng=110.4386&radius_km=2&q=%25")

	if !strings.Contains(string(nearby.body), store.id) || strings.Contains(string(wildcard.body), store.id) {
		t.Fatalf("nearby = %s\nwildcard = %s", nearby.body, wildcard.body)
	}
}
