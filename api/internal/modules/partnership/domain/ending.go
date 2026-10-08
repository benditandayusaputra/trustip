package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
)

const (
	PendingWindow      = 48 * time.Hour
	MaxEndReasonLength = 500
	EntityType         = "PARTNERSHIP"
	endedEvent         = "PARTNERSHIP_ENDED"
	endedAtLayout      = "2006-01-02T15:04:05Z"
)

type BlockerKind string

const (
	StockAtStore     BlockerKind = "STOCK_AT_STORE"
	PendingDrop      BlockerKind = "PENDING_DROP"
	PendingCount     BlockerKind = "PENDING_COUNT"
	PendingPayment   BlockerKind = "PENDING_PAYMENT"
	UnpaidSettlement BlockerKind = "UNPAID_SETTLEMENT"
)

type Blocker struct {
	Kind           BlockerKind
	DocumentID     string
	DocumentNumber string
	Qty            int
	Amount         int64
}

func (b Blocker) Describe() string {
	descriptions := map[BlockerKind]string{
		StockAtStore:     fmt.Sprintf("stok %d unit masih di toko", b.Qty),
		PendingDrop:      fmt.Sprintf("Titipan Masuk %s menunggu konfirmasi", b.DocumentNumber),
		PendingCount:     fmt.Sprintf("Hitung Sisa %s menunggu konfirmasi", b.DocumentNumber),
		PendingPayment:   fmt.Sprintf("pembayaran Setoran %s menunggu konfirmasi", b.DocumentNumber),
		UnpaidSettlement: fmt.Sprintf("Setoran %s belum lunas", b.DocumentNumber),
	}
	return descriptions[b.Kind]
}

type EndState struct {
	PartnershipID string
	Status        Status
	ConsignorID   string
	StoreID       string
	ConsignorName string
	StoreName     string
	Blockers      []Blocker
}

type Ender struct {
	UserID         string
	OrganizationID string
}

type Ending struct {
	PartnershipID string
	ConsignorID   string
	StoreID       string
	EnderName     string
	EndedBy       string
	EndedByOrgID  string
	Reason        string
	At            time.Time
}

func (s EndState) End(ender Ender, reason string, endedAt time.Time) (Ending, error) {
	if s.Status != Active {
		return Ending{}, apperror.New(apperror.CodePartnershipNotActive)
	}
	problems := field.Collect(field.CheckLength("reason", reason, field.Length{Max: MaxEndReasonLength}))
	if len(problems) > 0 {
		return Ending{}, apperror.Validation(problems...)
	}
	if len(s.Blockers) > 0 {
		return Ending{}, apperror.New(apperror.CodePartnershipBlocked).
			WithParam("daftar penghalang", s.describeBlockers())
	}
	enderName := s.StoreName
	if ender.OrganizationID == s.ConsignorID {
		enderName = s.ConsignorName
	}
	return Ending{
		PartnershipID: s.PartnershipID, ConsignorID: s.ConsignorID, StoreID: s.StoreID, EnderName: enderName,
		EndedBy: ender.UserID, EndedByOrgID: ender.OrganizationID, Reason: reason, At: endedAt,
	}, nil
}

func (s EndState) describeBlockers() string {
	descriptions := make([]string, 0, len(s.Blockers))
	for _, blocker := range s.Blockers {
		descriptions = append(descriptions, blocker.Describe())
	}
	return strings.Join(descriptions, ", ")
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

func (e Ending) SealRequest() SealRequest {
	return SealRequest{
		PartnershipID: e.PartnershipID, Event: endedEvent, DocumentType: EntityType, DocumentID: e.PartnershipID,
		DocumentNumber: "", ActorUserID: e.EndedBy, ActorOrgID: e.EndedByOrgID, At: e.At,
		Content: map[string]any{
			"partnership_id": e.PartnershipID, "consignor_id": e.ConsignorID, "store_id": e.StoreID,
			"ended_by_org_id": e.EndedByOrgID, "end_reason": e.Reason,
			"ended_at": e.At.UTC().Format(endedAtLayout),
		},
	}
}

type Notice struct {
	OrganizationID string
	ActorUserID    string
	Type           string
	Title          string
	Body           string
	Link           string
	OwnersOnly     bool
}

func (e Ending) Notice() Notice {
	counterpart := e.ConsignorID
	if e.EndedByOrgID == e.ConsignorID {
		counterpart = e.StoreID
	}
	return Notice{
		OrganizationID: counterpart, ActorUserID: e.EndedBy, Type: endedEvent, Title: "Kemitraan berakhir",
		Body: e.EnderName + " mengakhiri kemitraan.", Link: "/app/mitra/" + e.PartnershipID, OwnersOnly: true,
	}
}

type EndKey struct {
	Key
	ConfirmCutoff time.Time
}

type ListingChange struct {
	StoreID       string
	PartnershipID string
	ShowOnPublic  bool
	At            time.Time
}
