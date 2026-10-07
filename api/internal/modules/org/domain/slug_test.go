package domain

import (
	"strings"
	"testing"
)

func TestSlugFrom_BRORG02_KeepsLowercaseLettersDigitsAndHyphens(t *testing.T) {
	cases := map[string]Slug{
		"Warung Bu Tini":           "warung-bu-tini",
		"  Kripik  Mbok   Darmi  ": "kripik-mbok-darmi",
		"Toko Roti & Kue 99":       "toko-roti-kue-99",
		"Café Kopi Lèmbang":        "cafe-kopi-lembang",
		"Sambal_Bu--Ani!!":         "sambal-bu-ani",
		"Oleh-oleh Khas Solo":      "oleh-oleh-khas-solo",
	}
	for name, want := range cases {
		if got := SlugFrom(name); got != want {
			t.Errorf("SlugFrom(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestSlugFrom_NameWithoutLettersOrDigits_FallsBackToAFixedWord(t *testing.T) {
	if got := SlugFrom("!!! ---"); got != "usaha" {
		t.Fatalf("slug = %q", got)
	}
}

func TestSlugFrom_LongName_IsCutAtAWordBoundaryWithinTheLimit(t *testing.T) {
	name := strings.Repeat("panjang ", 20)

	got := SlugFrom(name)

	if len(got) > MaxSlugLength || strings.HasSuffix(string(got), "-") ||
		!strings.HasPrefix(string(got), "panjang-") {
		t.Fatalf("slug = %q (%d)", got, len(got))
	}
}

func TestNextFreeSlug_BRORG02_FreeBaseIsUsedAsIs(t *testing.T) {
	if got := NextFreeSlug("warung-bu-tini", nil); got != "warung-bu-tini" {
		t.Fatalf("slug = %q", got)
	}
}

func TestNextFreeSlug_BRORG02_TakenBaseGetsSuffixTwo(t *testing.T) {
	got := NextFreeSlug("warung-bu-tini", []Slug{"warung-bu-tini"})

	if got != "warung-bu-tini-2" {
		t.Fatalf("slug = %q", got)
	}
}

func TestNextFreeSlug_BRORG02_SuffixContinuesAfterTheHighestTaken(t *testing.T) {
	taken := []Slug{"warung-bu-tini", "warung-bu-tini-2", "warung-bu-tini-3"}

	if got := NextFreeSlug("warung-bu-tini", taken); got != "warung-bu-tini-4" {
		t.Fatalf("slug = %q", got)
	}
}

func TestNextFreeSlug_BRORG02_GapsAreFilledFromTheLowestNumber(t *testing.T) {
	taken := []Slug{"warung-bu-tini", "warung-bu-tini-3"}

	if got := NextFreeSlug("warung-bu-tini", taken); got != "warung-bu-tini-2" {
		t.Fatalf("slug = %q", got)
	}
}
