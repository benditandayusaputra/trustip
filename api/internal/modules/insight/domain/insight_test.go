package domain

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
)

func TestGenerate_INS01_ExpiredBatchWithStockIsHigh(t *testing.T) {
	snapshot := storeSnapshot()
	snapshot.Products = []ProductStock{aProduct().withBatchExpiringIn(-1, 4).build()}

	got := onlyInsight(t, Generate(snapshot, today()))

	if got.Kind != KindExpired || got.Level != LevelHigh || got.Action != ActionCountWithReturns {
		t.Fatalf("insight = %+v, want EXPIRED HIGH COUNT_WITH_RETURNS", got)
	}
	if got.Units != 4 || got.Amount != money.Rupiah(8000) || got.Subject.BatchCode != "B1" {
		t.Fatalf("figures = %+v, want 4 units worth Rp8.000 in B1", got)
	}
}

func TestGenerate_INS01_ConsignorIsAskedToRemindTheStore(t *testing.T) {
	snapshot := consignorSnapshot()
	snapshot.Products = []ProductStock{aProduct().withBatchExpiringIn(-1, 4).build()}

	got := onlyInsight(t, Generate(snapshot, today()))

	if got.Action != ActionRemindStore {
		t.Fatalf("action = %s, want REMIND_STORE", got.Action)
	}
}

func TestGenerate_INS01_EmptyExpiredBatchIsQuiet(t *testing.T) {
	snapshot := storeSnapshot()
	snapshot.Products = []ProductStock{aProduct().withBatchExpiringIn(-1, 0).build()}

	expectNone(t, Generate(snapshot, today()))
}

func TestGenerate_INS02_UnsoldEstimateMakesNearExpiryHigh(t *testing.T) {
	snapshot := storeSnapshot()
	product := aProduct().withShelfLife(3).withSoldLast14Days(14).withBatchExpiringIn(1, 10).build()
	snapshot.Products = []ProductStock{product}

	got := onlyInsight(t, Generate(snapshot, today()))

	if got.Kind != KindNearExpiry || got.Level != LevelHigh || got.Action != ActionCount {
		t.Fatalf("insight = %+v, want NEAR_EXPIRY HIGH COUNT", got)
	}
	if got.Units != 9 || got.Days != 1 || got.Amount != money.Rupiah(18000) {
		t.Fatalf("figures = %+v, want 9 unsold units in 1 day worth Rp18.000", got)
	}
}

func TestGenerate_INS02_SalesThatCoverTheBatchKeepItMedium(t *testing.T) {
	snapshot := consignorSnapshot()
	product := aProduct().withSoldLast14Days(28).withBatchExpiringIn(5, 10).build()
	snapshot.Products = []ProductStock{product}

	got := onlyInsight(t, Generate(snapshot, today()))

	if got.Level != LevelMedium || got.Units != 0 || got.Action != ActionRemindStore {
		t.Fatalf("insight = %+v, want MEDIUM with no unsold units and REMIND_STORE", got)
	}
}

func TestGenerate_INS02_BRSTK04_ShortShelfLifeWaitsUntilOneDayLeft(t *testing.T) {
	snapshot := storeSnapshot()
	product := aProduct().withShelfLife(3).withBatchExpiringIn(2, 10).build()
	snapshot.Products = []ProductStock{product}

	expectNone(t, Generate(snapshot, today()))
}

func TestGenerate_INS03_FewerThanThreeDaysOfStockIsMedium(t *testing.T) {
	snapshot := consignorSnapshot()
	snapshot.Products = []ProductStock{aProduct().withSoldLast14Days(14).withBatchWithoutExpiry(2).build()}

	got := onlyInsight(t, Generate(snapshot, today()))

	if got.Kind != KindLowStock || got.Level != LevelMedium || got.Action != ActionViewDropSuggestion {
		t.Fatalf("insight = %+v, want LOW_STOCK MEDIUM VIEW_DROP_SUGGESTION", got)
	}
	if got.Units != 2 || got.Days != 2 {
		t.Fatalf("figures = %+v, want 2 units lasting 2 days", got)
	}
}

func TestGenerate_INS03_ExactlyThreeDaysOfStockIsQuiet(t *testing.T) {
	snapshot := storeSnapshot()
	snapshot.Products = []ProductStock{aProduct().withSoldLast14Days(14).withBatchWithoutExpiry(3).build()}

	expectNone(t, Generate(snapshot, today()))
}

func TestGenerate_INS03_NoSalesIsQuiet(t *testing.T) {
	snapshot := storeSnapshot()
	snapshot.Products = []ProductStock{aProduct().withBatchWithoutExpiry(1).build()}

	expectNone(t, Generate(snapshot, today()))
}

func TestGenerate_INS03_YoungPartnershipAveragesOverItsOwnDays(t *testing.T) {
	snapshot := storeSnapshot()
	product := aProduct().withFirstDropDaysAgo(2).withSoldLast14Days(4).withBatchWithoutExpiry(5).build()
	snapshot.Products = []ProductStock{product}

	got := onlyInsight(t, Generate(snapshot, today()))

	if got.Kind != KindLowStock || got.Action != ActionAskRestock {
		t.Fatalf("insight = %+v, want LOW_STOCK ASK_RESTOCK from 2 units a day", got)
	}
}

