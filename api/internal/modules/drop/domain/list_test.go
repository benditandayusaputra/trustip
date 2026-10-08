package domain

import (
	"testing"
	"time"
)

func TestNewListQuery_FRDRP08_DatesCoverWholeJakartaDays(t *testing.T) {
	query, err := NewListQuery(ListFilter{From: "2026-11-01", To: "2026-11-10"}, nineInTheMorning())
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	jakarta := nineInTheMorning().Location()
	if !query.From.Equal(time.Date(2026, 11, 1, 0, 0, 0, 0, jakarta)) ||
		!query.To.Equal(time.Date(2026, 11, 11, 0, 0, 0, 0, jakarta)) {
		t.Fatalf("query = %+v", query)
	}
}

func TestNewListQuery_BRDRP08_CutoffIsFortyEightHoursBeforeNow(t *testing.T) {
	query, err := NewListQuery(ListFilter{Status: "CONFIRMED"}, nineInTheMorning())

	if err != nil || !query.ConfirmCutoff.Equal(nineInTheMorning().Add(-48*time.Hour)) || query.Status != Confirmed {
		t.Fatalf("query = %+v, err = %v", query, err)
	}
}

func TestNewListQuery_OpenRangeWithoutDates(t *testing.T) {
	query, err := NewListQuery(ListFilter{}, nineInTheMorning())

	if err != nil || !query.From.Before(query.To) || query.PartnershipID != "" {
		t.Fatalf("query = %+v, err = %v", query, err)
	}
}

func TestNewListQuery_RefusesUnknownFilters(t *testing.T) {
	cases := map[string]ListFilter{
		"status":         {Status: "LOST"},
		"partnership_id": {PartnershipID: "not-an-id"},
		"from":           {From: "01-11-2026"},
		"to":             {From: "2026-11-10", To: "2026-11-01"},
	}
	for field, filter := range cases {
		t.Run(field, func(t *testing.T) {
			_, err := NewListQuery(filter, nineInTheMorning())

			if _, found := fieldCodes(t, err)[field]; !found {
				t.Fatalf("codes = %v", fieldCodes(t, err))
			}
		})
	}
}
