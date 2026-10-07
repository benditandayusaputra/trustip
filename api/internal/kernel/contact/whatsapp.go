package contact

import (
	"errors"
	"regexp"
	"strings"
)

const (
	countryCode   = "62"
	trunkPrefix   = "0"
	mobilePrefix  = "8"
	formatMarkers = " -.()+"
)

var (
	ErrInvalidWhatsAppNumber = errors.New("whatsapp number must be an Indonesian number of 10 to 15 digits")
	normalizedWhatsApp       = regexp.MustCompile(`^62[1-9][0-9]{7,12}$`)
)

type WhatsAppNumber struct {
	digits string
}

func ParseWhatsAppNumber(raw string) (WhatsAppNumber, error) {
	digits := strings.Map(func(r rune) rune {
		if strings.ContainsRune(formatMarkers, r) {
			return -1
		}
		return r
	}, raw)
	normalized := withCountryCode(digits)
	if !normalizedWhatsApp.MatchString(normalized) {
		return WhatsAppNumber{}, ErrInvalidWhatsAppNumber
	}
	return WhatsAppNumber{digits: normalized}, nil
}

func withCountryCode(digits string) string {
	switch {
	case strings.HasPrefix(digits, trunkPrefix):
		return countryCode + strings.TrimPrefix(digits, trunkPrefix)
	case strings.HasPrefix(digits, mobilePrefix):
		return countryCode + digits
	}
	return digits
}

func (w WhatsAppNumber) String() string {
	return w.digits
}
