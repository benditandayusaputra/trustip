package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type vectorReason struct {
	Code         ReasonCode `json:"code"`
	DistanceM    int        `json:"distance_m"`
	CategoryName string     `json:"category"`
	OnTimeBPS    int        `json:"on_time_bps"`
}

type vectorStoreReputation struct {
	DueSettlements int `json:"due_settlements"`
	OnTimeBPS      int `json:"on_time_bps"`
}

type vectorConsignorReputation struct {
	ActivePartners int `json:"active_partners"`
	ProvenReports  int `json:"proven_reports"`
}

type vectorCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type matchCase struct {
	Name                string                     `json:"name"`
	DistanceM           *int                       `json:"distance_m"`
	DeliveryRadiusKm    int                        `json:"delivery_radius_km"`
	ConsignorCategories []vectorCategory           `json:"consignor_categories"`
	AcceptedCategoryIDs []string                   `json:"accepted_category_ids"`
	NeedCategoryIDs     []string                   `json:"need_category_ids"`
	PreferredStoreTypes []string                   `json:"preferred_store_types"`
	StoreType           string                     `json:"store_type"`
	StoreReputation     *vectorStoreReputation     `json:"store_reputation"`
	ConsignorReputation *vectorConsignorReputation `json:"consignor_reputation"`
	LastActiveAt        *time.Time                 `json:"last_active_at"`
	Score               int                        `json:"score"`
	Label               Label                      `json:"label"`
	Reasons             []vectorReason             `json:"reasons"`
}

type matchVectors struct {
	Now   time.Time   `json:"now"`
	Cases []matchCase `json:"cases"`
}

func loadMatchVectors(t *testing.T) matchVectors {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "testdata", "vectors", "match_score.json"))
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}
	var loaded matchVectors
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatalf("decode vectors: %v", err)
	}
	return loaded
}

func (c matchCase) facts(now time.Time) MatchFacts {
	facts := MatchFacts{
		DeliveryRadiusKm: c.DeliveryRadiusKm, AcceptedCategoryIDs: c.AcceptedCategoryIDs,
		NeedCategoryIDs: c.NeedCategoryIDs, PreferredStoreTypes: c.PreferredStoreTypes, StoreType: c.StoreType,
		Reputation: c.reputation(), Now: now,
	}
	if c.DistanceM != nil {
		facts.Distance = KnownDistance(*c.DistanceM)
	}
	if c.LastActiveAt != nil {
		facts.LastActiveAt = *c.LastActiveAt
	}
	for _, category := range c.ConsignorCategories {
		facts.ConsignorCategories = append(facts.ConsignorCategories, Category(category))
	}
	return facts
}

func (c matchCase) reputation() Reputation {
	if c.ConsignorReputation != nil {
		return ConsignorReputation(*c.ConsignorReputation)
	}
	return StoreReputation(*c.StoreReputation)
}

func (c matchCase) wantReasons() []Reason {
	reasons := []Reason{}
	for _, reason := range c.Reasons {
		reasons = append(reasons, Reason{
			Code: reason.Code, DistanceMeters: reason.DistanceM, CategoryName: reason.CategoryName,
			OnTimeBPS: reason.OnTimeBPS,
		})
	}
	return reasons
}

func TestScoreMatch_FRDIR10_SharedVectors(t *testing.T) {
	vectors := loadMatchVectors(t)
	for _, tc := range vectors.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			match := ScoreMatch(tc.facts(vectors.Now))

			isSame := match.Score == tc.Score && match.Label == tc.Label
			if !isSame || !reflect.DeepEqual(match.Reasons, tc.wantReasons()) {
				t.Fatalf("match = %+v, want score %d label %q reasons %+v", match, tc.Score, tc.Label, tc.wantReasons())
			}
		})
	}
}

func TestScoreMatch_FRDIR10_SectionFourExampleIsSeventySeven(t *testing.T) {
	match := ScoreMatch(MatchFacts{
		Distance:            KnownDistance(5000),
		ConsignorCategories: []Category{{ID: "snacks", Name: "Kue kering dan camilan"}},
		AcceptedCategoryIDs: []string{"snacks"},
		Reputation:          StoreReputation{DueSettlements: 12, OnTimeBPS: 9600},
		LastActiveAt:        time.Date(2026, 10, 19, 3, 0, 0, 0, time.UTC),
		Now:                 time.Date(2026, 10, 20, 3, 0, 0, 0, time.UTC),
	})

	if match.Score != 77 || match.Label != VeryGoodMatch {
		t.Fatalf("match = %+v", match)
	}
}

func TestScoreMatch_LabelBoundaries(t *testing.T) {
	cases := map[int]Label{74: GoodMatch, 75: VeryGoodMatch, 49: NoLabel, 50: GoodMatch}
	for score, want := range cases {
		if got := labelOf(score); got != want {
			t.Errorf("labelOf(%d) = %q, want %q", score, got, want)
		}
	}
}

func TestStoreReputation_BRTRS01_NewUntilThreeDueSettlements(t *testing.T) {
	if !(StoreReputation{DueSettlements: 2, OnTimeBPS: 10000}).IsNew() {
		t.Fatal("two due settlements should still be new")
	}
	if (StoreReputation{DueSettlements: 3}).IsNew() {
		t.Fatal("three due settlements should be rated")
	}
}

func TestStoreReputation_OnTimeShareRoundsToTheNearestPoint(t *testing.T) {
	cases := map[int]int{9499: 9, 9500: 10, 0: 0, 10000: 10}
	for onTime, want := range cases {
		got := StoreReputation{DueSettlements: 3, OnTimeBPS: onTime}.component().points
		if got != want {
			t.Errorf("on time %d bps = %d points, want %d", onTime, got, want)
		}
	}
}
