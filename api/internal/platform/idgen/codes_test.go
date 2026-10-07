package idgen

import (
	"strings"
	"testing"
)

func TestCodes_BRPRP07_RandomCharactersFromTheGivenAlphabet(t *testing.T) {
	codes := NewCodes("AB23", 10)
	seen := map[string]bool{}

	for range 20 {
		code, err := codes.NewCode()
		if err != nil || len(code) != 10 || strings.Trim(code, "AB23") != "" {
			t.Fatalf("code = %q, err = %v", code, err)
		}
		seen[code] = true
	}

	if len(seen) < 19 {
		t.Fatalf("only %d distinct codes out of 20", len(seen))
	}
}
