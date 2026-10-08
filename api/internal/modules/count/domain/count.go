package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/canonical"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
)

const (
	EntityType       = "COUNT"
	NumberPrefix     = "HS"
	recordedEvent    = "COUNT_RECORDED"
	saleMovement     = "SALE"
	returnMovement   = "RETURN"
	contentTimestamp = "2006-01-02T15:04:05Z"
	idsPerAllocation = 2
	returnedQtyKey   = "returned_qty"
	storeShareKey    = "store_share_amount"
	consignorKey     = "consignor_share_amount"
)

type Status string

const (
	Recorded  Status = "RECORDED"
	Confirmed Status = "CONFIRMED"
)

type SaleRecord struct {
	ID         string
	MovementID string
	SaleAllocation
}

type ReturnRecord struct {
	ID         string
	MovementID string
	stock.Allocation
}

type CountLine struct {
	Line
	ID      string
	Sold    []SaleRecord
	Taken   []ReturnRecord
	Ordinal int
}

type Count struct {
	ID             string
	Number         string
	PartnershipID  string
	StoreID        string
	ConsignorID    string
	Status         Status
	Occasion       Occasion
	Note           string
	CountedAt      time.Time
	ReceivedAt     time.Time
	RecordedBy     string
	IdempotencyKey string
	Fingerprint    []byte
	Lines          []CountLine
	SoldQty        int
	ReturnedQty    int
	Split          pricing.Split
}

type Recording struct {
	IDs            []string
	Sequence       int
	RecordedBy     string
	IdempotencyKey string
	Fingerprint    []byte
	At             time.Time
}

func (d Draft) IDCount() int {
	total := 1
	for _, line := range d.Tally.Lines {
		total += 1 + idsPerAllocation*(len(line.Sales)+len(line.Returns))
	}
	return total
}

type idQueue struct {
	ids []string
}

func (q *idQueue) next() string {
	id := q.ids[0]
	q.ids = q.ids[1:]
	return id
}

func Number(at time.Time, sequence int) string {
	return fmt.Sprintf("%s-%s-%04d", NumberPrefix, period.MonthKey(at), sequence)
}

func (d Draft) Record(recording Recording) Count {
	queue := &idQueue{ids: recording.IDs}
	count := Count{
		ID: queue.next(), Number: Number(recording.At, recording.Sequence), PartnershipID: d.sheet.PartnershipID,
		StoreID: d.sheet.StoreID, ConsignorID: d.sheet.ConsignorID, Status: Recorded, Occasion: d.Occasion,
		Note: d.Note, CountedAt: d.CountedAt, ReceivedAt: recording.At, RecordedBy: recording.RecordedBy,
		IdempotencyKey: recording.IdempotencyKey, Fingerprint: recording.Fingerprint,
		SoldQty: d.Tally.SoldQty, ReturnedQty: d.Tally.ReturnedQty, Split: d.Tally.Split,
	}
	for index, line := range d.Tally.Lines {
		count.Lines = append(count.Lines, recordLine(line, index+1, queue))
	}
	return count
}

func recordLine(line Line, ordinal int, queue *idQueue) CountLine {
	recorded := CountLine{Line: line, ID: queue.next(), Ordinal: ordinal}
	for _, sale := range line.Sales {
		record := SaleRecord{ID: queue.next(), MovementID: queue.next(), SaleAllocation: sale}
		recorded.Sold = append(recorded.Sold, record)
	}
	for _, taken := range line.Returns {
		record := ReturnRecord{ID: queue.next(), MovementID: queue.next(), Allocation: taken}
		recorded.Taken = append(recorded.Taken, record)
	}
	return recorded
}

type Movement struct {
	ID         string
	ProductID  string
	DropItemID string
	Type       string
	Qty        int
}

func (c Count) Movements() []Movement {
	var movements []Movement
	for _, line := range c.Lines {
		for _, sale := range line.Sold {
			movements = append(movements, Movement{
				ID: sale.MovementID, ProductID: line.ProductID, DropItemID: sale.BatchID, Type: saleMovement,
				Qty: -sale.Qty,
			})
		}
		for _, taken := range line.Taken {
			movements = append(movements, Movement{
				ID: taken.MovementID, ProductID: line.ProductID, DropItemID: taken.BatchID, Type: returnMovement,
				Qty: -taken.Qty,
			})
		}
	}
	return movements
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

func (c Count) Notice(storeName string) Notice {
	return Notice{
		OrganizationID: c.ConsignorID, ActorUserID: c.RecordedBy, Type: recordedEvent, Title: "Hitung Sisa " + c.Number,
		Body: fmt.Sprintf("%s menghitung sisa: %d terjual, %d dikembalikan (%s).",
			storeName, c.SoldQty, c.ReturnedQty, c.Number),
		Link: "/app/hitung/" + c.ID,
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

func (c Count) SealRequest() SealRequest {
	return SealRequest{
		PartnershipID: c.PartnershipID, Event: recordedEvent, DocumentType: EntityType, DocumentID: c.ID,
		DocumentNumber: c.Number, Content: c.Content(), ActorUserID: c.RecordedBy, ActorOrgID: c.StoreID,
		At: c.ReceivedAt,
	}
}

func (c Count) Content() map[string]any {
	lines := make([]any, 0, len(c.Lines))
	for _, line := range c.Lines {
		lines = append(lines, lineContent(line))
	}
	return map[string]any{
		"count_id": c.ID, "number": c.Number, "partnership_id": c.PartnershipID, "store_id": c.StoreID,
		"consignor_id": c.ConsignorID, "occasion": string(c.Occasion), "note": c.Note,
		"counted_at": c.CountedAt.UTC().Format(contentTimestamp), "recorded_by": c.RecordedBy, "lines": lines,
		"sold_qty": c.SoldQty, returnedQtyKey: c.ReturnedQty, "gross_amount": c.Split.Gross.Int64(),
		storeShareKey: c.Split.StoreShare.Int64(), consignorKey: c.Split.ConsignorShare.Int64(),
	}
}

func lineContent(line CountLine) map[string]any {
	sold := slices.SortedFunc(slices.Values(line.Sold), func(left, right SaleRecord) int {
		return strings.Compare(left.ID, right.ID)
	})
	taken := slices.SortedFunc(slices.Values(line.Taken), func(left, right ReturnRecord) int {
		return strings.Compare(left.ID, right.ID)
	})
	sales := make([]any, 0, len(sold))
	for _, sale := range sold {
		sales = append(sales, map[string]any{
			"id": sale.ID, "batch_id": sale.BatchID, "qty": sale.Qty, "gross_amount": sale.Split.Gross.Int64(),
			storeShareKey: sale.Split.StoreShare.Int64(), consignorKey: sale.Split.ConsignorShare.Int64(),
		})
	}
	returns := make([]any, 0, len(taken))
	for _, record := range taken {
		returns = append(returns, map[string]any{"id": record.ID, "batch_id": record.BatchID, "qty": record.Qty})
	}
	return map[string]any{
		"id": line.ID, "product_id": line.ProductID, "recorded_qty": line.RecordedQty, "counted_qty": line.CountedQty,
		"sold_qty": line.SoldQty, returnedQtyKey: line.ReturnedQty, "kept_qty": line.KeptQty,
		"return_reason": string(line.ReturnReason), "sales": sales, "returns": returns,
	}
}

func (c Count) ContentDigest() ([]byte, error) {
	digest, err := canonical.Digest(c.Content())
	if err != nil {
		return nil, fmt.Errorf("digest count %s: %w", c.ID, err)
	}
	return digest, nil
}
