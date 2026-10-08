package domain

import (
	"regexp"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
)

const (
	MinTitleLength       = 5
	MaxTitleLength       = 80
	MaxDescriptionLength = 500
	maxWeeklyQuantity    = 100_000
	maxRetailPrice       = 100_000_000
	maxCommissionBPS     = 5000
)

var (
	idFormat    = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	phoneInText = regexp.MustCompile(`(?:\+?\d[\s.\-()]*){9,}`)
	linkInText  = regexp.MustCompile(
		`(?i)(?:https?://|www\.|wa\.me|\b[a-z0-9-]+\.(?:com|id|net|org|co|me|ly|link|xyz)\b)`,
	)
)

func pricingChoices() []string {
	return []string{"FIXED_DEPOSIT", "COMMISSION"}
}

type Amount struct {
	Value   int64
	IsGiven bool
}

type Form struct {
	Title                string
	CategoryID           string
	Description          string
	WeeklyQuantity       Amount
	MaxRetailPrice       Amount
	PreferredModel       string
	OfferedCommissionBPS Amount
}

type Draft struct {
	Title                string
	CategoryID           string
	Description          string
	WeeklyQuantity       Amount
	MaxRetailPrice       Amount
	PreferredModel       string
	OfferedCommissionBPS Amount
}

func NewDraft(form Form) (Draft, error) {
	problems := field.Collect(
		field.CheckLength("title", form.Title, field.Length{Min: MinTitleLength, Max: MaxTitleLength}),
		findContactProblem("title", form.Title),
		field.CheckLength("description", form.Description, field.Length{Max: MaxDescriptionLength}),
		findContactProblem("description", form.Description),
		findCategoryProblem(form.CategoryID),
		findRangeProblem("est_qty_per_week", form.WeeklyQuantity, bounds{1, maxWeeklyQuantity}),
		findRangeProblem("max_retail_price", form.MaxRetailPrice, bounds{1, maxRetailPrice}),
		field.CheckChoice("preferred_model", form.PreferredModel, pricingChoices()),
		findRangeProblem("offered_commission_bps", form.OfferedCommissionBPS, bounds{0, maxCommissionBPS}),
	)
	if len(problems) > 0 {
		return Draft{}, apperror.Validation(problems...)
	}
	return Draft(form), nil
}

func findContactProblem(name, text string) field.Problem {
	if phoneInText.MatchString(text) || linkInText.MatchString(text) {
		return field.Invalid(name, apperror.FieldContactInText)
	}
	return field.Problem{}
}

func findCategoryProblem(categoryID string) field.Problem {
	if categoryID == "" {
		return field.Invalid("category_id", apperror.FieldRequired)
	}
	if !idFormat.MatchString(categoryID) {
		return field.Invalid("category_id", apperror.FieldUnknownChoice)
	}
	return field.Problem{}
}

type bounds struct {
	lowest  int64
	highest int64
}

func findRangeProblem(name string, amount Amount, allowed bounds) field.Problem {
	if amount.IsGiven && (amount.Value < allowed.lowest || amount.Value > allowed.highest) {
		return field.Invalid(name, apperror.FieldOutOfRange)
	}
	return field.Problem{}
}
