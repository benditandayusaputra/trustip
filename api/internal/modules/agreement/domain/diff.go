package domain

import (
	"strconv"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
)

type TermChange struct {
	Field  string
	Before string
	After  string
}

type ItemChangeKind string

const (
	ItemAdded   ItemChangeKind = "ADDED"
	ItemUpdated ItemChangeKind = "UPDATED"
	ItemStopped ItemChangeKind = "STOPPED"
)

type ItemChange struct {
	ProductID   string
	ProductName string
	Kind        ItemChangeKind
	Before      pricing.Terms
	After       pricing.Terms
}

type Diff struct {
	Terms []TermChange
	Items []ItemChange
}

func DiffOf(previous, next Agreement) Diff {
	return Diff{Terms: termChanges(previous, next), Items: itemChanges(previous, next)}
}

func termChanges(previous, next Agreement) []TermChange {
	pairs := []TermChange{
		{"settlement_schedule", string(previous.Terms.Schedule), string(next.Terms.Schedule)},
		{
			"payment_terms_days", strconv.Itoa(previous.Terms.PaymentTermsDays),
			strconv.Itoa(next.Terms.PaymentTermsDays),
		},
		{"count_rhythm", string(previous.Terms.Rhythm), string(next.Terms.Rhythm)},
		{"notes", previous.Terms.Notes, next.Terms.Notes},
	}
	changes := []TermChange{}
	for _, pair := range pairs {
		if pair.Before != pair.After {
			changes = append(changes, pair)
		}
	}
	return changes
}

func itemChanges(previous, next Agreement) []ItemChange {
	before := map[string]Item{}
	for _, item := range previous.Items {
		before[item.ProductID] = item
	}
	changes := []ItemChange{}
	for _, item := range next.Items {
		old, existed := before[item.ProductID]
		if kind, changed := changeKind(old, existed, item); changed {
			changes = append(changes, ItemChange{
				ProductID: item.ProductID, ProductName: item.ProductName, Kind: kind,
				Before: old.Terms, After: item.Terms,
			})
		}
	}
	return changes
}

func changeKind(old Item, existed bool, item Item) (ItemChangeKind, bool) {
	switch {
	case item.IsActive && (!existed || !old.IsActive):
		return ItemAdded, true
	case !item.IsActive && existed && old.IsActive:
		return ItemStopped, true
	case item.IsActive && old.Terms != item.Terms:
		return ItemUpdated, true
	}
	return "", false
}

type Version struct {
	Agreement Agreement
	Diff      Diff
}

func History(agreements []Agreement) []Version {
	versions := make([]Version, 0, len(agreements))
	var previous Agreement
	for _, agreement := range agreements {
		versions = append(versions, Version{Agreement: agreement, Diff: DiffOf(previous, agreement)})
		previous = agreement
	}
	return versions
}
