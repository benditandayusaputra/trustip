package handler

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/claimed"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
)

type lineRequest struct {
	ProductID    string  `json:"product_id"`
	RecordedQty  int     `json:"recorded_qty"`
	CountedQty   int     `json:"counted_qty"`
	ReturnedQty  int     `json:"returned_qty"`
	ReturnReason *string `json:"return_reason"`
}

type countRequest struct {
	PartnershipID    string        `json:"partnership_id"`
	AgreementVersion int           `json:"agreement_version"`
	CountedAt        *string       `json:"counted_at"`
	Occasion         string        `json:"occasion"`
	Note             string        `json:"note"`
	Lines            []lineRequest `json:"lines"`
}

func (r countRequest) toInput() (domain.RecordInput, error) {
	var problems []apperror.FieldError
	if !uuidFormat.MatchString(r.PartnershipID) {
		problems = append(problems, apperror.InvalidField("partnership_id", apperror.FieldUnknownChoice))
	}
	countedAt := claimed.Absent()
	if r.CountedAt != nil {
		parsed, err := time.Parse(time.RFC3339, *r.CountedAt)
		countedAt = claimed.At(parsed)
		if err != nil {
			problems = append(problems, apperror.InvalidField("counted_at", apperror.FieldInvalidDate))
		}
	}
	if len(problems) > 0 {
		return domain.RecordInput{}, apperror.Validation(problems...)
	}
	lines := make([]domain.RequestLine, 0, len(r.Lines))
	for _, line := range r.Lines {
		reason := ""
		if line.ReturnReason != nil {
			reason = *line.ReturnReason
		}
		lines = append(lines, domain.RequestLine{
			ProductID: line.ProductID, RecordedQty: line.RecordedQty, CountedQty: line.CountedQty,
			ReturnedQty: line.ReturnedQty, ReturnReason: reason,
		})
	}
	return domain.RecordInput{
		PartnershipID: r.PartnershipID, AgreementVersion: r.AgreementVersion, CountedAt: countedAt,
		Occasion: r.Occasion, Note: r.Note, Lines: lines,
	}, nil
}
