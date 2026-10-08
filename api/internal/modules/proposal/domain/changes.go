package domain

import "slices"

type ItemChange string

const (
	ItemAdded   ItemChange = "ADDED"
	ItemChanged ItemChange = "CHANGED"
	ItemKept    ItemChange = "KEPT"
	ItemRemoved ItemChange = "REMOVED"
)

type ItemDifference struct {
	Item   Item
	Change ItemChange
}

type RevisionChanges struct {
	TermFields []string
	Items      []ItemDifference
}

type RevisionRecord struct {
	Revision   Revision
	AuthorName string
	Changes    RevisionChanges
}

func (p Proposal) History(revisions []Revision) []RevisionRecord {
	records := make([]RevisionRecord, 0, len(revisions))
	previous := Revision{}
	for index, revision := range revisions {
		changes := Compare(previous, revision)
		if index == 0 {
			changes = firstRevisionChanges(revision)
		}
		records = append(records, RevisionRecord{
			Revision: revision, AuthorName: p.nameOf(revision.AuthorOrgID), Changes: changes,
		})
		previous = revision
	}
	return records
}

func firstRevisionChanges(revision Revision) RevisionChanges {
	items := make([]ItemDifference, 0, len(revision.Items))
	for _, item := range revision.Items {
		items = append(items, ItemDifference{Item: item, Change: ItemAdded})
	}
	return RevisionChanges{TermFields: []string{}, Items: items}
}

func Compare(previous, current Revision) RevisionChanges {
	return RevisionChanges{TermFields: changedTermFields(previous, current), Items: itemDifferences(previous, current)}
}

func changedTermFields(previous, current Revision) []string {
	before, after := previous.Terms, current.Terms
	checks := []struct {
		name      string
		isChanged bool
	}{
		{"settlement_schedule", before.Schedule != after.Schedule},
		{"payment_terms_days", before.PaymentTermsDays != after.PaymentTermsDays},
		{"count_rhythm", before.Rhythm != after.Rhythm},
		{"notes", before.Notes != after.Notes},
	}
	fields := []string{}
	for _, check := range checks {
		if check.isChanged {
			fields = append(fields, check.name)
		}
	}
	return fields
}

func itemDifferences(previous, current Revision) []ItemDifference {
	differences := make([]ItemDifference, 0, len(current.Items))
	for _, item := range current.Items {
		differences = append(differences, ItemDifference{Item: item, Change: changeOf(previous.Items, item)})
	}
	for _, item := range previous.Items {
		if !containsProduct(current.Items, item.ProductID) {
			differences = append(differences, ItemDifference{Item: item, Change: ItemRemoved})
		}
	}
	return differences
}

func changeOf(previousItems []Item, item Item) ItemChange {
	index := slices.IndexFunc(previousItems, func(candidate Item) bool { return candidate.ProductID == item.ProductID })
	if index < 0 {
		return ItemAdded
	}
	before := previousItems[index]
	if before.Action != item.Action || before.Terms != item.Terms {
		return ItemChanged
	}
	return ItemKept
}

func containsProduct(items []Item, productID string) bool {
	return slices.ContainsFunc(items, func(item Item) bool { return item.ProductID == productID })
}
