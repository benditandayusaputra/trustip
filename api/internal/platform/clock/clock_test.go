package clock_test

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
)

func TestSystem_Now_IsUTC(t *testing.T) {
	if location := (clock.System{}).Now().Location(); location != time.UTC {
		t.Fatalf("location = %s, want UTC", location)
	}
}

func TestManual_Advance_MovesTimeForward(t *testing.T) {
	start := time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
	manual := clock.NewManual(start)

	manual.Advance(48 * time.Hour)

	if got := manual.Now(); !got.Equal(start.Add(48 * time.Hour)) {
		t.Fatalf("now = %s", got)
	}
}
