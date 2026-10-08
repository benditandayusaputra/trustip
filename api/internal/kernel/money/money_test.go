package money

import (
	"errors"
	"testing"
)

func TestNewRupiah_BRAGR01_WholeRupiahFromZeroToOneHundredMillion(t *testing.T) {
	cases := map[int64]bool{
		-1:          false,
		0:           true,
		1:           true,
		100_000_000: true,
		100_000_001: false,
	}
	for amount, valid := range cases {
		got, err := NewRupiah(amount)

		if valid && (err != nil || got.Int64() != amount) {
			t.Errorf("%d: got %d, err = %v", amount, got, err)
		}
		if !valid && !errors.Is(err, ErrOutOfRange) {
			t.Errorf("%d: err = %v", amount, err)
		}
	}
}

func TestText_SpecTwentyFour_RupiahWithThousandsDots(t *testing.T) {
	cases := map[Rupiah]string{
		0: "Rp0", 750: "Rp750", 1000: "Rp1.000", 691187: "Rp691.187", 100_000_000: "Rp100.000.000",
	}
	for amount, want := range cases {
		if got := amount.Text(); got != want {
			t.Errorf("%d: got %q, want %q", amount, got, want)
		}
	}
}
