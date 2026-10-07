package domain

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const acceptableInput = "Kp7#mQz2vX"

func validForm() RegistrationForm {
	return RegistrationForm{
		FullName:        "Sari Wulandari",
		Email:           "Sari.W@Contoh.ID",
		Password:        acceptableInput,
		WhatsApp:        "0812-3456-7890",
		PrivacyAccepted: true,
	}
}

func problemCodes(t *testing.T, err error) map[string]string {
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

func TestNewRegistration_ValidForm_NormalizesEmailAndWhatsApp(t *testing.T) {
	registration, err := NewRegistration(validForm(), NewPasswordPolicy())
	if err != nil {
		t.Fatalf("new registration: %v", err)
	}

	if registration.Email != "sari.w@contoh.id" || registration.WhatsApp.String() != "6281234567890" {
		t.Fatalf("registration = %+v", registration)
	}
}

func TestNewRegistration_RSEC47_ReportsEveryInvalidFieldAtOnce(t *testing.T) {
	form := RegistrationForm{FullName: "S", Email: "bukan-email", Password: "pendek", WhatsApp: "12"}

	_, err := NewRegistration(form, NewPasswordPolicy())

	want := map[string]string{
		"full_name":        "TOO_SHORT",
		"email":            "INVALID_EMAIL",
		"password":         "TOO_SHORT",
		"whatsapp":         "INVALID_WHATSAPP",
		"privacy_accepted": "CONSENT_REQUIRED",
	}
	got := problemCodes(t, err)
	for field, code := range want {
		if got[field] != code {
			t.Errorf("%s = %q, want %q", field, got[field], code)
		}
	}
}

func TestNewRegistration_FRAUTH08_PrivacyConsentIsRequired(t *testing.T) {
	form := validForm()
	form.PrivacyAccepted = false

	_, err := NewRegistration(form, NewPasswordPolicy())

	if got := problemCodes(t, err); !slices.Equal(fieldsOf(got), []string{"privacy_accepted"}) {
		t.Fatalf("problems = %v", got)
	}
}

func TestNewRegistration_RSEC46_FullNameLengthBoundaries(t *testing.T) {
	cases := []struct {
		name     string
		fullName string
		want     string
	}{
		{"one character", "S", "TOO_SHORT"},
		{"two characters", "Si", ""},
		{"eighty characters", strings.Repeat("ā", 80), ""},
		{"eighty one characters", strings.Repeat("ā", 81), "TOO_LONG"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			form := validForm()
			form.FullName = tc.fullName

			_, err := NewRegistration(form, NewPasswordPolicy())

			if tc.want == "" && err != nil {
				t.Fatalf("err = %v", err)
			}
			if tc.want != "" && problemCodes(t, err)["full_name"] != tc.want {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestNewRegistration_EmailShapes(t *testing.T) {
	cases := map[string]bool{
		"sari@contoh.id":                    true,
		"sari+toko@contoh.co.id":            true,
		"sari@localhost":                    false,
		"Sari <sari@contoh.id>":             false,
		"sari@@contoh.id":                   false,
		"":                                  false,
		strings.Repeat("a", 248) + "@ab.id": true,
		strings.Repeat("a", 249) + "@ab.id": false,
	}
	for email, valid := range cases {
		form := validForm()
		form.Email = email

		_, err := NewRegistration(form, NewPasswordPolicy())

		if (err == nil) != valid {
			t.Errorf("%q: err = %v, want valid %v", email, err, valid)
		}
	}
}

func TestNewRegistration_PasswordMayNotContainTheGivenName(t *testing.T) {
	form := validForm()
	form.Password = "wulandari#2026"

	_, err := NewRegistration(form, NewPasswordPolicy())

	if problemCodes(t, err)["password"] != "CONTAINS_IDENTITY" {
		t.Fatalf("err = %v", err)
	}
}

func TestRegistration_Complete_RecordsConsentTime(t *testing.T) {
	registration, err := NewRegistration(validForm(), NewPasswordPolicy())
	if err != nil {
		t.Fatalf("new registration: %v", err)
	}
	now := time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
	password := StoredPassword{Hash: "$argon2id$x", PepperVersion: "v1"}

	user := registration.Complete("0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e", password, now)

	if user.PrivacyAcceptedAt != now || user.Password != password || user.Email != "sari.w@contoh.id" {
		t.Fatalf("user = %+v", user)
	}
}

func fieldsOf(codes map[string]string) []string {
	fields := make([]string, 0, len(codes))
	for field := range codes {
		fields = append(fields, field)
	}
	slices.Sort(fields)
	return fields
}
