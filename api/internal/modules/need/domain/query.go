package domain

import (
	"errors"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
)

var (
	ErrUnknownCategory  = errors.New("category does not exist")
	ErrOpenLimit        = errors.New("the store already has five open needs")
	ErrChangedMeanwhile = errors.New("need status changed meanwhile")
)

type Response struct {
	ProposalID    string
	Number        string
	Status        string
	FromOrgID     string
	AwaitingOrgID string
	ConsignorName string
	CreatedAt     time.Time
}

type Detail struct {
	Need      Need
	Responses []Response
}

func ParseStatus(raw string) (Status, error) {
	switch Status(raw) {
	case "", Open, Fulfilled, Closed, Expired:
		return Status(raw), nil
	}
	return "", apperror.Validation(apperror.InvalidField("status", apperror.FieldUnknownChoice))
}

type ListQuery struct {
	StoreID string
	Status  Status
	After   pagination.Position
	Limit   int
}
