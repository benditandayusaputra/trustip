package apperror_test

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

func TestInvalidField_UsesCatalogMessageWithParams(t *testing.T) {
	got := apperror.InvalidField("password", apperror.FieldTooShort, apperror.Param{Name: "min", Value: "10"})

	want := apperror.FieldError{Field: "password", Code: "TOO_SHORT", Message: "Minimal 10 karakter."}
	if got != want {
		t.Fatalf("field error = %+v, want %+v", got, want)
	}
}

func TestInvalidField_EveryCodeHasMessage(t *testing.T) {
	for _, code := range apperror.FieldCodes() {
		if got := apperror.InvalidField("x", code); got.Message == "" {
			t.Errorf("%s has no message", code)
		}
	}
}
