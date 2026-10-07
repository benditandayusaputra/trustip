package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const breadCategoryID = "01a1161e-0a30-752d-ab9b-7397976f0dba"

func validForm() ProductForm {
	return ProductForm{
		Name: "Roti Sobek Cokelat", CategoryID: breadCategoryID, Unit: "pcs",
		Prices:    SuggestedPrices{Retail: 2000, Deposit: 1750},
		ShelfLife: ShelfLife{Given: true, Days: 3},
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

func TestNewDraft_FRPRD01_ValidFormIsKept(t *testing.T) {
	draft, err := NewDraft(validForm())

	if err != nil || draft.Name != "Roti Sobek Cokelat" || draft.Prices.Deposit.Int64() != 1750 || !draft.HasAutoSKU() {
		t.Fatalf("draft = %+v, err = %v", draft, err)
	}
}

func TestNewDraft_BRAGR02_DepositPriceCannotExceedRetailPrice(t *testing.T) {
	form := validForm()
	form.Prices = SuggestedPrices{Retail: 2000, Deposit: 2001}

	_, err := NewDraft(form)

	if got := fieldCodes(t, err)["suggested_deposit_price"]; got != "EXCEEDS_RETAIL_PRICE" {
		t.Fatalf("code = %q", got)
	}
}

func TestNewDraft_BRAGR01_PricesStayWithinOneHundredMillion(t *testing.T) {
	form := validForm()
	form.Prices = SuggestedPrices{Retail: 100_000_001, Deposit: -1}

	codes := fieldCodes(t, mustFail(NewDraft(form)))

	if codes["suggested_retail_price"] != "OUT_OF_RANGE" || codes["suggested_deposit_price"] != "OUT_OF_RANGE" {
		t.Fatalf("codes = %v", codes)
	}
}

func TestNewDraft_RSEC47_ReportsEveryProblemAtOnce(t *testing.T) {
	form := ProductForm{Name: "R", SKU: "roti 1", CategoryID: "bread", ShelfLife: ShelfLife{Given: true}}

	codes := fieldCodes(t, mustFail(NewDraft(form)))

	for field, want := range map[string]string{
		"name": "TOO_SHORT", "sku": "INVALID_CHARACTERS", "category_id": "UNKNOWN_CHOICE",
		"unit": "REQUIRED", "shelf_life_days": "OUT_OF_RANGE",
	} {
		if codes[field] != want {
			t.Errorf("%s = %q, want %q", field, codes[field], want)
		}
	}
}

func TestNewDraft_SKUIsStoredInCapitals(t *testing.T) {
	form := validForm()
	form.SKU = "rot-0042"

	draft, err := NewDraft(form)

	if err != nil || draft.SKU != "ROT-0042" || draft.HasAutoSKU() {
		t.Fatalf("draft = %+v, err = %v", draft, err)
	}
}

func TestNewDraft_ShelfLifeIsOptionalButPositive(t *testing.T) {
	cases := map[ShelfLife]bool{
		{}:                                    true,
		{Given: true, Days: 1}:                true,
		{Given: true, Days: MaxShelfLife}:     true,
		{Given: true, Days: 0}:                false,
		{Given: true, Days: MaxShelfLife + 1}: false,
	}
	for shelfLife, valid := range cases {
		form := validForm()
		form.ShelfLife = shelfLife

		_, err := NewDraft(form)

		if (err == nil) != valid {
			t.Errorf("%+v: err = %v", shelfLife, err)
		}
	}
}

func TestSKUPrefix_BRPRD01_FirstThreeLettersOfTheName(t *testing.T) {
	cases := map[string]string{
		"Roti Sobek":       "ROT",
		"es jeruk":         "ESJ",
		"Kue 7 Lapis":      "KUE",
		"Éclair":           "CLA",
		"Es":               "ESX",
		"123":              "XXX",
		"Keripik Singkong": "KER",
	}
	for name, want := range cases {
		if got := SKUPrefix(name); got != want {
			t.Errorf("SKUPrefix(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestAutoSKU_BRPRD01_PrefixAndFourDigitSequence(t *testing.T) {
	if got := AutoSKU("ROT", 1); got != "ROT-0001" {
		t.Fatalf("AutoSKU = %q", got)
	}
	if got := AutoSKU("ROT", 12345); got != "ROT-12345" {
		t.Fatalf("AutoSKU = %q", got)
	}
}

func TestNewSearch_FRPRD05_DefaultsToActiveProducts(t *testing.T) {
	search, err := NewSearch(SearchForm{Text: "roti"})

	if err != nil || search.Status != Active || search.Text != "roti" {
		t.Fatalf("search = %+v, err = %v", search, err)
	}
}

func TestNewSearch_UnknownStatusOrCategoryIsRejected(t *testing.T) {
	codes := fieldCodes(t, mustFail(NewSearch(SearchForm{
		Status: "DELETED", CategoryID: "bread", Text: strings.Repeat("a", MaxSearchLength+1),
	})))

	if codes["status"] != "UNKNOWN_CHOICE" || codes["category_id"] != "UNKNOWN_CHOICE" || codes["q"] != "TOO_LONG" {
		t.Fatalf("codes = %v", codes)
	}
}

func mustFail[T any](_ T, err error) error {
	return err
}
