package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
)

type Status string

const (
	Active Status = "ACTIVE"
	Ended  Status = "ENDED"
)

func ParseStatus(raw string) (Status, error) {
	switch Status(raw) {
	case "", Active, Ended:
		return Status(raw), nil
	}
	return "", apperror.Validation(apperror.InvalidField("status", apperror.FieldUnknownChoice))
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

type Counterpart struct {
	ID          string
	Type        actor.OrganizationType
	Name        string
	Slug        string
	Description string
	WhatsApp    string
	CityName    string
	District    string
	Address     string
	HasLocation bool
	Exact       Coordinates
	Public      Coordinates
	LogoKey     string
}

type Partnership struct {
	ID          string
	Status      Status
	StartedAt   time.Time
	EndedAt     time.Time
	Counterpart Counterpart
}

type PartnerView struct {
	Counterpart Counterpart
	Location    Coordinates
}

func (p Partnership) View() PartnerView {
	counterpart := p.Counterpart
	if counterpart.Type == actor.Consignor && p.Status != Active {
		counterpart.Address = ""
		return PartnerView{Counterpart: counterpart, Location: counterpart.Public}
	}
	return PartnerView{Counterpart: counterpart, Location: counterpart.Exact}
}

type Key struct {
	OrganizationID string
	PartnershipID  string
}

type Summary struct {
	ID              string
	Status          Status
	StartedAt       time.Time
	EndedAt         time.Time
	CounterpartID   string
	CounterpartName string
	CounterpartType actor.OrganizationType
}

type ListQuery struct {
	OrganizationID string
	Status         Status
	After          pagination.Position
	Limit          int
}

type CatalogItem struct {
	ProductID    string
	Name         string
	SKU          string
	Unit         string
	Description  string
	NetContent   string
	CategoryName string
	PhotoKey     string
	IsAgreed     bool
	IsActive     bool
	Terms        pricing.Terms
}
