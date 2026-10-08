package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
)

const MaxNoteLength = 500

var idFormat = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type TargetType string

const (
	TargetOrganization TargetType = "ORGANIZATION"
	TargetProduct      TargetType = "PRODUCT"
	TargetNeed         TargetType = "NEED"
)

type Reason string

const (
	ReasonMisleading      Reason = "MISLEADING"
	ReasonProhibitedGoods Reason = "PROHIBITED_GOODS"
	ReasonOffensive       Reason = "OFFENSIVE"
	ReasonSpam            Reason = "SPAM"
	ReasonOther           Reason = "OTHER"
)

type ReportForm struct {
	TargetType string
	TargetID   string
	Reason     string
	Note       string
}

type Target struct {
	Type TargetType
	ID   string
}

type ReportDraft struct {
	Target Target
	Reason Reason
	Note   string
}

func NewReport(form ReportForm) (ReportDraft, error) {
	problems := field.Collect(
		requiredChoice("target_type", form.TargetType, targetTypes()),
		findTargetProblem(form.TargetID),
		requiredChoice("reason", form.Reason, reasons()),
		findNoteProblem(form),
	)
	if len(problems) > 0 {
		return ReportDraft{}, apperror.Validation(problems...)
	}
	return ReportDraft{
		Target: Target{Type: TargetType(form.TargetType), ID: form.TargetID},
		Reason: Reason(form.Reason), Note: form.Note,
	}, nil
}

func requiredChoice(name, value string, choices []string) field.Problem {
	if value == "" {
		return field.Invalid(name, apperror.FieldRequired)
	}
	return field.CheckChoice(name, value, choices)
}

func targetTypes() []string {
	return []string{string(TargetOrganization), string(TargetProduct), string(TargetNeed)}
}

func reasons() []string {
	return []string{
		string(ReasonMisleading), string(ReasonProhibitedGoods), string(ReasonOffensive), string(ReasonSpam),
		string(ReasonOther),
	}
}

func findTargetProblem(targetID string) field.Problem {
	if idFormat.MatchString(targetID) {
		return field.Problem{}
	}
	return field.Invalid("target_id", apperror.FieldUnknownChoice)
}

func findNoteProblem(form ReportForm) field.Problem {
	if form.Reason == string(ReasonOther) && form.Note == "" {
		return field.Invalid("note", apperror.FieldRequired)
	}
	return field.CheckLength("note", form.Note, field.Length{Max: MaxNoteLength})
}

var ErrTargetNotFound = errors.New("the reported target does not exist or belongs to the reporter")

type Filing struct {
	ID             string
	ReporterUserID string
	ReporterOrgID  string
	Draft          ReportDraft
	ReportedAt     time.Time
}

type Filed struct {
	ID     string
	Status string
}
