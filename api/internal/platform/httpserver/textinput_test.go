package httpserver_test

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

func TestLineInput_RSEC50_TrimsAndCollapsesSpaces(t *testing.T) {
	if got := httpserver.LineInput("name", "  Warung   Bu  Tini "); got.Value != "Warung Bu Tini" {
		t.Fatalf("value = %q", got.Value)
	}
}

func TestRejectUncleanInputs_RSEC51_NamesEveryFieldWithForbiddenCharacters(t *testing.T) {
	err := httpserver.RejectUncleanInputs(
		httpserver.LineInput("name", "Warung\u200bTini"),
		httpserver.LineInput("district", "Tembalang"),
		httpserver.LineInput("address", "Jl.\x07Sirojudin"),
	)

	fields := apperror.From(err).Fields()
	if len(fields) != 2 || fields[0].Field != "name" || fields[1].Field != "address" {
		t.Fatalf("fields = %+v", fields)
	}
}

func TestRejectUncleanInputs_CleanInputsPass(t *testing.T) {
	if err := httpserver.RejectUncleanInputs(httpserver.LineInput("name", "Warung Bu Tini")); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestMultilineInput_RSEC51_KeepsLineBreaksAndTidiesEachLine(t *testing.T) {
	got := httpserver.MultilineInput("description", "\r\n  Roti  manis \r\nTanpa   pengawet\n\n")

	if got.Value != "Roti manis\nTanpa pengawet" {
		t.Fatalf("value = %q", got.Value)
	}
}

func TestMultilineInput_RSEC51_OtherControlCharactersAreRejected(t *testing.T) {
	err := httpserver.RejectUncleanInputs(httpserver.MultilineInput("description", "Roti\tmanis"))

	if fields := apperror.From(err).Fields(); len(fields) != 1 || fields[0].Field != "description" {
		t.Fatalf("err = %v", err)
	}
}
