package pagination_test

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
)

func TestParseLimit_RRES04_DefaultsToTwentyAndCapsAtFifty(t *testing.T) {
	cases := map[string]int{"": 20, "1": 1, "50": 50, "51": 50, "0": 20, "-3": 20, "abc": 20}
	for raw, want := range cases {
		if got := pagination.ParseLimit(raw); got != want {
			t.Errorf("ParseLimit(%q) = %d, want %d", raw, got, want)
		}
	}
}

func TestCursor_RoundTripsThePositionOfTheLastItem(t *testing.T) {
	position := pagination.Position{
		At: time.Date(2026, 10, 19, 8, 0, 0, 123456000, time.UTC), ID: "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e",
	}

	parsed, err := pagination.ParseCursor(pagination.CursorOf(position))

	if err != nil || !parsed.At.Equal(position.At) || parsed.ID != position.ID {
		t.Fatalf("parsed = %+v, err = %v", parsed, err)
	}
}

func TestParseCursor_EmptyCursorStartsAtTheTop(t *testing.T) {
	position, err := pagination.ParseCursor("")

	if err != nil || !position.IsStart() {
		t.Fatalf("position = %+v, err = %v", position, err)
	}
}

func TestParseCursor_TamperedCursorIsAFieldError(t *testing.T) {
	for _, raw := range []string{"!!!", "bm90LWEtY3Vyc29y", "MjAyNi0xMC0xOXxub3QtYW4taWQ"} {
		_, err := pagination.ParseCursor(raw)

		fields := apperror.From(err).Fields()
		if len(fields) != 1 || fields[0].Field != "cursor" {
			t.Errorf("%q: err = %v", raw, err)
		}
	}
}
