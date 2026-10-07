package apperror

import (
	"maps"
	"net/http"
)

type Code string

const (
	CodeValidationFailed     Code = "VALIDATION_FAILED"
	CodeUnknownField         Code = "UNKNOWN_FIELD"
	CodeMalformedRequest     Code = "MALFORMED_REQUEST"
	CodePayloadTooLarge      Code = "PAYLOAD_TOO_LARGE"
	CodeUnsupportedMediaType Code = "UNSUPPORTED_MEDIA_TYPE"
	CodeUnauthenticated      Code = "UNAUTHENTICATED"
	CodeSessionExpired       Code = "SESSION_EXPIRED"
	CodeLoginFailed          Code = "LOGIN_FAILED"
	CodeEdgeSignatureInvalid Code = "EDGE_SIGNATURE_INVALID"
	CodeForbidden            Code = "FORBIDDEN"
	CodeNotFound             Code = "NOT_FOUND"
	CodeCaptchaFailed        Code = "CAPTCHA_FAILED"
	CodeRateLimited          Code = "RATE_LIMITED"
	CodeAccountLocked        Code = "ACCOUNT_LOCKED"
	CodeServiceUnavailable   Code = "SERVICE_UNAVAILABLE"
	CodeInternal             Code = "INTERNAL"
)

type Kind int

const (
	KindValidation Kind = iota + 1
	KindUnauthenticated
	KindForbidden
	KindNotFound
	KindConflict
	KindRateLimited
	KindUnavailable
	KindInternal
)

type entry struct {
	kind    Kind
	status  int
	message string
}

func catalog() map[Code]entry {
	entries := requestEntries()
	maps.Copy(entries, accessEntries())
	maps.Copy(entries, outcomeEntries())
	return entries
}

func requestEntries() map[Code]entry {
	return map[Code]entry{
		CodeValidationFailed: {KindValidation, http.StatusUnprocessableEntity, "Periksa kembali isian Anda."},
		CodeUnknownField: {
			KindValidation, http.StatusBadRequest,
			"Permintaan memuat isian yang tidak dikenal.",
		},
		CodeMalformedRequest: {KindValidation, http.StatusBadRequest, "Permintaan tidak bisa dibaca."},
		CodePayloadTooLarge:  {KindValidation, http.StatusRequestEntityTooLarge, "Isi permintaan terlalu besar."},
		CodeUnsupportedMediaType: {
			KindValidation, http.StatusUnsupportedMediaType,
			"Jenis isi permintaan tidak didukung.",
		},
		CodeCaptchaFailed: {KindValidation, http.StatusUnprocessableEntity, "Verifikasi keamanan gagal. Coba lagi."},
	}
}

func accessEntries() map[Code]entry {
	return map[Code]entry{
		CodeUnauthenticated: {KindUnauthenticated, http.StatusUnauthorized, "Silakan masuk dulu."},
		CodeSessionExpired: {
			KindUnauthenticated, http.StatusUnauthorized,
			"Sesi Anda berakhir. Silakan masuk lagi.",
		},
		CodeEdgeSignatureInvalid: {KindUnauthenticated, http.StatusUnauthorized, "Permintaan tidak sah."},
		CodeLoginFailed:          {KindUnauthenticated, http.StatusUnauthorized, "Email atau kata sandi salah."},
		CodeAccountLocked: {
			KindRateLimited, http.StatusTooManyRequests,
			"Akun dikunci sementara karena terlalu banyak percobaan masuk. Coba lagi dalam {n} menit.",
		},
		CodeForbidden: {KindForbidden, http.StatusForbidden, "Peran Anda tidak diizinkan melakukan ini."},
		CodeNotFound: {
			KindNotFound, http.StatusNotFound,
			"Data tidak ditemukan atau Anda tidak punya akses.",
		},
	}
}

func outcomeEntries() map[Code]entry {
	return map[Code]entry{
		CodeRateLimited: {
			KindRateLimited, http.StatusTooManyRequests,
			"Terlalu banyak percobaan. Coba lagi dalam {n} menit.",
		},
		CodeServiceUnavailable: {
			KindUnavailable, http.StatusServiceUnavailable,
			"Layanan sedang sibuk. Coba lagi sebentar lagi.",
		},
		CodeInternal: {
			KindInternal, http.StatusInternalServerError,
			"Terjadi kesalahan di sistem. Kode: {request_id}.",
		},
	}
}

func Codes() []Code {
	codes := make([]Code, 0, len(catalog()))
	for code := range catalog() {
		codes = append(codes, code)
	}
	return codes
}

func lookup(code Code) entry {
	if found, ok := catalog()[code]; ok {
		return found
	}
	return catalog()[CodeInternal]
}
