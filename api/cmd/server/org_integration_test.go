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

func TestUpdateProfile_FRORG01_LogoUploadedThenLinkedShowsInTheSettings(t *testing.T) {
	visitor := databaseBrowser(t)
	founded := visitor.postOrganization(onboarding(semarangCityCode))
	organizationID := quotedAfter(string(founded.body), `"id":"`)
	visitor.post("/api/v1/sessions/refresh")
	uploaded := visitor.upload(part{name: "purpose", data: []byte("LOGO")},
		part{name: "file", filename: "logo.jpg", data: photoWithLocation(t)})
	logoID := quotedAfter(string(uploaded.body), `"id":"`)

	saved := visitor.patch(organizationsPath+"/"+organizationID,
		strings.Replace(profileUpdate(""), `"logo_file_id":null`, `"logo_file_id":"`+logoID+`"`, 1))
	read := visitor.get(organizationsPath + "/" + organizationID)

	requireCode(t, saved, "OK")
	if !strings.Contains(string(read.body), `"logo":{"id":"`+logoID+`","url":"`+testFilesBaseURL+"/") ||
		!strings.Contains(string(read.body), `"name":"Warung Bu Tini Jaya"`) {
		t.Fatalf("read = %s", read.body)
	}
	loadContract(t).requireConforms(t, read)
}

func TestReadSettings_RSEC41_AnotherOrganizationIsNotFoundInTheDatabase(t *testing.T) {
	owner := databaseBrowser(t)
	founded := owner.postOrganization(onboarding(semarangCityCode))
	organizationID := quotedAfter(string(founded.body), `"id":"`)
	stranger := owner.anotherUserOnTheSameServer("tini@contoh.id")
	stranger.postOrganization(onboarding(semarangCityCode))
	stranger.post("/api/v1/sessions/refresh")

	read := stranger.get(organizationsPath + "/" + organizationID)
	changed := stranger.patch(organizationsPath+"/"+organizationID, profileUpdate(""))

	requireCode(t, read, "NOT_FOUND")
	requireCode(t, changed, "NOT_FOUND")
}
