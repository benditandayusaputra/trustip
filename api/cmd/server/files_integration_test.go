package main

import (
	"net/http"
	"testing"
)

func TestReadFile_RSEC41_FileOfAnotherOrganizationIsNotFound(t *testing.T) {
	owner := databaseBrowser(t)
	owner.postOrganization(onboarding(semarangCityCode))
	owner.post("/api/v1/sessions/refresh")
	uploaded := owner.upload(part{name: "purpose", data: []byte("LOGO")},
		part{name: "file", filename: "logo.jpg", data: photoWithLocation(t)})
	if uploaded.status != http.StatusCreated {
		t.Fatalf("upload = %d %s", uploaded.status, uploaded.body)
	}
	fileID := quotedAfter(string(uploaded.body), `"id":"`)
	stranger := owner.anotherUserOnTheSameServer("tini@contoh.id")
	stranger.postOrganization(onboarding(semarangCityCode))
	stranger.post("/api/v1/sessions/refresh")

	got := stranger.get("/api/v1/files/" + fileID)

	requireCode(t, got, "NOT_FOUND")
	if own := owner.get("/api/v1/files/" + fileID); own.status != http.StatusOK {
		t.Fatalf("owner = %d %s", own.status, own.body)
	}
}
