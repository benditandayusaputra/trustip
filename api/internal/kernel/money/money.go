package money

import "errors"

const MaxRupiah = 100_000_000

var ErrOutOfRange = errors.New("rupiah amount is outside 0 to 100,000,000")

type Rupiah int64

func NewRupiah(amount int64) (Rupiah, error) {
	if amount < 0 || amount > MaxRupiah {
		return 0, ErrOutOfRange
	}
	return Rupiah(amount), nil
}

func (r Rupiah) Int64() int64 {
	return int64(r)
}
