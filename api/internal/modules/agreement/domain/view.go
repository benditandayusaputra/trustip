package domain

import "time"

type Party struct {
	OrganizationID string
	Name           string
}

type Person struct {
	Name             string
	OrganizationName string
}

type View struct {
	Agreement Agreement
	Consignor Party
	Store     Party
	Proposer  Person
	Acceptor  Person
}

type PartnershipKey struct {
	OrganizationID string
	PartnershipID  string
}

type SealCheck struct {
	DocumentType string
	DocumentID   string
	Content      map[string]any
}

type SealEvent struct {
	Event         string
	ActorOrgName  string
	ActorUserName string
	At            time.Time
	Code          string
}

type SealVerdict struct {
	IsValid bool
	Events  []SealEvent
}

type SealedView struct {
	View View
	Seal SealVerdict
}

func (v View) SealCheck() SealCheck {
	return SealCheck{DocumentType: EntityType, DocumentID: v.Agreement.ID, Content: v.Agreement.Content()}
}
