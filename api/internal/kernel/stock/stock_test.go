package stock

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type vectorBatch struct {
	ID         string    `json:"id"`
	ExpiryDate *string   `json:"expiry_date"`
	ReceivedAt time.Time `json:"received_at"`
	Balance    int       `json:"balance"`
}

type vectorMovement struct {
	BatchID string `json:"batch_id"`
	Qty     int    `json:"qty"`
}

type vectorAllocation struct {
	BatchID string `json:"batch_id"`
	Qty     int    `json:"qty"`
}

type vectors struct {
	Balances []struct {
		Name      string           `json:"name"`
		Movements []vectorMovement `json:"movements"`
		Balances  map[string]int   `json:"balances"`
	} `json:"balances"`
	Allocations []struct {
		Name        string             `json:"name"`
		Batches     []vectorBatch      `json:"batches"`
		Qty         int                `json:"qty"`
		Allocations []vectorAllocation `json:"allocations"`
		Remaining   map[string]int     `json:"remaining"`
	} `json:"allocations"`
	Shortages []struct {
		Name    string        `json:"name"`
		Batches []vectorBatch `json:"batches"`
		Qty     int           `json:"qty"`
	} `json:"shortages"`
	Flags []struct {
		Name          string  `json:"name"`
		ExpiryDate    *string `json:"expiry_date"`
		ShelfLifeDays *int    `json:"shelf_life_days"`
		Today         string  `json:"today"`
		Flag          string  `json:"flag"`
	} `json:"flags"`
}

func loadVectors(t *testing.T) vectors {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "vectors", "stock.json"))
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}
	var loaded vectors
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatalf("decode vectors: %v", err)
	}
	return loaded
}

func dateOf(t *testing.T, raw string) time.Time {
	t.Helper()
	date, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		t.Fatalf("date %q: %v", raw, err)
	}
	return date
}

func expiryOf(t *testing.T, raw *string) Expiry {
	t.Helper()
	if raw == nil {
		return NoExpiry()
	}
	return ExpiresOn(dateOf(t, *raw))
}

func batchesOf(t *testing.T, raw []vectorBatch) []Batch {
	t.Helper()
	batches := make([]Batch, 0, len(raw))
	for _, item := range raw {
		batches = append(batches, Batch{
			ID: item.ID, Expiry: expiryOf(t, item.ExpiryDate), ReceivedAt: item.ReceivedAt, Balance: item.Balance,
		})
	}
	return batches
}

func asVectors(allocations []Allocation) []vectorAllocation {
	converted := make([]vectorAllocation, 0, len(allocations))
	for _, allocation := range allocations {
		converted = append(converted, vectorAllocation(allocation))
	}
	return converted
}

func balancesOf(batches []Batch) map[string]int {
	balances := map[string]int{}
	for _, batch := range batches {
		balances[batch.ID] = batch.Balance
	}
	return balances
}

func TestBalances_SumOfMovementsPerBatch(t *testing.T) {
	for _, vector := range loadVectors(t).Balances {
		t.Run(vector.Name, func(t *testing.T) {
			movements := make([]Movement, 0, len(vector.Movements))
			for _, movement := range vector.Movements {
				movements = append(movements, Movement(movement))
			}

			if got := Balances(movements); !reflect.DeepEqual(got, vector.Balances) {
				t.Fatalf("balances = %v, want %v", got, vector.Balances)
			}
		})
	}
}

func TestAllocate_BRSTK02_FirstExpiryFirstOut(t *testing.T) {
	for _, vector := range loadVectors(t).Allocations {
		t.Run(vector.Name, func(t *testing.T) {
			batches := batchesOf(t, vector.Batches)

			got, err := Allocate(batches, vector.Qty)
			if err != nil {
				t.Fatalf("allocate: %v", err)
			}

			if !reflect.DeepEqual(asVectors(got), vector.Allocations) {
				t.Fatalf("allocations = %v, want %v", asVectors(got), vector.Allocations)
			}
			if remaining := balancesOf(Deduct(batches, got)); !reflect.DeepEqual(remaining, vector.Remaining) {
				t.Fatalf("remaining = %v, want %v", remaining, vector.Remaining)
			}
		})
	}
}

func TestAllocate_BRSTK01_RefusesMoreThanTheStock(t *testing.T) {
	for _, vector := range loadVectors(t).Shortages {
		t.Run(vector.Name, func(t *testing.T) {
			_, err := Allocate(batchesOf(t, vector.Batches), vector.Qty)

			if !errors.Is(err, ErrInsufficientStock) {
				t.Fatalf("err = %v, want ErrInsufficientStock", err)
			}
		})
	}
}

func TestAllocate_BRSTK01_RefusesNegativeQuantity(t *testing.T) {
	_, err := Allocate([]Batch{{ID: "B1", Balance: 5}}, -1)

	if !errors.Is(err, ErrNegativeQuantity) {
		t.Fatalf("err = %v, want ErrNegativeQuantity", err)
	}
}

func TestAllocate_LeavesTheGivenBatchesInTheirOrder(t *testing.T) {
	batches := batchesOf(t, loadVectors(t).Allocations[0].Batches)
	before := append([]Batch(nil), batches...)

	if _, err := Allocate(batches, 8); err != nil {
		t.Fatalf("allocate: %v", err)
	}

	if !reflect.DeepEqual(batches, before) {
		t.Fatalf("batches changed to %v", batches)
	}
}

func TestFlagOn_BRSTK04_NearExpiryThresholds(t *testing.T) {
	for _, vector := range loadVectors(t).Flags {
		t.Run(vector.Name, func(t *testing.T) {
			shelfLife := UnknownShelfLife()
			if vector.ShelfLifeDays != nil {
				shelfLife = ShelfLifeOf(*vector.ShelfLifeDays)
			}

			got := expiryOf(t, vector.ExpiryDate).FlagOn(dateOf(t, vector.Today), shelfLife)

			if string(got) != vector.Flag {
				t.Fatalf("flag = %s, want %s", got, vector.Flag)
			}
		})
	}
}

func TestFlagOn_BRSTK04_TodayIsReadInItsOwnZone(t *testing.T) {
	jakarta := time.FixedZone("WIB", 7*60*60)
	lateNightInJakarta := time.Date(2026, 11, 10, 23, 30, 0, 0, jakarta)

	got := ExpiresOn(dateOf(t, "2026-11-10")).FlagOn(lateNightInJakarta, ShelfLifeOf(30))

	if got != NearExpiry {
		t.Fatalf("flag = %s, want NEAR_EXPIRY", got)
	}
}

func BenchmarkAllocate_RPERF06_SmallAllocations(b *testing.B) {
	november := func(day int) Expiry { return ExpiresOn(time.Date(2026, 11, day, 0, 0, 0, 0, time.UTC)) }
	batches := []Batch{{ID: "B2", Expiry: november(20), Balance: 12}, {ID: "B1", Expiry: november(10), Balance: 5}}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Allocate(batches, 8)
	}
}
