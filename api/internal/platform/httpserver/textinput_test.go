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
