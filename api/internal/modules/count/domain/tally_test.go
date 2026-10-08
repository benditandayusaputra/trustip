package domain

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
)

type vectorSplit struct {
	Gross          int64 `json:"gross"`
	StoreShare     int64 `json:"store_share"`
	ConsignorShare int64 `json:"consignor_share"`
}

type vectorBatch struct {
	ID         string      `json:"id"`
	ExpiryDate *string     `json:"expiry_date"`
	ReceivedAt time.Time   `json:"received_at"`
	Balance    int         `json:"balance"`
	RawTerms   vectorTerms `json:"terms"`
}

type vectorTerms struct {
	Model         string `json:"model"`
	RetailPrice   int64  `json:"retail_price"`
	CommissionBPS int    `json:"commission_bps"`
	DepositPrice  int64  `json:"deposit_price"`
}

type vectorLineInput struct {
	ProductID    string        `json:"product_id"`
	Batches      []vectorBatch `json:"batches"`
	CountedQty   int           `json:"counted_qty"`
	ReturnedQty  int           `json:"returned_qty"`
	ReturnReason string        `json:"return_reason"`
}

type vectorSale struct {
	BatchID string      `json:"batch_id"`
	Qty     int         `json:"qty"`
	Split   vectorSplit `json:"split"`
}

type vectorReturn struct {
	BatchID string `json:"batch_id"`
	Qty     int    `json:"qty"`
}

type vectorLine struct {
	ProductID    string         `json:"product_id"`
	RecordedQty  int            `json:"recorded_qty"`
	SoldQty      int            `json:"sold_qty"`
	ReturnedQty  int            `json:"returned_qty"`
	KeptQty      int            `json:"kept_qty"`
	ReturnReason string         `json:"return_reason"`
	Sales        []vectorSale   `json:"sales"`
	Returns      []vectorReturn `json:"returns"`
	Split        vectorSplit    `json:"split"`
}

type vectorTally struct {
	Lines       []vectorLine `json:"lines"`
	SoldQty     int          `json:"sold_qty"`
	ReturnedQty int          `json:"returned_qty"`
	Split       vectorSplit  `json:"split"`
}

type vectors struct {
	Counts []struct {
		Name   string            `json:"name"`
		Lines  []vectorLineInput `json:"lines"`
		Result vectorTally       `json:"result"`
	} `json:"counts"`
	Rejections []struct {
		Name string `json:"name"`
		Line struct {
			RecordedQty  int    `json:"recorded_qty"`
			CountedQty   int    `json:"counted_qty"`
			ReturnedQty  int    `json:"returned_qty"`
			ReturnReason string `json:"return_reason"`
		} `json:"line"`
		Field string `json:"field"`
		Code  string `json:"code"`
	} `json:"rejections"`
}

func loadVectors(t *testing.T) vectors {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "testdata", "vectors", "count.json"))
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}
	var loaded vectors
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatalf("decode vectors: %v", err)
	}
	return loaded
}

func pricedBatchOf(t *testing.T, raw vectorBatch) PricedBatch {
	t.Helper()
	terms, err := pricing.NewTerms(pricing.TermsInput(raw.RawTerms))
	if err != nil {
		t.Fatalf("terms %+v: %v", raw.RawTerms, err)
	}
	expiry := stock.NoExpiry()
	if raw.ExpiryDate != nil {
		date, err := time.Parse(time.DateOnly, *raw.ExpiryDate)
		if err != nil {
			t.Fatalf("expiry %q: %v", *raw.ExpiryDate, err)
		}
		expiry = stock.ExpiresOn(date)
	}
	return PricedBatch{
		Stock: stock.Batch{ID: raw.ID, Expiry: expiry, ReceivedAt: raw.ReceivedAt, Balance: raw.Balance},
		Terms: terms,
	}
}

func lineInputsOf(t *testing.T, raw []vectorLineInput) []LineInput {
	t.Helper()
	inputs := make([]LineInput, 0, len(raw))
	for _, line := range raw {
		batches := make([]PricedBatch, 0, len(line.Batches))
		for _, batch := range line.Batches {
			batches = append(batches, pricedBatchOf(t, batch))
		}
		inputs = append(inputs, LineInput{
			ProductID: line.ProductID, Batches: batches,
			CountedQty: line.CountedQty, ReturnedQty: line.ReturnedQty, ReturnReason: line.ReturnReason,
		})
	}
	return inputs
}

func splitVector(split pricing.Split) vectorSplit {
	return vectorSplit{
		Gross: split.Gross.Int64(), StoreShare: split.StoreShare.Int64(), ConsignorShare: split.ConsignorShare.Int64(),
	}
}

func lineVector(line Line) vectorLine {
	sales := make([]vectorSale, 0, len(line.Sales))
	for _, sale := range line.Sales {
		sales = append(sales, vectorSale{BatchID: sale.BatchID, Qty: sale.Qty, Split: splitVector(sale.Split)})
	}
	returns := make([]vectorReturn, 0, len(line.Returns))
	for _, returned := range line.Returns {
		returns = append(returns, vectorReturn(returned))
	}
	return vectorLine{
		ProductID: line.ProductID, RecordedQty: line.RecordedQty, SoldQty: line.SoldQty,
		ReturnedQty: line.ReturnedQty, KeptQty: line.KeptQty, ReturnReason: string(line.ReturnReason),
		Sales: sales, Returns: returns, Split: splitVector(line.Split),
	}
}

