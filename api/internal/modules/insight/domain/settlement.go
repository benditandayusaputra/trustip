package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
)

type OpenSettlement struct {
	ID              string
	Number          string
	PartnershipID   string
	CounterpartName string
	DueOn           time.Time
	Outstanding     money.Rupiah
}

func findOverdue(settlement OpenSettlement, today time.Time) []Insight {
	daysLate := daysBetween(settlement.DueOn, today)
	if daysLate <= 0 {
		return nil
	}
	return []Insight{{
		Kind: KindSettlementOverdue, Level: LevelHigh, Days: daysLate, Amount: settlement.Outstanding,
		Subject: Subject{
			PartnershipID: settlement.PartnershipID, CounterpartName: settlement.CounterpartName,
			SettlementID: settlement.ID, SettlementNumber: settlement.Number,
		},
	}}
}
