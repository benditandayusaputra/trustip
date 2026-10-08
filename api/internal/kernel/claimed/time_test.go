package claimed

import (
	"testing"
	"time"
)

func nine() time.Time {
	return time.Date(2026, 11, 10, 2, 0, 0, 0, time.UTC)
}

func TestOr_AbsentTimeTakesTheFallback(t *testing.T) {
	if got := Absent().Or(nine()); !got.Equal(nine()) {
		t.Fatalf("Or = %s, want %s", got, nine())
	}
}

func TestOr_GivenTimeWins(t *testing.T) {
	given := nine().Add(-time.Hour)

	if got := At(given).Or(nine()); !got.Equal(given) {
		t.Fatalf("Or = %s, want %s", got, given)
	}
}

func TestText_RRES06_AbsentTimeIsEmptySoARetryKeepsItsFingerprint(t *testing.T) {
	if Absent().Text() != "" {
		t.Fatalf("Text = %q, want empty", Absent().Text())
	}
}

func TestText_GivenTimeIsWrittenInUTC(t *testing.T) {
	jakarta := time.FixedZone("WIB", 7*60*60)

	if got := At(nine().In(jakarta)).Text(); got != "2026-11-10T02:00:00Z" {
		t.Fatalf("Text = %q", got)
	}
}
