package domain

import (
	"errors"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

var (
	ErrUnknownProduct  = errors.New("product is not an active product of the consignor")
	ErrClaimRaced      = errors.New("invitation was claimed or decided meanwhile")
	ErrAlreadyOpen     = errors.New("an open proposal of this kind already exists for the pair")
	ErrAlreadyDecided  = errors.New("proposal was decided or revised meanwhile")
	ErrAlreadyPartners = errors.New("the pair already has an active partnership")
	ErrQuotaExhausted  = errors.New("the daily proposal quota is used up")
	ErrSenderSuspended = errors.New("the sending organization is suspended")
)

type Party struct {
	OrganizationID string
	Name           string
}

type Revision struct {
	ID           string
	No           int
	AuthorOrgID  string
	AuthorUserID string
	Terms        terms.General
	Message      string
	Items        []Item
	At           time.Time
}

type Proposal struct {
	ID             string
	Number         string
	Kind           Kind
	Status         Status
	Consignor      Party
	Store          Party
	FromOrgID      string
	AwaitingOrgID  string
	InviteCode     string
	PartnershipID  string
	DecisionReason string
	ExpiresAt      time.Time
	CreatedAt      time.Time
	Revision       Revision
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

type PartnershipKey struct {
	OrganizationID string
	PartnershipID  string
}
