package contact_test

import (
	"errors"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/contact"
)

func TestParseWhatsAppNumber_RSEC52_NormalizesTo62(t *testing.T) {
	cases := map[string]string{
		"0812-3456-7890":     "6281234567890",
		"+62 812 3456 7890":  "6281234567890",
		"62.812.3456.7890":   "6281234567890",
		"(0812) 3456 7890":   "6281234567890",
		"6281234567890":      "6281234567890",
		"81234567890":        "6281234567890",
		"081234567":          "6281234567",
		"62 8123 4567 89012": "628123456789012",
	}
	for raw, want := range cases {
		t.Run(raw, func(t *testing.T) {
			got, err := contact.ParseWhatsAppNumber(raw)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got.String() != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestParseWhatsAppNumber_DigitCountBoundaries(t *testing.T) {
	cases := []struct {
		name  string
		raw   string
		valid bool
	}{
		{"nine digits", "628123456", false},
		{"ten digits", "6281234567", true},
		{"fifteen digits", "628123456789012", true},
		{"sixteen digits", "6281234567890123", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := contact.ParseWhatsAppNumber(tc.raw)
			if (err == nil) != tc.valid {
				t.Fatalf("err = %v, want valid %v", err, tc.valid)
			}
		})
	}
}

func TestParseWhatsAppNumber_RejectsWhatIsNotAnIndonesianNumber(t *testing.T) {
	for _, raw := range []string{"", "0812-ABCD-7890", "+1 415 555 0100", "12345678901", "62-0812-3456"} {
		if _, err := contact.ParseWhatsAppNumber(raw); !errors.Is(err, contact.ErrInvalidWhatsAppNumber) {
			t.Errorf("%q: err = %v", raw, err)
		}
	}
}
