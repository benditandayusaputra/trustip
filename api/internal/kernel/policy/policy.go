package policy

import (
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

type Action string

const (
	EditProfile        Action = "EDIT_PROFILE"
	ManageMembers      Action = "MANAGE_MEMBERS"
	ManageProducts     Action = "MANAGE_PRODUCTS"
	ManageNeeds        Action = "MANAGE_NEEDS"
	NegotiateProposals Action = "NEGOTIATE_PROPOSALS"
	AmendAgreements    Action = "AMEND_AGREEMENTS"
	RecordDrops        Action = "RECORD_DROPS"
	AnnounceShipments  Action = "ANNOUNCE_SHIPMENTS"
	ConfirmDocuments   Action = "CONFIRM_DOCUMENTS"
	RecordCounts       Action = "RECORD_COUNTS"
	ViewDocumentValues Action = "VIEW_DOCUMENT_VALUES"
	RecordPayments     Action = "RECORD_PAYMENTS"
	ReviewPayments     Action = "REVIEW_PAYMENTS"
	ViewReports        Action = "VIEW_REPORTS"
	EndPartnerships    Action = "END_PARTNERSHIPS"
	ReportListings     Action = "REPORT_LISTINGS"
	EditOwnAccount     Action = "EDIT_OWN_ACCOUNT"
)

type Grant int

const (
	Denied Grant = iota
	Allowed
	DelegatedByOwner
)

type Grants struct {
	ConsignorOwner Grant
	ConsignorStaff Grant
	StoreOwner     Grant
	StoreStaff     Grant
}

func Table() map[Action]Grants {
	return map[Action]Grants{
		EditProfile:        ownersOnly(),
		ManageMembers:      ownersOnly(),
		ManageProducts:     {ConsignorOwner: Allowed, ConsignorStaff: Allowed},
		ManageNeeds:        {StoreOwner: Allowed},
		NegotiateProposals: ownersOnly(),
		AmendAgreements:    ownersOnly(),
		RecordDrops:        {StoreOwner: Allowed, StoreStaff: Allowed},
		AnnounceShipments:  {ConsignorOwner: Allowed, ConsignorStaff: Allowed},
		ConfirmDocuments:   {ConsignorOwner: Allowed, ConsignorStaff: Allowed},
		RecordCounts:       {StoreOwner: Allowed, StoreStaff: Allowed},
		ViewDocumentValues: everyone(),
		RecordPayments:     {StoreOwner: Allowed, StoreStaff: DelegatedByOwner},
		ReviewPayments:     {ConsignorOwner: Allowed},
		ViewReports:        ownersOnly(),
		EndPartnerships:    ownersOnly(),
		ReportListings:     everyone(),
		EditOwnAccount:     everyone(),
	}
}

func ownersOnly() Grants {
	return Grants{ConsignorOwner: Allowed, StoreOwner: Allowed}
}

func everyone() Grants {
	return Grants{ConsignorOwner: Allowed, ConsignorStaff: Allowed, StoreOwner: Allowed, StoreStaff: Allowed}
}

func Require(signedIn actor.Actor, action Action) error {
	if !signedIn.IsMember() || Table()[action].For(signedIn.Membership) != Allowed {
		return apperror.New(apperror.CodeForbidden)
	}
	return nil
}

type seat struct {
	organizationType actor.OrganizationType
	role             actor.Role
}

func (g Grants) For(membership actor.Membership) Grant {
	seats := map[seat]Grant{
		{actor.Consignor, actor.Owner}: g.ConsignorOwner,
		{actor.Consignor, actor.Staff}: g.ConsignorStaff,
		{actor.Store, actor.Owner}:     g.StoreOwner,
		{actor.Store, actor.Staff}:     g.StoreStaff,
	}
	return seats[seat{membership.OrganizationType, membership.Role}]
}

func RequireNewcomer(signedIn actor.Actor) error {
	if signedIn.IsMember() {
		return apperror.New(apperror.CodeOrganizationExists)
	}
	return nil
}

func RequireMember(signedIn actor.Actor) error {
	if !signedIn.IsMember() {
		return apperror.New(apperror.CodeForbidden)
	}
	return nil
}

func RequireVisitorOrMember(viewer actor.Actor) error {
	if viewer.UserID != "" && !viewer.IsMember() {
		return apperror.New(apperror.CodeForbidden)
	}
	return nil
}
