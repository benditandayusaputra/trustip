package apperror_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

func TestCatalog_EveryCodeHasStatusKindAndMessage(t *testing.T) {
	for _, code := range apperror.Codes() {
		err := apperror.New(code)
		if err.Status() < http.StatusBadRequest || err.Message() == "" || err.Kind() == 0 {
			t.Errorf("%s: status %d, kind %d, message %q", code, err.Status(), err.Kind(), err.Message())
		}
	}
}

func TestCatalog_StatusesFollowContractTable(t *testing.T) {
	cases := map[apperror.Code]int{
		apperror.CodeValidationFailed:     http.StatusUnprocessableEntity,
		apperror.CodeUnknownField:         http.StatusBadRequest,
		apperror.CodeMalformedRequest:     http.StatusBadRequest,
		apperror.CodePayloadTooLarge:      http.StatusRequestEntityTooLarge,
		apperror.CodeUnsupportedMediaType: http.StatusUnsupportedMediaType,
		apperror.CodeUnauthenticated:      http.StatusUnauthorized,
		apperror.CodeSessionExpired:       http.StatusUnauthorized,
		apperror.CodeEdgeSignatureInvalid: http.StatusUnauthorized,
		apperror.CodeForbidden:            http.StatusForbidden,
		apperror.CodeNotFound:             http.StatusNotFound,
		apperror.CodeCaptchaFailed:        http.StatusUnprocessableEntity,
		apperror.CodeLoginFailed:          http.StatusUnauthorized,
		apperror.CodeAccountLocked:        http.StatusTooManyRequests,
		apperror.CodeSessionRevoked:       http.StatusUnauthorized,
		apperror.CodeCrossSiteRejected:    http.StatusForbidden,
		apperror.CodeRateLimited:          http.StatusTooManyRequests,
		apperror.CodeServiceUnavailable:   http.StatusServiceUnavailable,
		apperror.CodeInternal:             http.StatusInternalServerError,
		apperror.CodeProposalLimitReached: http.StatusConflict,
		apperror.CodeProposalCooldown:     http.StatusConflict,
		apperror.CodeNeedLimitReached:     http.StatusConflict,
		apperror.CodeNeedNotOpen:          http.StatusConflict,
	}
	for code, want := range cases {
		if got := apperror.New(code).Status(); got != want {
			t.Errorf("%s status = %d, want %d", code, got, want)
		}
	}
}

func TestMessage_SpecMessagesAreUsed(t *testing.T) {
	cases := map[apperror.Code]string{
		apperror.CodeNotFound:           "Data tidak ditemukan atau Anda tidak punya akses.",
		apperror.CodeValidationFailed:   "Periksa kembali isian Anda.",
		apperror.CodeSessionExpired:     "Sesi Anda berakhir. Silakan masuk lagi.",
		apperror.CodeForbidden:          "Peran Anda tidak diizinkan melakukan ini.",
		apperror.CodeServiceUnavailable: "Layanan sedang sibuk. Coba lagi sebentar lagi.",
		apperror.CodeCaptchaFailed:      "Verifikasi keamanan gagal. Coba lagi.",
	}
	for code, want := range cases {
		if got := apperror.New(code).Message(); got != want {
			t.Errorf("%s message = %q, want %q", code, got, want)
		}
	}
}

func TestMessage_PlaceholderIsFilledFromParams(t *testing.T) {
	err := apperror.New(apperror.CodeInternal).WithParam("request_id", "0192f0c4")

	if got := err.Message(); got != "Terjadi kesalahan di sistem. Kode: 0192f0c4." {
		t.Fatalf("message = %q", got)
	}
}

func TestRateLimited_MessageRoundsRetryUpToMinutes(t *testing.T) {
	err := apperror.RateLimited(61)

	if err.Message() != "Terlalu banyak percobaan. Coba lagi dalam 2 menit." || err.RetryAfter() != 61 {
		t.Fatalf("message = %q, retry after = %d", err.Message(), err.RetryAfter())
	}
}

func TestValidation_CarriesEveryFieldError(t *testing.T) {
	err := apperror.Validation(
		apperror.FieldError{Field: "items[0].qty", Code: "MIN", Message: "Jumlah minimal 1."},
		apperror.FieldError{Field: "name", Code: "REQUIRED", Message: "Nama wajib diisi."},
	)

	if err.Code() != apperror.CodeValidationFailed || len(err.Fields()) != 2 {
		t.Fatalf("code = %s, fields = %v", err.Code(), err.Fields())
	}
}

func TestFrom_WrappedAppError_IsFound(t *testing.T) {
	wrapped := fmt.Errorf("get drop 7: %w", apperror.NotFound())

	if got := apperror.From(wrapped); got.Code() != apperror.CodeNotFound {
		t.Fatalf("code = %s", got.Code())
	}
}

func TestFrom_UnknownError_BecomesInternalKeepingCause(t *testing.T) {
	cause := errors.New("connection reset")

	got := apperror.From(cause)

	if got.Code() != apperror.CodeInternal || !errors.Is(got, cause) {
		t.Fatalf("code = %s, wraps cause = %v", got.Code(), errors.Is(got, cause))
	}
}

func TestError_TextNamesCodeAndCause(t *testing.T) {
	err := apperror.Internal(errors.New("boom"))

	if err.Error() != "INTERNAL: boom" {
		t.Fatalf("Error() = %q", err.Error())
	}
}

func TestWithParam_DoesNotChangeOriginal(t *testing.T) {
	original := apperror.New(apperror.CodeInternal)

	derived := original.WithParam("request_id", "x")

	if original.Message() == derived.Message() {
		t.Fatalf("original changed: %q", original.Message())
	}
}

func TestWithCause_KeepsCodeAndWrapsCause(t *testing.T) {
	cause := errors.New("dial tcp: connection refused")

	err := apperror.Unavailable(5).WithCause(cause)

	if err.Code() != apperror.CodeServiceUnavailable || !errors.Is(err, cause) || err.RetryAfter() != 5 {
		t.Fatalf("code = %s, wraps = %v, retry = %d", err.Code(), errors.Is(err, cause), err.RetryAfter())
	}
}

func TestAccountLocked_CarriesRetryAfterAndMinutes(t *testing.T) {
	err := apperror.AccountLocked(900)

	if err.RetryAfter() != 900 || err.Message() !=
		"Akun dikunci sementara karena terlalu banyak percobaan masuk. Coba lagi dalam 15 menit." {
		t.Fatalf("retry = %d, message = %q", err.RetryAfter(), err.Message())
	}
}
