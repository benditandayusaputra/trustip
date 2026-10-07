package field

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

func TestCheckLength_RSEC48_CountsCharactersNotBytes(t *testing.T) {
	cases := map[string]string{
		"é":   "TOO_SHORT",
		"éé":  "",
		"ééé": "TOO_LONG",
	}
	for value, want := range cases {
		got := Collect(CheckLength("name", value, Length{Min: 2, Max: 2}))

		if (len(got) == 0 && want != "") || (len(got) > 0 && got[0].Code != want) {
			t.Errorf("%q: %+v, want %q", value, got, want)
		}
	}
}

func TestCollect_RSEC47_KeepsEveryProblemInOrder(t *testing.T) {
	got := Collect(
		Invalid("name", apperror.FieldRequired),
		Problem{},
		CheckChoice("unit", "ton", []string{"pcs"}),
	)

	if len(got) != 2 || got[0].Field != "name" || got[1].Code != "UNKNOWN_CHOICE" {
		t.Fatalf("problems = %+v", got)
	}
}

func TestCheckChoice_EmptyMeansNotChosen(t *testing.T) {
	if got := Collect(CheckChoice("store_type", "", []string{"CAFE"})); len(got) != 0 {
		t.Fatalf("problems = %+v", got)
	}
}
