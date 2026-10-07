package pricing

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

type vectorSplit struct {
	Gross          int64 `json:"gross"`
	StoreShare     int64 `json:"store_share"`
	ConsignorShare int64 `json:"consignor_share"`
}

type vectorTerms struct {
	Model         string `json:"model"`
	RetailPrice   int64  `json:"retail_price"`
	CommissionBPS int    `json:"commission_bps"`
	DepositPrice  int64  `json:"deposit_price"`
}

type vectors struct {
	Allocations []struct {
		Name  string      `json:"name"`
		Terms vectorTerms `json:"terms"`
		Qty   int         `json:"qty"`
		Split vectorSplit `json:"split"`
	} `json:"allocations"`
	Summary struct {
		Lines []int       `json:"lines"`
		Total vectorSplit `json:"total"`
	} `json:"summary"`
	ExampleQuantity int `json:"example_quantity"`
}

func loadVectors(t *testing.T) vectors {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "vectors", "pricing.json"))
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}
	var loaded vectors
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatalf("decode vectors: %v", err)
	}
	return loaded
}

func termsOf(t *testing.T, raw vectorTerms) Terms {
	t.Helper()
	terms, err := NewTerms(TermsInput(raw))
	if err != nil {
		t.Fatalf("terms %+v: %v", raw, err)
	}
	return terms
}

func asVector(split Split) vectorSplit {
	return vectorSplit{
		Gross: split.Gross.Int64(), StoreShare: split.StoreShare.Int64(), ConsignorShare: split.ConsignorShare.Int64(),
	}
}

func TestSplitOf_BRSTL04_SharedVectors(t *testing.T) {
	for _, vector := range loadVectors(t).Allocations {
		t.Run(vector.Name, func(t *testing.T) {
			got := termsOf(t, vector.Terms).SplitOf(vector.Qty)

			if asVector(got) != vector.Split {
				t.Fatalf("split = %+v, want %+v", asVector(got), vector.Split)
			}
		})
	}
}

func TestSummarize_BRSTL04_TotalsAddLinesWithoutRecomputingCommission(t *testing.T) {
	loaded := loadVectors(t)
	splits := make([]Split, 0, len(loaded.Summary.Lines))
	for _, index := range loaded.Summary.Lines {
		line := loaded.Allocations[index]
		splits = append(splits, termsOf(t, line.Terms).SplitOf(line.Qty))
	}

	if got := asVector(Summarize(splits)); got != loaded.Summary.Total {
		t.Fatalf("total = %+v, want %+v", got, loaded.Summary.Total)
	}
}

func TestExample_FRPRP09_PreviewSellsTenUnits(t *testing.T) {
	terms := termsOf(t, vectorTerms{Model: "FIXED_DEPOSIT", RetailPrice: 2000, DepositPrice: 1750})

	got := terms.Example()

	if loadVectors(t).ExampleQuantity != ExampleQuantity || got.Gross != 20000 || got.ConsignorShare != 17500 {
		t.Fatalf("example = %+v", got)
	}
}

func TestPerUnit_FRAGR01_ThreeNumbersPerProduct(t *testing.T) {
	terms := termsOf(t, vectorTerms{Model: "COMMISSION", RetailPrice: 15000, CommissionBPS: 2000})

	got := terms.PerUnit()

	if got.Gross != 15000 || got.StoreShare != 3000 || got.ConsignorShare != 12000 {
		t.Fatalf("per unit = %+v", got)
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

func TestNewTerms_BRAGR01To03_Boundaries(t *testing.T) {
	cases := []struct {
		name  string
		input TermsInput
		field string
		code  string
	}{
		{"retail at the top", TermsInput{Model: "COMMISSION", RetailPrice: 100_000_000, CommissionBPS: 5000}, "", ""},
		{
			"retail above the top",
			TermsInput{Model: "COMMISSION", RetailPrice: 100_000_001},
			"retail_price", "OUT_OF_RANGE",
		},
		{"negative retail", TermsInput{Model: "COMMISSION", RetailPrice: -1}, "retail_price", "OUT_OF_RANGE"},
		{
			"commission above half",
			TermsInput{Model: "COMMISSION", RetailPrice: 1, CommissionBPS: 5001},
			"commission_bps", "OUT_OF_RANGE",
		},
		{
			"negative commission",
			TermsInput{Model: "COMMISSION", RetailPrice: 1, CommissionBPS: -1},
			"commission_bps", "OUT_OF_RANGE",
		},
		{"deposit equal to retail", TermsInput{Model: "FIXED_DEPOSIT", RetailPrice: 2000, DepositPrice: 2000}, "", ""},
		{
			"deposit above retail",
			TermsInput{Model: "FIXED_DEPOSIT", RetailPrice: 2000, DepositPrice: 2001},
			"deposit_price", "EXCEEDS_RETAIL_PRICE",
		},
		{"unknown model", TermsInput{Model: "SHELF_RENT", RetailPrice: 2000}, "pricing_model", "UNKNOWN_CHOICE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewTerms(tc.input)

			if tc.field == "" && err != nil {
				t.Fatalf("err = %v", err)
			}
			if tc.field != "" && fieldCodes(t, err)[tc.field] != tc.code {
				t.Fatalf("codes = %v, want %s %s", fieldCodes(t, err), tc.field, tc.code)
			}
		})
	}
}

func TestNewTerms_FixedDepositIgnoresCommissionAndTheOtherWayRound(t *testing.T) {
	deposit, depositErr := NewTerms(TermsInput{
		Model: "FIXED_DEPOSIT", RetailPrice: 2000, DepositPrice: 1750,
		CommissionBPS: 9999,
	})
	commission, commissionErr := NewTerms(TermsInput{
		Model: "COMMISSION", RetailPrice: 2000, CommissionBPS: 1000,
		DepositPrice: 99999,
	})

	if depositErr != nil || commissionErr != nil || deposit.CommissionBPS != 0 || commission.DepositPrice != 0 {
		t.Fatalf("deposit = %+v %v, commission = %+v %v", deposit, depositErr, commission, commissionErr)
	}
}

func BenchmarkSplitOf_RPERF06_DoesNotAllocate(b *testing.B) {
	terms := Terms{Model: Commission, RetailPrice: 22500, CommissionBPS: 1750}
	b.ReportAllocs()
	for b.Loop() {
		_ = terms.SplitOf(3)
	}
}
