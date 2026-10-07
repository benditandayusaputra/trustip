package domain

import (
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pagination"
)

const MaxSearchLength = 80

type Status string

const (
	Active   Status = "ACTIVE"
	Archived Status = "ARCHIVED"
)

type SearchForm struct {
	Text       string
	CategoryID string
	Status     string
}

type Search struct {
	Text       string
	CategoryID string
	Status     Status
}

func NewSearch(form SearchForm) (Search, error) {
	status := Status(form.Status)
	if form.Status == "" {
		status = Active
	}
	problems := field.Collect(
		field.CheckLength("q", form.Text, field.Length{Max: MaxSearchLength}),
		findOptionalIDProblem("category_id", form.CategoryID),
		field.CheckChoice("status", string(status), []string{string(Active), string(Archived)}),
	)
	if len(problems) > 0 {
		return Search{}, apperror.Validation(problems...)
	}
	return Search{Text: form.Text, CategoryID: form.CategoryID, Status: status}, nil
}

type PageQuery struct {
	ConsignorID string
	Search      Search
	After       pagination.Position
	Limit       int
}
