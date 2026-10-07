package main

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateMessage(t *testing.T) {
	cases := []struct {
		name    string
		message string
		want    error
	}{
		{"subject only", "✨ Tambah katalog produk penitip\n", nil},
		{"emoji with variation selector", "🔒️ Verifikasi tanda tangan HMAC dari proxy di API\n", nil},
		{"subject with body and reference", "✨ Tambah katalog\n\nAlasan.\n\nRujukan: FR-PRD-01, BR-PRD-03\n", nil},
		{"git comment lines are ignored", "✨ Tambah katalog\n# Please enter the commit message\n", nil},
		{"empty", "\n\n", errEmpty},
		{"no gitmoji", "Tambah katalog produk\n", errSubjectShape},
		{"unknown emoji", "🍕 Tambah katalog produk\n", errSubjectShape},
		{"lowercase sentence", "✨ tambah katalog produk\n", errSubjectShape},
		{"trailing period", "✨ Tambah katalog produk.\n", errSubjectPeriod},
		{"stage number", "✨ Kerjakan fitur 5.1\n", errSubjectNumbering},
		{"too long", "✨ " + strings.Repeat("A", maxSubjectLength) + "\n", errSubjectLength},
		{"body without blank line", "✨ Tambah katalog\nAlasan.\n", errMissingBlankLine},
		{"foreign trailer", "✨ Tambah katalog\n\nSigned-off-by: someone\n", errForeignTrailer},
		{"co-author trailer", "✨ Tambah katalog\n\nCo-Authored-By: someone <a@b.c>\n", errForeignTrailer},
		{"malformed reference", "✨ Tambah katalog\n\nRujukan: FR-PRD-01; BR-PRD-03\n", errReferenceShape},
		{"link in body", "✨ Tambah katalog\n\nLihat https://example.com/sesi\n", errForbiddenContent},
		{"generator note in body", "✨ Tambah katalog\n\nGenerated with a tool\n", errForbiddenContent},
		{"robot emoji in body", "✨ Tambah katalog\n\n\U0001F916 dibuat otomatis\n", errForbiddenContent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateMessage(tc.message); !errors.Is(err, tc.want) {
				t.Fatalf("validateMessage = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestValidateMessage_SubjectAtLimit_IsAccepted(t *testing.T) {
	subject := "✨ T" + strings.Repeat("a", maxSubjectLength-3)
	if err := validateMessage(subject); err != nil {
		t.Fatalf("validateMessage = %v, want nil", err)
	}
}
