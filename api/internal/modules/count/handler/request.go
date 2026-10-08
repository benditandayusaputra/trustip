package handler

import (
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
	ClientRecordedAt *string       `json:"client_recorded_at"`
	RecordedOffline  bool          `json:"recorded_offline"`
	Occasion         string        `json:"occasion"`
	Note             string        `json:"note"`
	Lines            []lineRequest `json:"lines"`
}

func (r countRequest) toInput() (domain.RecordInput, error) {
	var problems []apperror.FieldError
	if !uuidFormat.MatchString(r.PartnershipID) {
		problems = append(problems, apperror.InvalidField("partnership_id", apperror.FieldUnknownChoice))
	}
	countedAt, isCountedValid := claimed.Parse(r.CountedAt)
	if !isCountedValid {
		problems = append(problems, apperror.InvalidField("counted_at", apperror.FieldInvalidDate))
	}
	clientRecordedAt, isClientValid := claimed.Parse(r.ClientRecordedAt)
	if !isClientValid {
		problems = append(problems, apperror.InvalidField("client_recorded_at", apperror.FieldInvalidDate))
	}
	if len(problems) > 0 {
		return domain.RecordInput{}, apperror.Validation(problems...)
	}
	return domain.RecordInput{
		PartnershipID: r.PartnershipID, AgreementVersion: r.AgreementVersion, CountedAt: countedAt,
		ClientRecordedAt: clientRecordedAt, RecordedOffline: r.RecordedOffline, Occasion: r.Occasion, Note: r.Note,
		Lines: r.lineInputs(),
	}, nil
}

func (r countRequest) lineInputs() []domain.RequestLine {
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
	return lines
}