func TestGenerate_INS06_UnpaidPastDueDateIsHigh(t *testing.T) {
	snapshot := consignorSnapshot()
	snapshot.Settlements = []OpenSettlement{aSettlement(-1, 31500)}

	got := onlyInsight(t, Generate(snapshot, today()))

	if got.Kind != KindSettlementOverdue || got.Level != LevelHigh || got.Action != ActionRemindPayment {
		t.Fatalf("insight = %+v, want SETTLEMENT_OVERDUE HIGH REMIND_PAYMENT", got)
	}
	if got.Days != 1 || got.Amount != money.Rupiah(31500) || got.Subject.SettlementNumber != "ST-202611-0001" {
		t.Fatalf("figures = %+v, want 1 day late for Rp31.500", got)
	}
}

func TestGenerate_INS06_StoreIsAskedToRecordPayment(t *testing.T) {
	snapshot := storeSnapshot()
	snapshot.Settlements = []OpenSettlement{aSettlement(-3, 31500)}

	if got := onlyInsight(t, Generate(snapshot, today())); got.Action != ActionRecordPayment {
		t.Fatalf("action = %s, want RECORD_PAYMENT", got.Action)
	}
}

func TestGenerate_INS06_DueTodayIsQuiet(t *testing.T) {
	snapshot := storeSnapshot()
	snapshot.Settlements = []OpenSettlement{aSettlement(0, 31500)}

	expectNone(t, Generate(snapshot, today()))
}

func TestGenerate_INS07_TwoDailyRhythmsWithoutCountIsLow(t *testing.T) {
	snapshot := storeSnapshot()
	snapshot.Partnerships = []PartnershipCadence{aPartnership(terms.RhythmDaily, 2, 8)}

	got := onlyInsight(t, Generate(snapshot, today()))

	if got.Kind != KindNotCounted || got.Level != LevelLow || got.Action != ActionCount || got.Days != 2 {
		t.Fatalf("insight = %+v, want NOT_COUNTED LOW COUNT after 2 days", got)
	}
}

func TestGenerate_INS07_ConsignorIsAskedToRemind(t *testing.T) {
	snapshot := consignorSnapshot()
	snapshot.Partnerships = []PartnershipCadence{aPartnership(terms.RhythmWeekly, 14, 8)}

	if got := onlyInsight(t, Generate(snapshot, today())); got.Action != ActionRemindCount {
		t.Fatalf("action = %s, want REMIND_COUNT", got.Action)
	}
}

func TestGenerate_INS07_QuietBeforeTwoRhythms(t *testing.T) {
	cases := map[terms.Rhythm]int{terms.RhythmDaily: 1, terms.RhythmWeekly: 13, terms.RhythmEachVisit: 13}
	for rhythm, days := range cases {
		t.Run(string(rhythm), func(t *testing.T) {
			snapshot := storeSnapshot()
			snapshot.Partnerships = []PartnershipCadence{aPartnership(rhythm, days, 8)}

			expectNone(t, Generate(snapshot, today()))
		})
	}
}

func TestGenerate_INS07_PartnershipWithoutStockIsQuiet(t *testing.T) {
	snapshot := storeSnapshot()
	snapshot.Partnerships = []PartnershipCadence{aPartnership(terms.RhythmEachVisit, 30, 0)}

	expectNone(t, Generate(snapshot, today()))
}

func TestGenerate_INS10_ScoreOfSeventyFiveSuggestsTheOrganization(t *testing.T) {
	snapshot := storeSnapshot()
	snapshot.Candidates = []Candidate{aCandidate(75)}

	got := onlyInsight(t, Generate(snapshot, today()))

	if got.Kind != KindMatchingPartner || got.Level != LevelLow || got.Action != ActionOpenProfile {
		t.Fatalf("insight = %+v, want MATCHING_PARTNER LOW OPEN_PROFILE", got)
	}
	if got.Subject.OrganizationID != "org-9" || got.Score != 75 {
		t.Fatalf("subject = %+v, want org-9 with score 75", got)
	}
}

func TestGenerate_INS10_QuietBelowThresholdOrWhenAlreadyEngaged(t *testing.T) {
	partner := aCandidate(90)
	partner.IsPartner = true
	proposed := aCandidate(90)
	proposed.HasOpenProposal = true
	snapshot := consignorSnapshot()
	snapshot.Candidates = []Candidate{aCandidate(74), partner, proposed}

	expectNone(t, Generate(snapshot, today()))
}

func TestGenerate_OrdersByLevelThenAffectedRupiah(t *testing.T) {
	snapshot := storeSnapshot()
	snapshot.Candidates = []Candidate{aCandidate(80)}
	snapshot.Products = []ProductStock{aProduct().withSoldLast14Days(14).withBatchWithoutExpiry(1).build()}
	smaller := aSettlement(-1, 10000)
	larger := aSettlement(-1, 50000)
	larger.ID = "settlement-2"
	snapshot.Settlements = []OpenSettlement{smaller, larger}

	got := Generate(snapshot, today())

	want := []string{"settlement-2", "settlement-1", string(KindLowStock), string(KindMatchingPartner)}
	if len(got) != len(want) {
		t.Fatalf("got %d insights, want %d: %+v", len(got), len(want), got)
	}
	for i, insight := range got {
		key := insight.Subject.SettlementID
		if key == "" {
			key = string(insight.Kind)
		}
		if key != want[i] {
			t.Fatalf("insight %d = %s, want %s", i, key, want[i])
		}
	}
}

func TestMarkOpenProposals_INS10_FlagsCandidatesAlreadyInAnOpenProposal(t *testing.T) {
	other := aCandidate(80)
	other.OrganizationID = "org-7"

	got := MarkOpenProposals([]Candidate{aCandidate(80), other}, []string{"org-9"})

	if !got[0].HasOpenProposal || got[1].HasOpenProposal {
		t.Fatalf("candidates = %+v, want only org-9 flagged", got)
	}
}
