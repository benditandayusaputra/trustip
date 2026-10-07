package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
)

type Box string

const (
	Incoming Box = "INCOMING"
	Outgoing Box = "OUTGOING"
)

func ParseBox(raw string) (Box, error) {
	switch Box(raw) {
	case "", Incoming:
		return Incoming, nil
	case Outgoing:
		return Outgoing, nil
	}
	return "", apperror.Validation(apperror.InvalidField("box", apperror.FieldUnknownChoice))
}

type ListQuery struct {
	OrganizationID string
	Box            Box
	After          pagination.Position
	Limit          int
}

type Summary struct {
	ID              string
	Number          string
	Kind            Kind
	Status          Status
	FromOrgID       string
	AwaitingOrgID   string
	CounterpartName string
	ExpiresAt       time.Time
	CreatedAt       time.Time
}
