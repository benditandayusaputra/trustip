package domain

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
)

const (
	MaxMessageLength = 1000
	MaxItems         = 50
)

var idFormat = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type Action string

const (
	Add    Action = "ADD"
	Update Action = "UPDATE"
	Stop   Action = "STOP"
)

type ItemInput struct {
	ProductID string
	Terms     pricing.TermsInput
}

type Item struct {
	ProductID   string
	ProductName string
	SKU         string
	Unit        string
	Action      Action
	Terms       pricing.Terms
}

type RevisionInput struct {
	Terms   TermsInput
	Items   []ItemInput
	Message string
}

type RevisionDraft struct {
	Terms   GeneralTerms
	Items   []Item
	Message string
}

func NewRevisionDraft(input RevisionInput) (RevisionDraft, error) {
	terms, termProblems := checkTerms(input.Terms)
	problems := field.Collect(append(termProblems,
		field.CheckLength("message", input.Message, field.Length{Max: MaxMessageLength}),
		findItemCountProblem(len(input.Items)),
	)...)
	items, itemProblems := checkItems(input.Items)
	problems = append(problems, itemProblems...)
	if len(problems) > 0 {
		return RevisionDraft{}, apperror.Validation(problems...)
	}
	return RevisionDraft{Terms: terms, Items: items, Message: input.Message}, nil
}

func findItemCountProblem(count int) field.Problem {
	if count > MaxItems {
		return field.Invalid("items", apperror.FieldTooMany, field.Limit("max", MaxItems))
	}
	return field.Problem{}
}

func checkItems(inputs []ItemInput) ([]Item, []apperror.FieldError) {
	items := make([]Item, 0, len(inputs))
	var problems []apperror.FieldError
	seen := map[string]bool{}
	for index, input := range inputs {
		prefix := fmt.Sprintf("items[%d]", index)
		if seen[input.ProductID] || !idFormat.MatchString(input.ProductID) {
			problems = append(problems, apperror.InvalidField(prefix+".product_id", apperror.FieldUnknownChoice))
		}
		seen[input.ProductID] = true
		terms, err := pricing.NewTerms(input.Terms)
		problems = append(problems, prefixed(prefix, err)...)
		items = append(items, Item{ProductID: input.ProductID, Action: Add, Terms: terms})
	}
	return items, problems
}

func prefixed(prefix string, err error) []apperror.FieldError {
	appErr, ok := errors.AsType[*apperror.Error](err)
	if !ok {
		return nil
	}
	problems := make([]apperror.FieldError, 0, len(appErr.Fields()))
	for _, problem := range appErr.Fields() {
		problem.Field = prefix + "." + problem.Field
		problems = append(problems, problem)
	}
	return problems
}