func tallyVector(tally Tally) vectorTally {
	lines := make([]vectorLine, 0, len(tally.Lines))
	for _, line := range tally.Lines {
		lines = append(lines, lineVector(line))
	}
	return vectorTally{
		Lines: lines, SoldQty: tally.SoldQty, ReturnedQty: tally.ReturnedQty, Split: splitVector(tally.Split),
	}
}

func TestTallyCount_BRCNT08_SharedVectors(t *testing.T) {
	for _, vector := range loadVectors(t).Counts {
		t.Run(vector.Name, func(t *testing.T) {
			got, err := TallyCount(lineInputsOf(t, vector.Lines))
			if err != nil {
				t.Fatalf("tally: %v", err)
			}

			if !reflect.DeepEqual(tallyVector(got), vector.Result) {
				t.Fatalf("tally = %+v\nwant  %+v", tallyVector(got), vector.Result)
			}
		})
	}
}

func TestTallyCount_BRCNT01_BRCNT02_Rejections(t *testing.T) {
	for _, vector := range loadVectors(t).Rejections {
		t.Run(vector.Name, func(t *testing.T) {
			input := LineInput{
				ProductID:    "roti",
				Batches:      []PricedBatch{{Stock: stock.Batch{ID: "roti-1", Balance: vector.Line.RecordedQty}}},
				CountedQty:   vector.Line.CountedQty,
				ReturnedQty:  vector.Line.ReturnedQty,
				ReturnReason: vector.Line.ReturnReason,
			}

			_, err := TallyCount([]LineInput{input})

			if got := fieldCodes(t, err)[vector.Field]; got != vector.Code {
				t.Fatalf("codes = %v, want %s %s", fieldCodes(t, err), vector.Field, vector.Code)
			}
		})
	}
}

func TestTallyCount_BRCNT01_ExceedingStockNamesTheRecordedStock(t *testing.T) {
	batches := []PricedBatch{{Stock: stock.Batch{ID: "b", Balance: 20}}}
	input := LineInput{ProductID: "roti", Batches: batches, CountedQty: 21}

	_, err := TallyCount([]LineInput{input})

	appErr, _ := errors.AsType[*apperror.Error](err)
	if len(appErr.Fields()) != 1 || appErr.Fields()[0].Message !=
		"Sisa melebihi stok tercatat (20). Catat dulu titipan masuk yang belum dicatat." {
		t.Fatalf("fields = %+v", appErr.Fields())
	}
}

func TestTallyCount_BRCNT09_ReturnedUnitsCarryNoMoney(t *testing.T) {
	terms := pricing.Terms{Model: pricing.FixedDeposit, RetailPrice: 2000, DepositPrice: 1750}
	input := LineInput{
		ProductID: "roti", Batches: []PricedBatch{{Stock: stock.Batch{ID: "b", Balance: 5}, Terms: terms}},
		CountedQty: 5, ReturnedQty: 5, ReturnReason: string(ReasonExpired),
	}

	got, err := TallyCount([]LineInput{input})

	if err != nil || got.Split != (pricing.Split{}) || got.ReturnedQty != 5 {
		t.Fatalf("tally = %+v, err = %v", got, err)
	}
}

func TestTallyCount_ReasonIsDroppedWhenNothingIsReturned(t *testing.T) {
	input := LineInput{
		ProductID: "roti", Batches: []PricedBatch{{Stock: stock.Batch{ID: "b", Balance: 5}}},
		CountedQty: 5, ReturnReason: string(ReasonOther),
	}

	got, err := TallyCount([]LineInput{input})

	if err != nil || got.Lines[0].ReturnReason != "" {
		t.Fatalf("tally = %+v, err = %v", got, err)
	}
}

func TestTallyCount_CollectsProblemsOfEveryLine(t *testing.T) {
	batches := []PricedBatch{{Stock: stock.Batch{ID: "b", Balance: 5}}}
	inputs := []LineInput{
		{ProductID: "a", Batches: batches, CountedQty: 6},
		{ProductID: "b", Batches: batches, CountedQty: 5, ReturnedQty: 1},
	}

	_, err := TallyCount(inputs)

	codes := fieldCodes(t, err)
	if codes["lines[0].counted_qty"] != "COUNT_EXCEEDS_STOCK" || codes["lines[1].return_reason"] != "REQUIRED" {
		t.Fatalf("codes = %v", codes)
	}
}

func fieldCodes(t *testing.T, err error) map[string]string {
	t.Helper()
	appErr, ok := errors.AsType[*apperror.Error](err)
	if !ok {
		t.Fatalf("err = %v, want a validation error", err)
	}
	codes := map[string]string{}
	for _, problem := range appErr.Fields() {
		codes[problem.Field] = problem.Code
	}
	return codes
}

func BenchmarkTallyCount_RPERF06_OneLine(b *testing.B) {
	terms := pricing.Terms{Model: pricing.Commission, RetailPrice: 15000, CommissionBPS: 2000}
	input := []LineInput{{
		ProductID: "keripik", CountedQty: 9, ReturnedQty: 4, ReturnReason: string(ReasonUnsold),
		Batches: []PricedBatch{
			{Stock: stock.Batch{ID: "B1", Balance: 5}, Terms: terms},
			{Stock: stock.Batch{ID: "B2", Balance: 12}, Terms: terms},
		},
	}}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = TallyCount(input)
	}
}
