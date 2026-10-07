package idgen

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

type Codes struct {
	alphabet string
	length   int
}

func NewCodes(alphabet string, length int) Codes {
	return Codes{alphabet: alphabet, length: length}
}

func (c Codes) NewCode() (string, error) {
	var code strings.Builder
	limit := big.NewInt(int64(len(c.alphabet)))
	for range c.length {
		index, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("draw code character: %w", err)
		}
		code.WriteByte(c.alphabet[index.Int64()])
	}
	return code.String(), nil
}
