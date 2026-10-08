package domain

import (
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const productID = "0192f0c4-0000-7000-8000-000000000101"

func fieldsOf(t *testing.T, err error) map[string]apperror.FieldCode {
	t.Helper()
	codes := map[string]apperror.FieldCode{}
	for _, field := range apperror.From(err).Fields() {
		codes[field.Field] = apperror.FieldCode(field.Code)
	}
	return codes
}

func TestNewReport_FRTRS04_AcceptsAReasonForAProduct(t *testing.T) {
	got, err := NewReport(ReportForm{TargetType: "PRODUCT", TargetID: productID, Reason: "MISLEADING"})

	if err != nil || got.Target != (Target{Type: TargetProduct, ID: productID}) || got.Reason != ReasonMisleading {
		t.Fatalf("report = %+v, err = %v", got, err)
	}
}

func TestNewReport_FRTRS04_CollectsEveryProblem(t *testing.T) {
	_, err := NewReport(ReportForm{TargetType: "USER", TargetID: "abc", Reason: "BORING"})

	codes := fieldsOf(t, err)
	if codes["target_type"] != apperror.FieldUnknownChoice || codes["target_id"] != apperror.FieldUnknownChoice ||
		codes["reason"] != apperror.FieldUnknownChoice {
		t.Fatalf("codes = %v", codes)
	}
}

func TestNewReport_FRTRS04_TargetAndReasonAreRequired(t *testing.T) {
	_, err := NewReport(ReportForm{})

	codes := fieldsOf(t, err)
	if codes["target_type"] != apperror.FieldRequired || codes["reason"] != apperror.FieldRequired {
		t.Fatalf("codes = %v", codes)
	}
}

func TestNewReport_FRTRS04_OtherReasonNeedsANote(t *testing.T) {
	_, err := NewReport(ReportForm{TargetType: "NEED", TargetID: productID, Reason: "OTHER"})

	if fieldsOf(t, err)["note"] != apperror.FieldRequired {
		t.Fatalf("err = %v, want note required", err)
	}
}

func TestNewReport_NoteIsAtMost500Characters(t *testing.T) {
	form := ReportForm{TargetType: "ORGANIZATION", TargetID: productID, Reason: "SPAM"}
	form.Note = strings.Repeat("é", MaxNoteLength)
	if _, err := NewReport(form); err != nil {
		t.Fatalf("500 characters: err = %v", err)
	}

	form.Note += "x"
	if fieldsOf(t, func() error { _, err := NewReport(form); return err }())["note"] != apperror.FieldTooLong {
		t.Fatal("501 characters must be too long")
	}
}
