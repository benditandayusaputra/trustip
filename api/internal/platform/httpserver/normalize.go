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

func NormalizeMultiline(raw string) (string, bool) {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		value, clean := NormalizeLine(line)
		if !clean {
			return "", false
		}
		normalized = append(normalized, value)
	}
	return strings.Trim(strings.Join(normalized, "\n"), "\n"), true
}
