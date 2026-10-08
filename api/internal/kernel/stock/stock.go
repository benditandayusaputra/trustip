package stock

import (
	"errors"
	"slices"
	"strings"
	"time"
)

var (
	ErrInsufficientStock = errors.New("quantity exceeds the stock of the given batches")
	ErrNegativeQuantity  = errors.New("quantity must not be negative")
)

type Movement struct {
	BatchID string
	Qty     int
}

type Batch struct {
	ID         string
	Expiry     Expiry
	ReceivedAt time.Time
	Balance    int
}

type Allocation struct {
	BatchID string
	Qty     int
}

func Balances(movements []Movement) map[string]int {
	balances := make(map[string]int, len(movements))
	for _, movement := range movements {
		balances[movement.BatchID] += movement.Qty
	}
	return balances
}

func Allocate(batches []Batch, qty int) ([]Allocation, error) {
	if qty < 0 {
		return nil, ErrNegativeQuantity
	}
	if qty > totalBalance(batches) {
		return nil, ErrInsufficientStock
	}
	allocations := []Allocation{}
	unallocated := qty
	for _, batch := range firstExpiryFirst(batches) {
		taken := min(unallocated, batch.Balance)
		if taken > 0 {
			allocations = append(allocations, Allocation{BatchID: batch.ID, Qty: taken})
			unallocated -= taken
		}
	}
	return allocations, nil
}

func Deduct(batches []Batch, allocations []Allocation) []Batch {
	taken := make(map[string]int, len(allocations))
	for _, allocation := range allocations {
		taken[allocation.BatchID] += allocation.Qty
	}
	remaining := slices.Clone(batches)
	for i := range remaining {
		remaining[i].Balance -= taken[remaining[i].ID]
	}
	return remaining
}

func totalBalance(batches []Batch) int {
	total := 0
	for _, batch := range batches {
		total += max(batch.Balance, 0)
	}
	return total
}

func firstExpiryFirst(batches []Batch) []Batch {
	ordered := slices.Clone(batches)
	slices.SortStableFunc(ordered, compareLeavingOrder)
	return ordered
}

func compareLeavingOrder(left, right Batch) int {
	if byExpiry := left.Expiry.compare(right.Expiry); byExpiry != 0 {
		return byExpiry
	}
	if byDrop := left.ReceivedAt.Compare(right.ReceivedAt); byDrop != 0 {
		return byDrop
	}
	return strings.Compare(left.ID, right.ID)
}
