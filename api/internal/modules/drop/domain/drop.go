package domain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/canonical"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
)

const (
	EntityType       = "DROP"
	NumberPrefix     = "TM"
	noticeType       = "DROP_RECORDED"
	recordedEvent    = "DROP_RECORDED"
	receiveMovement  = "RECEIVE"
	contentTimestamp = "2006-01-02T15:04:05Z"
)

type Status string

const (
	Recorded  Status = "RECORDED"
	Confirmed Status = "CONFIRMED"
	Cancelled Status = "CANCELLED"
)

type Item struct {
	ID         string
	MovementID string
	ProductID  string
	BatchCode  string
	Expiry     stock.Expiry
	Qty        int
	Terms      pricing.Terms
}

type Drop struct {
	ID               string
	Number           string
	PartnershipID    string
	AgreementID      string
	AgreementVersion int
	StoreID          string
	ConsignorID      string
	Status           Status
	OccurredAt       time.Time
	ReceivedAt       time.Time
	RecordedBy       string
	IdempotencyKey   string
	Fingerprint      []byte
	Items            []Item
}

func (d Drop) TotalQty() int {
	total := 0
	for _, item := range d.Items {
		total += item.Qty
	}
	return total
}

type Movement struct {
	ID            string
	PartnershipID string
	ProductID     string
	DropItemID    string
	Type          string
	Qty           int
	OccurredAt    time.Time
	SourceType    string
	SourceID      string
	SourceNumber  string
	CreatedBy     string
}

func (d Drop) Movements() []Movement {
	movements := make([]Movement, 0, len(d.Items))
	for _, item := range d.Items {
		movements = append(movements, Movement{
			ID: item.MovementID, PartnershipID: d.PartnershipID, ProductID: item.ProductID, DropItemID: item.ID,
			Type: receiveMovement, Qty: item.Qty, OccurredAt: d.OccurredAt, SourceType: EntityType, SourceID: d.ID,
			SourceNumber: d.Number, CreatedBy: d.RecordedBy,
		})
	}
	return movements
}

type Notice struct {
	OrganizationID  string
	RecipientUserID string
	ActorUserID     string
	Type            string
	Title           string
	Body            string
	Link            string
	OwnersOnly      bool
}

func (d Drop) Notice(storeName string) Notice {
	return Notice{
		OrganizationID: d.ConsignorID, Type: noticeType, Title: "Titipan masuk " + d.Number,
		Body: fmt.Sprintf("%s mencatat %d unit titipan masuk (%s).", storeName, d.TotalQty(), d.Number),
		Link: "/app/titipan/" + d.ID, ActorUserID: d.RecordedBy,
	}
}

type SealRequest struct {
	PartnershipID  string
	Event          string
	DocumentType   string
	DocumentID     string
	DocumentNumber string
	Content        map[string]any
	ActorUserID    string
	ActorOrgID     string
	At             time.Time
}

func (d Drop) SealRequest() SealRequest {
	return SealRequest{
		PartnershipID: d.PartnershipID, Event: recordedEvent, DocumentType: EntityType, DocumentID: d.ID,
		DocumentNumber: d.Number, Content: d.Content(), ActorUserID: d.RecordedBy, ActorOrgID: d.StoreID,
		At: d.ReceivedAt,
	}
}

func (d Drop) Content() map[string]any {
	items := make([]any, 0, len(d.Items))
	for _, item := range d.Items {
		items = append(items, map[string]any{
			"id": item.ID, "product_id": item.ProductID, "batch_code": item.BatchCode,
			"expiry_date": expiryText(item.Expiry), "qty": item.Qty, "pricing_model": string(item.Terms.Model),
			"retail_price": item.Terms.RetailPrice.Int64(), "commission_bps": item.Terms.CommissionBPS,
			"deposit_price": item.Terms.DepositPrice.Int64(),
		})
	}
	return map[string]any{
		"drop_id": d.ID, "number": d.Number, "partnership_id": d.PartnershipID, "agreement_id": d.AgreementID,
		"agreement_version": d.AgreementVersion, "store_id": d.StoreID, "consignor_id": d.ConsignorID,
		"occurred_at": d.OccurredAt.UTC().Format(contentTimestamp), "recorded_by": d.RecordedBy, "items": items,
	}
}

func (d Drop) ContentDigest() ([]byte, error) {
	digest, err := canonical.Digest(d.Content())
	if err != nil {
		return nil, fmt.Errorf("digest drop %s: %w", d.ID, err)
	}
	return digest, nil
}

func expiryText(expiry stock.Expiry) string {
	date, isGiven := expiry.Date()
	if !isGiven {
		return ""
	}
	return date.Format(time.DateOnly)
}

func batchCode(number string, line int) string {
	return number + "/" + strconv.Itoa(line)
}
