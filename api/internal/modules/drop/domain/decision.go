package domain

import (
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	ConfirmWindow     = 48 * time.Hour
	CancelWindow      = 15 * time.Minute
	confirmedEvent    = "DROP_CONFIRMED"
	cancelledEvent    = "DROP_CANCELLED"
	adjustmentType    = "ADJUSTMENT"
	confirmNoticeType = "DROP_CONFIRMED"
)

type Decider struct {
	UserID         string
	OrganizationID string
}

type Decision struct {
	DropID   string
	Status   Status
	By       string
	ByOrgID  string
	At       time.Time
	Event    string
	Notice   Notice
	reversed []Item
	drop     Drop
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
		DropID: v.ID, Status: Confirmed, By: decider.UserID, ByOrgID: decider.OrganizationID, At: decidedAt,
		Event: confirmedEvent, drop: v.Drop,
		Notice: Notice{
			OrganizationID: v.StoreID, RecipientUserID: v.RecordedBy, ActorUserID: decider.UserID,
			Type: confirmNoticeType, Title: "Titipan masuk " + v.Number,
			Body: fmt.Sprintf("%s menyatakan %s sesuai.", v.ConsignorName, v.Number), Link: "/app/titipan/" + v.ID,
		},
	}, nil
}

type Cancellation struct {
	Decider     Decider
	At          time.Time
	IsStockUsed bool
}

func (v View) Cancel(cancellation Cancellation) (Decision, error) {
	if cancellation.Decider.OrganizationID != v.StoreID {
		return Decision{}, apperror.New(apperror.CodeForbidden)
	}
	if v.Effective(cancellation.At).Status != Recorded {
		return Decision{}, apperror.New(apperror.CodeDocumentNotPending)
	}
	if cancellation.At.After(v.ReceivedAt.Add(CancelWindow)) {
		return Decision{}, apperror.New(apperror.CodeCancelWindowPassed)
	}
	if cancellation.IsStockUsed {
		return Decision{}, apperror.New(apperror.CodeDropAlreadyCounted)
	}
	return Decision{
		DropID: v.ID, Status: Cancelled, By: cancellation.Decider.UserID,
		ByOrgID: cancellation.Decider.OrganizationID, At: cancellation.At, Event: cancelledEvent,
		reversed: v.Items, drop: v.Drop,
	}, nil
}

func (d Decision) Reversal(ids []string) []Movement {
	movements := make([]Movement, 0, len(d.reversed))
	for index, item := range d.reversed {
		movements = append(movements, Movement{
			ID: ids[index], PartnershipID: d.drop.PartnershipID, ProductID: item.ProductID, DropItemID: item.ID,
			Type: adjustmentType, Qty: -item.Qty, OccurredAt: d.At, SourceType: EntityType, SourceID: d.drop.ID,
			SourceNumber: d.drop.Number, CreatedBy: d.By,
		})
	}
	return movements
}

func (d Decision) ReversedLines() int {
	return len(d.reversed)
}

func (v View) SealRequestOf(decision Decision) SealRequest {
	request := v.SealRequest()
	request.Event = decision.Event
	request.ActorUserID = decision.By
	request.ActorOrgID = decision.ByOrgID
	request.At = decision.At
	return request
}
