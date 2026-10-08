package domain

import "github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"

type PartnershipKey struct {
	OrganizationID string
	PartnershipID  string
}

type CardKey struct {
	PartnershipKey
	ProductID string
}

type CardQuery struct {
	CardKey
	After pagination.Position
	Limit int
}

type CardProduct struct {
	ID   string
	Name string
	SKU  string
	Unit string
}
