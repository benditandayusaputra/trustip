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
