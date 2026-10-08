package domain

import (
	"errors"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

var (
	ErrDuplicateKey = errors.New("a count with this idempotency key already exists")
	ErrNotPending   = errors.New("the count is no longer waiting for a response")
)

type SequenceKey struct {
	OrganizationID string
	Prefix         string
	Period         string
}

func SequenceOf(storeID string, at time.Time) SequenceKey {
	return SequenceKey{OrganizationID: storeID, Prefix: NumberPrefix, Period: period.MonthKey(at)}
}

type SheetKey struct {
	StoreID       string
	PartnershipID string
}

type IdempotencyKey struct {
	StoreID string
	Key     string
}

type Stored struct {
	CountID     string
	Fingerprint []byte
}

type CountKey struct {
	OrganizationID string
	CountID        string
}

type LineProduct struct {
	Name string
	SKU  string
	Unit string
}

type SettlementRef struct {
	ID     string
	Number string
}

type View struct {
	Count
	StoreName       string
	ConsignorName   string
	RecordedByName  string
	ConfirmedByName string
	ConfirmedAt     time.Time
	IsConfirmed     bool
	AutoConfirmed   bool
	RecordedOffline bool
	Settlement      SettlementRef
	Products        map[string]LineProduct
}

func (v View) SealCheck() SealCheck {
	return SealCheck{DocumentType: EntityType, DocumentID: v.ID, Content: v.Content()}
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
	View
	Seal SealVerdict
}
