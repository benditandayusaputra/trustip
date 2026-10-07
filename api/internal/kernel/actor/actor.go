package actor

type OrganizationType string

const (
	Consignor OrganizationType = "CONSIGNOR"
	Store     OrganizationType = "STORE"
)

type Role string

const (
	Owner Role = "OWNER"
	Staff Role = "STAFF"
)

type Membership struct {
	OrganizationID   string
	OrganizationType OrganizationType
	Role             Role
}

type Actor struct {
	UserID        string
	SessionFamily string
	Membership    Membership
}

func (a Actor) IsMember() bool {
	return a.Membership.OrganizationID != ""
}
