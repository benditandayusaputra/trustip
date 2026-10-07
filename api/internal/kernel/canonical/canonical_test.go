package canonical

import (
	"errors"
	"testing"
)

func TestEncode_RSEC28_SortedKeysWithoutWhitespace(t *testing.T) {
	got, err := Encode(map[string]any{
		"number": "TM-202610-0007", "items": []any{map[string]any{"qty": 20, "product_id": "p1"}},
		"amount": int64(31500),
	})

	want := `{"amount":31500,"items":[{"product_id":"p1","qty":20}],"number":"TM-202610-0007"}`
	if err != nil || string(got) != want {
		t.Fatalf("got = %s, err = %v", got, err)
	}
}

func TestEncode_RSEC28_EscapesLikeRFC8785(t *testing.T) {
	got, err := Encode(map[string]any{"note": "Roti <3 \"Keju\"\n\t\\ é   \x01"})

	want := `{"note":"Roti <3 \"Keju\"\n\t\\ é ` + " " + ` \u0001"}`
	if err != nil || string(got) != want {
		t.Fatalf("got = %s, want %s, err = %v", got, want, err)
	}
}

func TestEncode_RSEC28_OnlyStringsAndIntegers(t *testing.T) {
	for _, value := range []any{1.5, true, nil, map[string]any{"x": []any{float32(1)}}} {
		if _, err := Encode(value); !errors.Is(err, ErrUnsupported) {
			t.Errorf("%v: err = %v", value, err)
		}
	}
}

func TestEncode_RSEC28_KeysSortByUTF16CodeUnits(t *testing.T) {
	got, err := Encode(map[string]any{"\U0001F600": "emoji", "�": "replacement", "b": "b"})

	want := `{"b":"b","` + "\U0001F600" + `":"emoji","` + "�" + `":"replacement"}`
	if err != nil || string(got) != want {
		t.Fatalf("got = %s, want %s", got, want)
	}
}

func TestDigest_RSEC28_SHA256OfTheCanonicalForm(t *testing.T) {
	digest, err := Digest(map[string]any{})

	if err != nil || len(digest) != 32 ||
		digest[0] != 0x44 || digest[1] != 0x13 {
		t.Fatalf("digest = %x, err = %v", digest, err)
	}
}
