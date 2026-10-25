package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
)

func boardSortChoices() []string {
	return []string{string(Nearest), string(Newest)}
}

type NeedSearch struct {
	Search
}

func NewNeedSearch(form SearchForm) (NeedSearch, error) {
	search, problems := readSearch(form)
	problems = append(problems, field.Collect(field.CheckChoice("sort", form.Sort, boardSortChoices()))...)
	if len(problems) > 0 {
		return NeedSearch{}, apperror.Validation(problems...)
	}
	return NeedSearch{Search: search}, nil
}

type NeedQuery struct {
	Listing
	CategoryIDs []string
	Text        string
	Now         time.Time
}

type NeedStore struct {
	ID             string
	Slug           string
	Name           string
	CityName       string
	StoreType      string
	LogoStorageKey string
}

type NeedCard struct {
	ID                   string
	Title                string
	Description          string
	Category             Category
	WeeklyQuantity       int
	MaxRetailPrice       int64
	PreferredModel       string
	OfferedCommissionBPS int
	HasOfferedCommission bool
	ExpiresAt            time.Time
	CreatedAt            time.Time
	Store                NeedStore
	Distance             Distance
	IsOwn                bool
	Position             Position
}

func (c NeedCard) rankKey() rankKey {
	return rankKey{meters: c.Distance.Meters, id: c.ID}
}

func (c NeedCard) SortPosition() Position {
	return c.Position
}
