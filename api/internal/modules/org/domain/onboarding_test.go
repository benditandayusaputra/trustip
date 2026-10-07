package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

func validForm() OnboardingForm {
	return OnboardingForm{
		Type:        "STORE",
		Name:        "Warung Bu Tini",
		Description: "Warung kelontong dekat SD.",
		WhatsApp:    "0812-3456-7890",
		CityCode:    "33.74",
		District:    "Tembalang",
		Address:     "Jl. Sirojudin 12",
	}
}

func fieldCodes(t *testing.T, err error) map[string]string {
	t.Helper()
	appErr := apperror.From(err)
	if appErr.Code() != apperror.CodeValidationFailed {
		t.Fatalf("err = %v, want validation failure", err)
	}
	codes := map[string]string{}
	for _, field := range appErr.Fields() {
		codes[field.Field] = field.Code
	}
	return codes
}

func TestNewOnboarding_FRAUTH04_ValidFormNormalizesWhatsApp(t *testing.T) {
	onboarding, err := NewOnboarding(validForm())
	if err != nil {
		t.Fatalf("new onboarding: %v", err)
	}

	if onboarding.Type != actor.Store || onboarding.WhatsApp.String() != "6281234567890" ||
		onboarding.CityCode != "33.74" {
		t.Fatalf("onboarding = %+v", onboarding)
	}
}

func TestNewOnboarding_RSEC47_ReportsEveryInvalidFieldAtOnce(t *testing.T) {
	form := OnboardingForm{Type: "ADMIN", Name: "W", WhatsApp: "12", CityCode: "semarang"}

	_, err := NewOnboarding(form)

	want := map[string]string{
		"type":      "UNKNOWN_CHOICE",
		"name":      "TOO_SHORT",
		"whatsapp":  "INVALID_WHATSAPP",
		"city_code": "UNKNOWN_CHOICE",
	}
	got := fieldCodes(t, err)
	for field, code := range want {
		if got[field] != code {
			t.Errorf("%s = %q, want %q (all: %v)", field, got[field], code, got)
		}
	}
}

func TestNewOnboarding_RequiredCityIsReportedAsRequired(t *testing.T) {
	form := validForm()
	form.CityCode = ""

	_, err := NewOnboarding(form)

	if got := fieldCodes(t, err)["city_code"]; got != "REQUIRED" {
		t.Fatalf("city_code = %q", got)
	}
}

func TestNewOnboarding_RSEC46_NameLengthBoundaries(t *testing.T) {
	cases := map[int]string{1: "TOO_SHORT", 2: "", 80: "", 81: "TOO_LONG"}
	for length, want := range cases {
		form := validForm()
		form.Name = strings.Repeat("é", length)

		_, err := NewOnboarding(form)

		if got := codeOrEmpty(t, err, "name"); got != want {
			t.Errorf("length %d: %q, want %q", length, got, want)
		}
	}
}

func TestNewOnboarding_RSEC46_OptionalTextLengthBoundaries(t *testing.T) {
	limits := map[string]int{"description": 1000, "district": 80, "address": 200}
	for field, limit := range limits {
		for length, want := range map[int]string{limit: "", limit + 1: "TOO_LONG"} {
			form := validForm()
			setText(&form, field, strings.Repeat("a", length))

			_, err := NewOnboarding(form)

			if got := codeOrEmpty(t, err, field); got != want {
				t.Errorf("%s length %d: %q, want %q", field, length, got, want)
			}
		}
	}
}

func TestNewOnboarding_OptionalTextMayBeEmpty(t *testing.T) {
	form := validForm()
	form.Description, form.District, form.Address = "", "", ""

	if _, err := NewOnboarding(form); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func setText(form *OnboardingForm, field, value string) {
	switch field {
	case "description":
		form.Description = value
	case "district":
		form.District = value
	case "address":
		form.Address = value
	}
}

func codeOrEmpty(t *testing.T, err error, field string) string {
	t.Helper()
	if err == nil {
		return ""
	}
	return fieldCodes(t, err)[field]
}

func TestFound_BRAUTH02_FounderBecomesTheOwner(t *testing.T) {
	onboarding, err := NewOnboarding(validForm())
	if err != nil {
		t.Fatalf("new onboarding: %v", err)
	}
	now := time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)

	founding := onboarding.Found(FoundingIDs{
		OrganizationID: "org-1", MembershipID: "membership-1", FounderID: "user-1",
	}, "warung-bu-tini", now)

	if founding.Organization.ID != "org-1" || founding.Organization.Slug != "warung-bu-tini" ||
		!founding.Organization.CreatedAt.Equal(now) {
		t.Fatalf("organization = %+v", founding.Organization)
	}
	want := Membership{ID: "membership-1", UserID: "user-1", Role: actor.Owner}
	if founding.Owner != want {
		t.Fatalf("owner = %+v", founding.Owner)
	}
}
