package domain

import "errors"

var (
	ErrSlugTaken       = errors.New("organization slug is already taken")
	ErrUnknownCity     = errors.New("city code is not in the city list")
	ErrAlreadyMember   = errors.New("user already belongs to an organization")
	ErrUnknownLogo     = errors.New("logo is not a logo file of this organization")
	ErrUnknownCategory = errors.New("category is not in the category list")
	ErrUnknownPhoto    = errors.New("photo is not a store photo of this organization")
	ErrUnknownQR       = errors.New("qr is not a payment qr of this organization")
)
