package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	ConfirmWindow  = 48 * time.Hour
	confirmedEvent = "COUNT_CONFIRMED"
)

type Decider struct {
	UserID         string
	OrganizationID string
}

type Decision struct {
	CountID string
	By      string
	ByOrgID string
	At      time.Time
	Event   string
}

func (v View) Effective(now time.Time) View {
	deadline := v.ReceivedAt.Add(ConfirmWindow)
	if v.Status != Recorded || now.Before(deadline) {
		return v
	}
	v.Status = Confirmed
	v.IsConfirmed = true
	v.AutoConfirmed = true
	v.ConfirmedAt = deadline
	return v
}

func (v View) Confirm(decider Decider, decidedAt time.Time) (Decision, error) {
	if decider.OrganizationID != v.ConsignorID {
		return Decision{}, apperror.New(apperror.CodeForbidden)
	}
	if v.Effective(decidedAt).Status != Recorded {
		return Decision{}, apperror.New(apperror.CodeDocumentNotPending)
	}
	return Decision{
		CountID: v.ID, By: decider.UserID, ByOrgID: decider.OrganizationID, At: decidedAt, Event: confirmedEvent,
	}, nil
}

func (v View) SealRequestOf(decision Decision) SealRequest {
	request := v.SealRequest()
	request.Event = decision.Event
	request.ActorUserID = decision.By
	request.ActorOrgID = decision.ByOrgID
	request.At = decision.At
	return request
}

type Deadline struct {
	Cutoff time.Time
	Window time.Duration
	At     time.Time
}

func DeadlineAt(now time.Time) Deadline {
	return Deadline{Cutoff: now.Add(-ConfirmWindow), Window: ConfirmWindow, At: now}
}

type SettlementApproval struct {
	SettlementID string
	By           string
	At           time.Time
}
