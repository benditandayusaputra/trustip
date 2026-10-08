package domain

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

func anIssuedSettlement(t *testing.T) Settlement {
	t.Helper()
	counted := jakarta("2026-11-30 20:00")
	draft, hasSales := aLedger(terms.ScheduleOnCount, sectionThirteen(counted)).DraftUpTo(counted)
	if !hasSales {
		t.Fatal("the ledger has sales")
	}
	return draft.Issue(Issuing{ID: "0192f0c4-0000-7000-8000-000000000701", Sequence: 1, At: counted})
}

func TestSealRequest_RSEC27_IssueIsSealedAsASystemEvent(t *testing.T) {
	settlement := anIssuedSettlement(t)

	request := settlement.SealRequest()

	if request.Event != "SETTLEMENT_ISSUED" || request.DocumentType != "SETTLEMENT" ||
		request.DocumentNumber != "ST-202611-0001" || request.ActorUserID != "" || request.ActorOrgID != "" ||
		request.Content["consignor_share_amount"] != int64(691187) {
		t.Fatalf("request = %+v", request)
	}
}

func TestContentDigest_RSEC28_ChangesWithAnAllocationAmount(t *testing.T) {
	settlement := anIssuedSettlement(t)
	tampered := settlement
	tampered.Allocations = append([]Allocation(nil), settlement.Allocations...)
	tampered.Allocations[1].Split.StoreShare = 11812

	original, originalErr := settlement.ContentDigest()
	changed, changedErr := tampered.ContentDigest()

	if errors.Join(originalErr, changedErr) != nil || bytes.Equal(original, changed) {
		t.Fatalf("digests = %x %x", original, changed)
	}
}

func TestContent_RSEC28_AllocationOrderDoesNotChangeTheDigest(t *testing.T) {
	settlement := anIssuedSettlement(t)
	reversed := settlement
	reversed.Allocations = []Allocation{}
	for index := len(settlement.Allocations) - 1; index >= 0; index-- {
		reversed.Allocations = append(reversed.Allocations, settlement.Allocations[index])
	}

	original, originalErr := settlement.ContentDigest()
	shuffled, shuffledErr := reversed.ContentDigest()

	if errors.Join(originalErr, shuffledErr) != nil || !bytes.Equal(original, shuffled) {
		t.Fatalf("digests = %x %x", original, shuffled)
	}
}

func TestNotices_FRSTL02_BothOwnersHearOfTheIssue(t *testing.T) {
	notices := anIssuedSettlement(t).Notices("Warung Bu Sri")

	if len(notices) != 2 || notices[0].OrganizationID != consignorID || notices[1].OrganizationID != storeID ||
		!notices[0].OwnersOnly || !notices[1].OwnersOnly || notices[0].Link != "/app/setoran/"+
		"0192f0c4-0000-7000-8000-000000000701" ||
		notices[0].Body != "Setoran ST-202611-0001 dari Warung Bu Sri terbit. Jatuh tempo 7 Desember 2026." {
		t.Fatalf("notices = %+v", notices)
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

func TestNewListQuery_FRSTL09_AgeBecomesADueDateRange(t *testing.T) {
	now := jakarta("2026-11-20 09:00")
	cases := []struct {
		age, from, to string
	}{
		{"", "0001-01-01", "9999-12-31"},
		{"NOT_DUE", "2026-11-20", "9999-12-31"},
		{"LATE_1_7", "2026-11-13", "2026-11-19"},
		{"LATE_8_30", "2026-10-21", "2026-11-12"},
		{"LATE_OVER_30", "0001-01-01", "2026-10-20"},
	}
	for _, tc := range cases {
		t.Run(tc.age, func(t *testing.T) {
			query, err := NewListQuery(ListFilter{Age: tc.age}, now)
			if err != nil || query.DueFrom != tc.from || query.DueTo != tc.to {
				t.Fatalf("query = %+v, err = %v", query, err)
			}
			if tc.age != "" && query.Status != string(Issued) {
				t.Fatalf("status = %q, want ISSUED", query.Status)
			}
		})
	}
}

func TestNewListQuery_RefusesUnknownFilters(t *testing.T) {
	for name, filter := range map[string]ListFilter{
		"partnership_id": {PartnershipID: "x"}, "status": {Status: "LATE"}, "age": {Age: "OLD"},
	} {
		_, err := NewListQuery(filter, jakarta("2026-11-20 09:00"))
		if _, found := fieldCodes(t, err)[name]; !found {
			t.Fatalf("%s: codes = %v", name, fieldCodes(t, err))
		}
	}
}

func TestNewListQuery_PaidCannotBeLate(t *testing.T) {
	_, err := NewListQuery(ListFilter{Status: "PAID", Age: "LATE_1_7"}, jakarta("2026-11-20 09:00"))
	if fieldCodes(t, err)["age"] != "OUT_OF_RANGE" {
		t.Fatalf("codes = %v", fieldCodes(t, err))
	}
}

func TestView_FRSTL04_StandingUsesTheReadTime(t *testing.T) {
	view := View{Settlement: anIssuedSettlement(t)}

	standing := view.StandingAt(view.DueDate.Add(24 * time.Hour))

	if !standing.IsOverdue || standing.DaysLate != 1 {
		t.Fatalf("standing = %+v", standing)
	}
}
