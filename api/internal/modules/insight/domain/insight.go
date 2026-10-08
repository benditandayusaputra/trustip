package domain

import (
	"cmp"
	"slices"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

type Level string

const (
	LevelHigh   Level = "HIGH"
	LevelMedium Level = "MEDIUM"
	LevelLow    Level = "LOW"
)

type Kind string

const (
	KindExpired           Kind = "EXPIRED"
	KindNearExpiry        Kind = "NEAR_EXPIRY"
	KindLowStock          Kind = "LOW_STOCK"
	KindSettlementOverdue Kind = "SETTLEMENT_OVERDUE"
	KindNotCounted        Kind = "NOT_COUNTED"
	KindMatchingPartner   Kind = "MATCHING_PARTNER"
)

type Action string

const (
	ActionCountWithReturns   Action = "COUNT_WITH_RETURNS"
	ActionCount              Action = "COUNT"
	ActionRemindStore        Action = "REMIND_STORE"
	ActionViewDropSuggestion Action = "VIEW_DROP_SUGGESTION"
	ActionAskRestock         Action = "ASK_RESTOCK"
	ActionRemindPayment      Action = "REMIND_PAYMENT"
	ActionRecordPayment      Action = "RECORD_PAYMENT"
	ActionRemindCount        Action = "REMIND_COUNT"
	ActionOpenProfile        Action = "OPEN_PROFILE"
)

type Subject struct {
	PartnershipID    string
	CounterpartName  string
	ProductID        string
	ProductName      string
	BatchID          string
	BatchCode        string
	SettlementID     string
	SettlementNumber string
	OrganizationID   string
	OrganizationName string
}

type Insight struct {
	Kind    Kind
	Level   Level
	Action  Action
	Subject Subject
	Units   int
	Days    int
	Amount  money.Rupiah
	Score   int
}

type Snapshot struct {
	Side         actor.OrganizationType
	Products     []ProductStock
	Settlements  []OpenSettlement
	Partnerships []PartnershipCadence
	Candidates   []Candidate
}

func Generate(snapshot Snapshot, now time.Time) []Insight {
	today := now.In(period.Jakarta())
	insights := []Insight{}
	for _, product := range snapshot.Products {
		insights = append(insights, findExpiryInsights(product, today)...)
		insights = append(insights, findLowStock(product, today)...)
	}
	for _, settlement := range snapshot.Settlements {
		insights = append(insights, findOverdue(settlement, today)...)
	}
	for _, partnership := range snapshot.Partnerships {
		insights = append(insights, findNotCounted(partnership, today)...)
	}
	for _, candidate := range snapshot.Candidates {
		insights = append(insights, findMatchingPartner(candidate)...)
	}
	return ordered(withActions(insights, snapshot.Side))
}

func withActions(insights []Insight, side actor.OrganizationType) []Insight {
	actions := actionsFor(side)
	for i := range insights {
		insights[i].Action = actions[insights[i].Kind]
	}
	return insights
}

func actionsFor(side actor.OrganizationType) map[Kind]Action {
	if side == actor.Consignor {
		return map[Kind]Action{
			KindExpired: ActionRemindStore, KindNearExpiry: ActionRemindStore,
			KindLowStock: ActionViewDropSuggestion, KindSettlementOverdue: ActionRemindPayment,
			KindNotCounted: ActionRemindCount, KindMatchingPartner: ActionOpenProfile,
		}
	}
	return map[Kind]Action{
		KindExpired: ActionCountWithReturns, KindNearExpiry: ActionCount,
		KindLowStock: ActionAskRestock, KindSettlementOverdue: ActionRecordPayment,
		KindNotCounted: ActionCount, KindMatchingPartner: ActionOpenProfile,
	}
}

func ordered(insights []Insight) []Insight {
	slices.SortStableFunc(insights, func(left, right Insight) int {
		return cmp.Or(
			cmp.Compare(levelRank(left.Level), levelRank(right.Level)),
			cmp.Compare(right.Amount, left.Amount),
		)
	})
	return insights
}

func levelRank(level Level) int {
	return slices.Index([]Level{LevelHigh, LevelMedium, LevelLow}, level)
}
