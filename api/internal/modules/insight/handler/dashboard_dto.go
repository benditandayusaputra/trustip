package handler

import "github.com/benditandayusaputra/trustip/api/internal/modules/insight/domain"

type tallyResponse struct {
	Count  int   `json:"count"`
	Amount int64 `json:"amount"`
}

type figuresResponse struct {
	ActivePartners int           `json:"active_partners"`
	StockUnits     int           `json:"stock_units"`
	Unpaid         tallyResponse `json:"unpaid"`
	Overdue        tallyResponse `json:"overdue"`
	DueSoon        tallyResponse `json:"due_soon"`
}

type tasksResponse struct {
	DropsToConfirm    int `json:"drops_to_confirm"`
	CountsToConfirm   int `json:"counts_to_confirm"`
	PaymentsToReview  int `json:"payments_to_review"`
	PaymentsRejected  int `json:"payments_rejected"`
	ProposalsAwaiting int `json:"proposals_awaiting"`
	NeedsExpiring     int `json:"needs_expiring"`
}

type uncountedPartnerResponse struct {
	PartnershipID   string `json:"partnership_id"`
	CounterpartName string `json:"counterpart_name"`
	StockUnits      int    `json:"stock_units"`
}

type dashboardResponse struct {
	Figures    figuresResponse            `json:"figures"`
	Tasks      tasksResponse              `json:"tasks"`
	QuickCount []uncountedPartnerResponse `json:"quick_count"`
}

func toDashboard(dashboard domain.Dashboard) dashboardResponse {
	figures, tasks := dashboard.Figures, dashboard.Tasks
	return dashboardResponse{
		Figures: figuresResponse{
			ActivePartners: figures.ActivePartners, StockUnits: figures.StockUnits,
			Unpaid: toTally(figures.Unpaid), Overdue: toTally(figures.Overdue), DueSoon: toTally(figures.DueSoon),
		},
		Tasks: tasksResponse{
			DropsToConfirm: tasks.DropsToConfirm, CountsToConfirm: tasks.CountsToConfirm,
			PaymentsToReview: tasks.PaymentsToReview, PaymentsRejected: tasks.PaymentsRejected,
			ProposalsAwaiting: tasks.ProposalsAwaiting, NeedsExpiring: tasks.NeedsExpiring,
		},
		QuickCount: toUncountedPartners(dashboard.QuickCount),
	}
}

func toTally(tally domain.Tally) tallyResponse {
	return tallyResponse{Count: tally.Count, Amount: tally.Amount.Int64()}
}

func toUncountedPartners(partners []domain.UncountedPartner) []uncountedPartnerResponse {
	responses := make([]uncountedPartnerResponse, 0, len(partners))
	for _, partner := range partners {
		responses = append(responses, uncountedPartnerResponse{
			PartnershipID: partner.PartnershipID, CounterpartName: partner.CounterpartName,
			StockUnits: partner.StockUnits,
		})
	}
	return responses
}
