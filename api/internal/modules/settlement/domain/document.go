package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/canonical"
)

const (
	issuedEvent   = "SETTLEMENT_ISSUED"
	issuedNotice  = "SETTLEMENT_ISSUED"
	dateLayout    = time.DateOnly
	consignorKey  = "consignor_share_amount"
	storeShareKey = "store_share_amount"
	grossKey      = "gross_amount"
)

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

func (s Settlement) SealRequest() SealRequest {
	return SealRequest{
		PartnershipID: s.PartnershipID, Event: issuedEvent, DocumentType: EntityType, DocumentID: s.ID,
		DocumentNumber: s.Number, Content: s.Content(), At: s.IssuedAt,
	}
}

func (s Settlement) Content() map[string]any {
	sorted := slices.SortedFunc(slices.Values(s.Allocations), func(left, right Allocation) int {
		return strings.Compare(left.ID, right.ID)
	})
	allocations := make([]any, 0, len(sorted))
	for _, allocation := range sorted {
		allocations = append(allocations, map[string]any{
			"id": allocation.ID, "product_id": allocation.ProductID, "qty": allocation.Qty,
			grossKey: allocation.Split.Gross.Int64(), storeShareKey: allocation.Split.StoreShare.Int64(),
			consignorKey: allocation.Split.ConsignorShare.Int64(),
		})
	}
	return map[string]any{
		"settlement_id": s.ID, "number": s.Number, "partnership_id": s.PartnershipID, "consignor_id": s.ConsignorID,
		"store_id": s.StoreID, "schedule": string(s.Schedule), "period_start": instant(s.Period.Start),
		"period_end": instant(s.Period.End), "due_date": s.DueDate.Format(dateLayout), "sold_qty": s.Summary.Qty,
		grossKey: s.Summary.Split.Gross.Int64(), storeShareKey: s.Summary.Split.StoreShare.Int64(),
		consignorKey: s.Summary.Split.ConsignorShare.Int64(), "allocations": allocations,
	}
}

func instant(at time.Time) string {
	return at.UTC().Format(time.RFC3339Nano)
}

func (s Settlement) ContentDigest() ([]byte, error) {
	digest, err := canonical.Digest(s.Content())
	if err != nil {
		return nil, fmt.Errorf("digest settlement %s: %w", s.ID, err)
	}
	return digest, nil
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

func (s Settlement) Notices(storeName string) []Notice {
	body := fmt.Sprintf("Setoran %s dari %s terbit. Jatuh tempo %s.", s.Number, storeName, longDate(s.DueDate))
	notice := Notice{
		Type: issuedNotice, Title: "Setoran " + s.Number, Body: body, Link: "/app/setoran/" + s.ID, OwnersOnly: true,
	}
	toConsignor, toStore := notice, notice
	toConsignor.OrganizationID = s.ConsignorID
	toStore.OrganizationID = s.StoreID
	return []Notice{toConsignor, toStore}
}

func longDate(day time.Time) string {
	months := []string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	local := day.In(startOfDay(day).Location())
	return fmt.Sprintf("%d %s %d", local.Day(), months[local.Month()-1], local.Year())
}
