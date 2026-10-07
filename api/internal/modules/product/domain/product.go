package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
)

const (
	MinNameLength        = 2
	MaxNameLength        = 80
	MaxSKULength         = 20
	MaxUnitLength        = 20
	MaxDescriptionLength = 1000
	MaxCompositionLength = 500
	MaxNetContentLength  = 40
	MaxShelfLife         = 730
	skuPrefixLength      = 3
	skuPadding           = "X"
)

var (
	ErrSKUTaken        = errors.New("sku is already used by another product of this consignor")
	ErrUnknownCategory = errors.New("category is not in the category list")
	ErrUnknownPhoto    = errors.New("photo is not a product photo of this consignor")
	skuFormat          = regexp.MustCompile(`^[A-Z0-9]+(-[A-Z0-9]+)*$`)
	idFormat           = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

type SuggestedPrices struct {
	Retail  int64
	Deposit int64
}

type Prices struct {
	Retail  money.Rupiah
	Deposit money.Rupiah
}

type ShelfLife struct {
	Given bool
	Days  int
}

type ProductForm struct {
	SKU         string
	Name        string
	CategoryID  string
	Unit        string
	Description string
	Composition string
	NetContent  string
	PhotoFileID string
	Prices      SuggestedPrices
	ShelfLife   ShelfLife
}

type Details struct {
	Name        string
	CategoryID  string
	Unit        string
	Description string
	Composition string
	NetContent  string
	PhotoFileID string
	Prices      Prices
	ShelfLife   ShelfLife
}

type Draft struct {
	SKU string
	Details
}

func (d Draft) HasAutoSKU() bool {
	return d.SKU == ""
}

func NewDraft(form ProductForm) (Draft, error) {
	sku := strings.ToUpper(form.SKU)
	prices, priceProblems := checkPrices(form.Prices)
	problems := field.Collect(
		findSKUProblem(sku),
		field.CheckLength("name", form.Name, field.Length{Min: MinNameLength, Max: MaxNameLength}),
		findIDProblem("category_id", form.CategoryID),
		findRequiredTextProblem("unit", form.Unit, MaxUnitLength),
		field.CheckLength("description", form.Description, field.Length{Max: MaxDescriptionLength}),
		field.CheckLength("composition", form.Composition, field.Length{Max: MaxCompositionLength}),
		field.CheckLength("net_content", form.NetContent, field.Length{Max: MaxNetContentLength}),
		findOptionalIDProblem("photo_file_id", form.PhotoFileID),
		findShelfLifeProblem(form.ShelfLife),
	)
	problems = append(problems, priceProblems...)
	if len(problems) > 0 {
		return Draft{}, apperror.Validation(problems...)
	}
	return Draft{SKU: sku, Details: Details{
		Name: form.Name, CategoryID: form.CategoryID, Unit: form.Unit, Description: form.Description,
		Composition: form.Composition, NetContent: form.NetContent, PhotoFileID: form.PhotoFileID,
		Prices: prices, ShelfLife: form.ShelfLife,
	}}, nil
}

func findSKUProblem(sku string) field.Problem {
	if sku == "" {
		return field.Problem{}
	}
	if !skuFormat.MatchString(sku) {
		return field.Invalid("sku", apperror.FieldInvalidCharacters)
	}
	return field.CheckLength("sku", sku, field.Length{Max: MaxSKULength})
}

func findRequiredTextProblem(name, value string, maxLength int) field.Problem {
	if value == "" {
		return field.Invalid(name, apperror.FieldRequired)
	}
	return field.CheckLength(name, value, field.Length{Max: maxLength})
}

func findIDProblem(name, id string) field.Problem {
	if id == "" {
		return field.Invalid(name, apperror.FieldRequired)
	}
	return findOptionalIDProblem(name, id)
}

func findOptionalIDProblem(name, id string) field.Problem {
	if id == "" || idFormat.MatchString(id) {
		return field.Problem{}
	}
	return field.Invalid(name, apperror.FieldUnknownChoice)
}

func findShelfLifeProblem(shelfLife ShelfLife) field.Problem {
	if !shelfLife.Given || (shelfLife.Days >= 1 && shelfLife.Days <= MaxShelfLife) {
		return field.Problem{}
	}
	return field.Invalid("shelf_life_days", apperror.FieldOutOfRange)
}

func checkPrices(suggested SuggestedPrices) (Prices, []apperror.FieldError) {
	retail, retailErr := money.NewRupiah(suggested.Retail)
	deposit, depositErr := money.NewRupiah(suggested.Deposit)
	var problems []field.Problem
	if retailErr != nil {
		problems = append(problems, field.Invalid("suggested_retail_price", apperror.FieldOutOfRange))
	}
	switch {
	case depositErr != nil:
		problems = append(problems, field.Invalid("suggested_deposit_price", apperror.FieldOutOfRange))
	case retailErr == nil && deposit > retail:
		problems = append(problems, field.Invalid("suggested_deposit_price", apperror.FieldExceedsRetail))
	}
	return Prices{Retail: retail, Deposit: deposit}, field.Collect(problems...)
}

func SKUPrefix(name string) string {
	var letters strings.Builder
	for _, char := range strings.ToUpper(name) {
		if char >= 'A' && char <= 'Z' && letters.Len() < skuPrefixLength {
			letters.WriteRune(char)
		}
	}
	return letters.String() + strings.Repeat(skuPadding, skuPrefixLength-letters.Len())
}

func AutoSKU(prefix string, sequence int) string {
	return fmt.Sprintf("%s-%04d", prefix, sequence)
}

type FileRef struct {
	ID         string
	StorageKey string
}

type Product struct {
	ID           string
	ConsignorID  string
	SKU          string
	Details      Details
	CategoryName string
	Photo        FileRef
	IsArchived   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (p Product) Status() Status {
	if p.IsArchived {
		return Archived
	}
	return Active
}

type Change struct {
	ProductID   string
	ConsignorID string
	SKU         string
	Details     Details
	At          time.Time
}

type StatusChange struct {
	ProductID   string
	ConsignorID string
	IsArchived  bool
	At          time.Time
}

type ProductKey struct {
	ConsignorID string
	ProductID   string
}

type SKUSeries struct {
	ConsignorID string
	Prefix      string
}
