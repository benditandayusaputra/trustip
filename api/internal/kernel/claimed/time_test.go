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

func TestSettledBy_BRPWA07_DeviceTimeUpToFiveMinutesAheadIsKept(t *testing.T) {
	ahead := nine().Add(ClockTolerance)

	if got := At(ahead).SettledBy(nine()); !got.Equal(ahead) {
		t.Fatalf("SettledBy = %s, want %s", got, ahead)
	}
}

func TestSettledBy_BRPWA07_DeviceTimeMoreThanFiveMinutesAheadBecomesTheReceipt(t *testing.T) {
	ahead := nine().Add(ClockTolerance + time.Second)

	if got := At(ahead).SettledBy(nine()); !got.Equal(nine()) {
		t.Fatalf("SettledBy = %s, want %s", got, nine())
	}
}

func TestSettledBy_BRPWA07_EarlierDeviceTimeIsKept(t *testing.T) {
	earlier := nine().Add(-48 * time.Hour)

	if got := At(earlier).SettledBy(nine()); !got.Equal(earlier) {
		t.Fatalf("SettledBy = %s, want %s", got, earlier)
	}
}

func TestSettledBy_AbsentDeviceTimeIsTheReceipt(t *testing.T) {
	if got := Absent().SettledBy(nine()); !got.Equal(nine()) {
		t.Fatalf("SettledBy = %s, want %s", got, nine())
	}
}

func TestParse_MissingTextIsAbsent(t *testing.T) {
	parsed, isValid := Parse(nil)

	if !isValid || parsed.Text() != "" {
		t.Fatalf("Parse(nil) = %+v, %t", parsed, isValid)
	}
}

func TestParse_RFC3339TextIsGiven(t *testing.T) {
	raw := "2026-11-10T09:00:00+07:00"

	parsed, isValid := Parse(&raw)

	if !isValid || parsed.Text() != "2026-11-10T02:00:00Z" {
		t.Fatalf("Parse = %+v, %t", parsed, isValid)
	}
}

func TestParse_OtherTextIsRefused(t *testing.T) {
	raw := "kemarin"

	if _, isValid := Parse(&raw); isValid {
		t.Fatal("Parse accepted text that is not RFC 3339")
	}
}
