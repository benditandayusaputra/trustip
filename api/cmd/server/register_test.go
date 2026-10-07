package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

type registerEnvelope struct {
	Code   string `json:"code"`
	Errors []struct {
		Field   string `json:"field"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	Data struct {
		Email string `json:"email"`
	} `json:"data"`
}

func decodeRegister(t *testing.T, body []byte) registerEnvelope {
	t.Helper()
	var envelope registerEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return envelope
}

func TestRegister_FRAUTH01_CreatesAccountAndPointsToIt(t *testing.T) {
	server := newTestServer(alwaysReady)

	got := perform(t, server, postUsers(t, registration("  Sari@Contoh.ID ")))

	if got.status != http.StatusCreated || got.header.Get("Location") != "/api/v1/me" {
		t.Fatalf("response = %d %v %s", got.status, got.header, got.body)
	}
	if email := decodeRegister(t, got.body).Data.Email; email != "sari@contoh.id" {
		t.Fatalf("email = %q", email)
	}
}

func TestRegister_SameEmailTwice_TellsTheUserToSignIn(t *testing.T) {
	server := newTestServer(alwaysReady)
	first := perform(t, server, postUsers(t, registration("sari@contoh.id")))
	if first.status != http.StatusCreated {
		t.Fatalf("first = %d %s", first.status, first.body)
	}

	got := perform(t, server, postUsers(t, registration("SARI@contoh.id")))

	envelope := decodeRegister(t, got.body)
	if got.status != http.StatusUnprocessableEntity || len(envelope.Errors) != 1 ||
		envelope.Errors[0].Message != "Email sudah terdaftar. Silakan masuk." {
		t.Fatalf("response = %d %s", got.status, got.body)
	}
}

func TestRegister_RSEC51_ControlCharactersInNameAreRejected(t *testing.T) {
	body := `{"full_name":"Sari\u0000","email":"sari@contoh.id","password":"Kp7#mQz2vX",` +
		`"whatsapp":"081234567890","privacy_accepted":true,"captcha_token":"` + solvedCaptcha + `"}`

	got := perform(t, newTestServer(alwaysReady), postUsers(t, body))

	envelope := decodeRegister(t, got.body)
	if got.status != http.StatusUnprocessableEntity || envelope.Errors[0].Code != "INVALID_CHARACTERS" {
		t.Fatalf("response = %d %s", got.status, got.body)
	}
}

func TestRegister_FourthSignupFromOneIPInAnHour_IsRateLimited(t *testing.T) {
	server := newTestServer(alwaysReady)
	for i, email := range []string{"a@contoh.id", "b@contoh.id", "c@contoh.id"} {
		if got := perform(t, server, postUsers(t, registration(email))); got.status != 201 {
			t.Fatalf("signup %d = %d %s", i+1, got.status, got.body)
		}
	}

	got := perform(t, server, postUsers(t, registration("d@contoh.id")))

	if got.status != http.StatusTooManyRequests || got.header.Get("Retry-After") == "" {
		t.Fatalf("response = %d %s", got.status, got.body)
	}
}
