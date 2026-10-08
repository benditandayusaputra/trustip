package domain

import "time"

type Opening struct {
	Kind          Kind
	IDs           InvitationIDs
	Number        string
	Sender        Sender
	ConsignorID   string
	StoreID       string
	AwaitingID    string
	PartnershipID string
	NeedID        string
	SenderName    string
	Draft         RevisionDraft
	At            time.Time
	ExpiresAt     time.Time
}

func (i Invitation) Opening() Opening {
	return Opening{
		Kind: New, IDs: i.IDs, Number: i.Number, Sender: i.Sender, ConsignorID: i.ConsignorID(),
		StoreID: i.StoreID(), Draft: i.Draft, At: i.At, ExpiresAt: i.ExpiresAt,
	}
}
