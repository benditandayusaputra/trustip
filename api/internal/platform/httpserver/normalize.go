package httpserver

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func NormalizeLine(raw string) (string, bool) {
	if strings.IndexFunc(raw, isForbiddenInLine) >= 0 {
		return "", false
	}
	return strings.Join(strings.Fields(norm.NFC.String(raw)), " "), true
}

func NormalizePassword(raw string) string {
	return norm.NFKC.String(raw)
}

func isForbiddenInLine(r rune) bool {
	return unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r)
}
