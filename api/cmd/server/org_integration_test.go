package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestOnboarding_BRAUTH02_RefreshedSessionCarriesTheNewMembership(t *testing.T) {
	visitor := databaseBrowser(t)
	before := visitor.getAccount()

	founded := visitor.postOrganization(onboarding(semarangCityCode))
	visitor.post("/api/v1/sessions/refresh")
	after := visitor.getAccount()

	if founded.status != http.StatusCreated {
		t.Fatalf("founded = %d %s", founded.status, founded.body)
	}
	if !strings.Contains(string(before.body), `"membership":null`) {
		t.Fatalf("before = %s", before.body)
	}
	if !strings.Contains(string(after.body), `"organization_type":"STORE","role":"OWNER"`) {
		t.Fatalf("after = %s", after.body)
	}
	loadContract(t).requireConforms(t, after)
}

func TestOnboarding_SecondOrganizationForTheSameUserIsAConflict(t *testing.T) {
	visitor := databaseBrowser(t)
	visitor.postOrganization(onboarding(semarangCityCode))
	visitor.post("/api/v1/sessions/refresh")

	got := visitor.postOrganization(onboarding(semarangCityCode))

	requireCode(t, got, "ORGANIZATION_EXISTS")
	loadContract(t).requireConforms(t, got)
}

func TestOnboarding_BRORG02_SameNameGetsTheNextSlug(t *testing.T) {
	first := databaseBrowser(t)
	first.postOrganization(onboarding(semarangCityCode))
	second := first.anotherUserOnTheSameServer("tini@contoh.id")

	got := second.postOrganization(onboarding(semarangCityCode))

	if !strings.Contains(string(got.body), `"slug":"warung-bu-tini-2"`) {
		t.Fatalf("response = %d %s", got.status, got.body)
	}
}
