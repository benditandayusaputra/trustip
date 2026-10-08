package money

import (
	"errors"
	"strconv"
	"strings"
)

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

const thousandsGroup = 3

func (r Rupiah) Text() string {
	digits := strconv.FormatInt(int64(r), 10)
	var grouped strings.Builder
	for index, digit := range digits {
		if index > 0 && (len(digits)-index)%thousandsGroup == 0 {
			grouped.WriteByte('.')
		}
		grouped.WriteRune(digit)
	}
	return "Rp" + grouped.String()
}
