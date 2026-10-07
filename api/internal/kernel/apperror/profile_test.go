package apperror_test

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

func TestProfileIncomplete_BRORG05_NamesWhatIsMissing(t *testing.T) {
	err := apperror.ProfileIncomplete([]string{"STORE_TYPE", "ACCEPTED_CATEGORY"})

	if err.Message() != "Lengkapi profil dulu: jenis toko, kategori yang diterima." || err.Status() != 422 {
		t.Fatalf("message = %q, status = %d", err.Message(), err.Status())
	}
}
