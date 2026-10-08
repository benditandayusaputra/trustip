package handler

import "github.com/benditandayusaputra/trustip/api/internal/modules/insight/domain"

type insightSubjectResponse struct {
	PartnershipID    *string `json:"partnership_id"`
	CounterpartName  string  `json:"counterpart_name"`
	ProductID        *string `json:"product_id"`
	ProductName      string  `json:"product_name"`
	BatchID          *string `json:"batch_id"`
	BatchCode        string  `json:"batch_code"`
	SettlementID     *string `json:"settlement_id"`
	SettlementNumber string  `json:"settlement_number"`
	OrganizationID   *string `json:"organization_id"`
	OrganizationName string  `json:"organization_name"`
}

type insightResponse struct {
	Kind    string                 `json:"kind"`
	Level   string                 `json:"level"`
	Action  string                 `json:"action"`
	Subject insightSubjectResponse `json:"subject"`
	Units   int                    `json:"units"`
	Days    int                    `json:"days"`
	Amount  int64                  `json:"amount"`
	Score   int                    `json:"score"`
}

func toInsights(insights []domain.Insight) []insightResponse {
	responses := make([]insightResponse, 0, len(insights))
	for _, insight := range insights {
		responses = append(responses, insightResponse{
			Kind: string(insight.Kind), Level: string(insight.Level), Action: string(insight.Action),
			Subject: toSubject(insight.Subject), Units: insight.Units, Days: insight.Days,
			Amount: insight.Amount.Int64(), Score: insight.Score,
		})
	}
	return responses
}

func toSubject(subject domain.Subject) insightSubjectResponse {
	return insightSubjectResponse{
		PartnershipID: optionalID(subject.PartnershipID), CounterpartName: subject.CounterpartName,
		ProductID: optionalID(subject.ProductID), ProductName: subject.ProductName,
		BatchID: optionalID(subject.BatchID), BatchCode: subject.BatchCode,
		SettlementID: optionalID(subject.SettlementID), SettlementNumber: subject.SettlementNumber,
		OrganizationID: optionalID(subject.OrganizationID), OrganizationName: subject.OrganizationName,
	}
}

func optionalID(id string) *string {
	if id == "" {
		return nil
	}
	return &id
}
