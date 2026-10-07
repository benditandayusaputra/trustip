package domain

import "errors"

var (
	ErrSlugTaken     = errors.New("organization slug is already taken")
	ErrUnknownCity   = errors.New("city code is not in the city list")
	ErrAlreadyMember = errors.New("user already belongs to an organization")
	ErrUnknownLogo   = errors.New("logo is not a logo file of this organization")
)
