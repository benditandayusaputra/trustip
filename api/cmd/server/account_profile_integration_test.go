package main

import (
	"net/http"
	"strings"
	"testing"
)

const renewedInput = "Baru-sandi-2026"

func accountUpdate(avatar string) string {
	return `{"full_name":"Sari W","whatsapp":"0813 1111 2222","avatar_file_id":` + avatar + `}`
}

func TestAccountProfile_FRAUTH13_MemberSavesNameWhatsAppAndOwnAvatar(t *testing.T) {
	owner := databaseBrowser(t)
	owner.postOrganization(onboarding(semarangCityCode))
	owner.post("/api/v1/sessions/refresh")
	uploaded := owner.upload(part{name: "purpose", data: []byte("AVATAR")},
		part{name: "file", filename: "foto.jpg", data: photoWithLocation(t)})
	if uploaded.status != http.StatusCreated || !strings.Contains(string(uploaded.body), `"visibility":"PRIVATE"`) {
		t.Fatalf("upload = %d %s", uploaded.status, uploaded.body)
	}
	avatarID := quotedAfter(string(uploaded.body), `"id":"`)

	saved := owner.patch("/api/v1/me", accountUpdate(`"`+avatarID+`"`))

	requireCode(t, saved, "OK")
	body := string(saved.body)
	avatarURL := `"/api/v1/files/` + avatarID + `"`
	if !strings.Contains(body, `"whatsapp":"6281311112222"`) || !strings.Contains(body, avatarURL) {
		t.Fatalf("body = %s", body)
	}
	loadContract(t).requireConforms(t, saved)
	read := owner.getAccount()
	requireCode(t, read, "OK")
	loadContract(t).requireConforms(t, read)
	if !strings.Contains(string(read.body), `"full_name":"Sari W"`) {
		t.Fatalf("account = %s", read.body)
	}
}

func TestAccountProfile_FRAUTH13_AvatarOfAnotherUserIsRefused(t *testing.T) {
	owner := databaseBrowser(t)
	owner.postOrganization(onboarding(semarangCityCode))
	owner.post("/api/v1/sessions/refresh")
	uploaded := owner.upload(part{name: "purpose", data: []byte("AVATAR")},
		part{name: "file", filename: "foto.jpg", data: photoWithLocation(t)})
	avatarID := quotedAfter(string(uploaded.body), `"id":"`)
	stranger := owner.anotherUserOnTheSameServer("tini@contoh.id")
	stranger.postOrganization(onboarding(semarangCityCode))
	stranger.post("/api/v1/sessions/refresh")

	got := stranger.patch("/api/v1/me", accountUpdate(`"`+avatarID+`"`))

	requireCode(t, got, "VALIDATION_FAILED")
	if !strings.Contains(string(got.body), `"field":"avatar_file_id"`) {
		t.Fatalf("body = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestChangePassword_FRAUTH14_NeedsTheCurrentPasswordAndEndsOtherDevices(t *testing.T) {
	visitor := databaseBrowser(t)
	phone := &browser{t: t, server: visitor.server, now: visitor.now, visits: 50, pool: visitor.pool}
	if got := phone.send(postSessions(t, attemptAt(credentials("sari@contoh.id", acceptedInput),
		visitor.now.Now()))); got.status != http.StatusOK {
		t.Fatalf("phone login = %d %s", got.status, got.body)
	}

	wrong := visitor.postJSON("/api/v1/me/password",
		`{"current_password":"Bukan-sandinya-1","new_password":"`+renewedInput+`"}`)
	changed := visitor.postJSON("/api/v1/me/password",
		`{"current_password":"`+acceptedInput+`","new_password":"`+renewedInput+`"}`)

	requireCode(t, wrong, "VALIDATION_FAILED")
	requireCode(t, changed, "OK")
	loadContract(t).requireConforms(t, wrong)
	loadContract(t).requireConforms(t, changed)
	requireCode(t, visitor.getAccount(), "OK")
	if got := phone.post("/api/v1/sessions/refresh"); got.status == http.StatusOK {
		t.Fatalf("the other device kept its session: %s", got.body)
	}
	fresh := &browser{t: t, server: visitor.server, now: visitor.now, visits: 80, pool: visitor.pool}
	if got := fresh.send(postSessions(t, attemptAt(credentials("sari@contoh.id", renewedInput),
		visitor.now.Now()))); got.status != http.StatusOK {
		t.Fatalf("login with the new password = %d %s", got.status, got.body)
	}
}
