package domain

import (
	"errors"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

var (
	ErrUnknownProduct = errors.New("product is not an active product of the consignor")
	ErrClaimRaced     = errors.New("invitation was claimed or decided meanwhile")
	ErrAlreadyOpen    = errors.New("an open proposal of this kind already exists for the pair")
)

type Party struct {
	OrganizationID string
	Name           string
}

type Revision struct {
	ID          string
	No          int
	AuthorOrgID string
	Terms       terms.General
	Message     string
	Items       []Item
	At          time.Time
}

type Proposal struct {
	ID            string
	Number        string
	Kind          Kind
	Status        Status
	Consignor     Party
	Store         Party
	FromOrgID     string
	AwaitingOrgID string
	InviteCode    string
	ExpiresAt     time.Time
	CreatedAt     time.Time
	Revision      Revision
}

type ProposalKey struct {
	OrganizationID string
	ProposalID     string
}

type Pair struct {
	ConsignorID string
	StoreID     string
}

func (c Claim) Pair() Pair {
	return Pair{ConsignorID: c.ConsignorID, StoreID: c.StoreID}
}

type SequenceKey struct {
	OrganizationID string
	Prefix         string
	Period         string
}
