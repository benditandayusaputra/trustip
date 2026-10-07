package audit

import "time"

type Action string

const (
	SignedIn                Action = "SIGNED_IN"
	OrganizationCreated     Action = "ORGANIZATION_CREATED"
	ProfileUpdated          Action = "PROFILE_UPDATED"
	StoreProfileUpdated     Action = "STORE_PROFILE_UPDATED"
	ConsignorProfileUpdated Action = "CONSIGNOR_PROFILE_UPDATED"
	PaymentInfoUpdated      Action = "PAYMENT_INFO_UPDATED"
	ListingChanged          Action = "LISTING_CHANGED"
)

type EntityType string

const (
	UserEntity         EntityType = "USER"
	OrganizationEntity EntityType = "ORGANIZATION"
)

type Entry struct {
	ID             string
	OrganizationID string
	ActorID        string
	Action         Action
	EntityType     EntityType
	EntityID       string
	Metadata       map[string]string
	IP             string
	At             time.Time
}
