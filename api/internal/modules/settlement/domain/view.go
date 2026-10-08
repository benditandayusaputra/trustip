package domain

import (
	"errors"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
)

var ErrAlreadyLinked = errors.New("a sale allocation is already linked to another settlement")

type SequenceKey struct {
	OrganizationID string
	Prefix         string
	Period         string
}

func SequenceOf(consignorID string, at time.Time) SequenceKey {
	return SequenceKey{OrganizationID: consignorID, Prefix: NumberPrefix, Period: period.MonthKey(at)}
}

func ParseDueDate(raw string) (time.Time, error) {
	day, err := time.ParseInLocation(dateLayout, raw, period.Jakarta())
	if err != nil {
		return time.Time{}, errors.Join(errors.New("read due date"), err)
	}
	return day, nil
}

type SettlementKey struct {
	OrganizationID string
	SettlementID   string
}

type ProductInfo struct {
	Name string
	SKU  string
	Unit string
}

type SourceCount struct {
	ID        string
	Number    string
	CountedAt time.Time
}

type View struct {
	Settlement
	ConsignorName string
	StoreName     string
	PaidAt        time.Time
	Products      map[string]ProductInfo
	Counts        []SourceCount
}

func (v View) StandingAt(now time.Time) Standing {
	return StandingOf(v.Status, v.DueDate, now)
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
	Standing Standing
	Seal     SealVerdict
}

type RunningKey struct {
	OrganizationID string
	PartnershipID  string
}

type Running struct {
	PartnershipID string
	HasSales      bool
	Since         time.Time
	SoldQty       int
	Split         pricing.Split
}
