package idgen_test

import (
	"regexp"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/platform/idgen"
)

var uuidV7 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestUUIDv7_NewID_IsVersionSevenAndUnique(t *testing.T) {
	generator := idgen.UUIDv7{}

	first, errFirst := generator.NewID()
	second, errSecond := generator.NewID()

	if errFirst != nil || errSecond != nil {
		t.Fatalf("NewID errors: %v, %v", errFirst, errSecond)
	}
	if !uuidV7.MatchString(first) || first == second {
		t.Fatalf("ids = %s, %s", first, second)
	}
}

func TestSequence_NewID_IsPredictable(t *testing.T) {
	sequence := idgen.NewSequence()

	first, _ := sequence.NewID()
	second, _ := sequence.NewID()

	if first != "00000000-0000-7000-8000-000000000001" || second != "00000000-0000-7000-8000-000000000002" {
		t.Fatalf("ids = %s, %s", first, second)
	}
	if !uuidV7.MatchString(first) {
		t.Fatalf("sequence id %s is not shaped like a UUIDv7", first)
	}
}
