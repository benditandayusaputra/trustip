package apperror

import "strings"

func requirementLabels() map[string]string {
	return map[string]string{
		"NAME":              "nama usaha",
		"CITY":              "kota",
		"LOCATION":          "lokasi",
		"STORE_TYPE":        "jenis toko",
		"ACCEPTED_CATEGORY": "kategori yang diterima",
		"ACTIVE_PRODUCT":    "produk aktif",
	}
}

func ProfileIncomplete(missing []string) *Error {
	labels := make([]string, 0, len(missing))
	for _, requirement := range missing {
		labels = append(labels, requirementLabels()[requirement])
	}
	return New(CodeProfileIncomplete).WithParam("daftar", strings.Join(labels, ", "))
}
