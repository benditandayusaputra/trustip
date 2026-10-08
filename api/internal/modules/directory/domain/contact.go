package domain

import "github.com/benditandayusaputra/trustip/api/internal/kernel/actor"

type Contact struct {
	OrganizationID string
	Name           string
	Type           actor.OrganizationType
	CityName       string
	WhatsApp       string
}
