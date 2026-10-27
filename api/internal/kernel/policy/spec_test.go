package policy_test

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
)

const (
	specPath          = "../../../../docs/02-SPESIFIKASI-FITUR.md"
	roleSectionTitle  = "## 1. Peran dan hak akses"
	roleColumnsInSpec = 4
)

func specRowActions() map[string]policy.Action {
	return map[string]policy.Action{
		"Ubah profil, lokasi, dan tampil di Cari Mitra":           policy.EditProfile,
		"Kelola anggota (P1)":                                     policy.ManageMembers,
		"Kelola produk":                                           policy.ManageProducts,
		"Pasang dan tutup Kebutuhan Toko":                         policy.ManageNeeds,
		"Kirim, tawar, terima, tolak Ajuan":                       policy.NegotiateProposals,
		"Usul dan setujui perubahan Kesepakatan":                  policy.AmendAgreements,
		"Catat Titipan Masuk":                                     policy.RecordDrops,
		"Umumkan kiriman (P1)":                                    policy.AnnounceShipments,
		"Konfirmasi Titipan Masuk dan Hitung Sisa":                policy.ConfirmDocuments,
		"Isi Hitung Sisa":                                         policy.RecordCounts,
		"Lihat rekap satu dokumen beserta nilainya":               policy.ViewDocumentValues,
		"Catat pembayaran Setoran":                                policy.RecordPayments,
		"Konfirmasi atau tolak pembayaran":                        policy.ReviewPayments,
		"Lihat laporan dan angka gabungan":                        policy.ViewReports,
		"Akhiri kemitraan":                                        policy.EndPartnerships,
		"Laporkan profil atau unggahan":                           policy.ReportListings,
		"Ubah akun sendiri: nama, WhatsApp, foto, dan kata sandi": policy.EditOwnAccount,
	}
}

func specGrant(t *testing.T, cell string) policy.Grant {
	t.Helper()
	grants := map[string]policy.Grant{
		"Ya":                   policy.Allowed,
		"Tidak":                policy.Denied,
		"Bila diizinkan OWNER": policy.DelegatedByOwner,
	}
	grant, ok := grants[cell]
	if !ok {
		t.Fatalf("unknown cell %q in the role table", cell)
	}
	return grant
}

type specRow struct {
	label string
	cells []string
}

func readRoleTable(t *testing.T) []specRow {
	t.Helper()
	file, err := os.Open(filepath.Clean(specPath))
	if err != nil {
		t.Fatalf("open spec: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close spec: %v", err)
		}
	}()
	var rows []specRow
	inSection := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			inSection = line == roleSectionTitle
		}
		if row, ok := parseRoleRow(line); inSection && ok {
			rows = append(rows, row)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read spec: %v", err)
	}
	return rows
}

func parseRoleRow(line string) (specRow, bool) {
	if !strings.HasPrefix(line, "| ") {
		return specRow{}, false
	}
	cells := strings.Split(strings.Trim(line, "| "), "|")
	if len(cells) != roleColumnsInSpec+1 {
		return specRow{}, false
	}
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	if cells[0] == "Aksi" {
		return specRow{}, false
	}
	return specRow{label: cells[0], cells: cells[1:]}, true
}

func TestTable_RSEC37_MatchesTheRoleTableInTheSpecRowByRow(t *testing.T) {
	rows := readRoleTable(t)
	actions := specRowActions()
	table := policy.Table()
	if len(rows) != len(actions) {
		t.Fatalf("spec has %d rows, test knows %d", len(rows), len(actions))
	}
	for _, row := range rows {
		action, known := actions[row.label]
		if !known {
			t.Errorf("spec row %q has no action in code", row.label)
			continue
		}
		want := policy.Grants{
			ConsignorOwner: specGrant(t, row.cells[0]),
			ConsignorStaff: specGrant(t, row.cells[1]),
			StoreOwner:     specGrant(t, row.cells[2]),
			StoreStaff:     specGrant(t, row.cells[3]),
		}
		if got := table[action]; got != want {
			t.Errorf("%s (%q) = %+v, spec says %+v", action, row.label, got, want)
		}
	}
}

func TestTable_RSEC37_HasNoActionMissingFromTheSpec(t *testing.T) {
	known := map[policy.Action]bool{}
	for _, action := range specRowActions() {
		known[action] = true
	}
	for action := range policy.Table() {
		if !known[action] {
			t.Errorf("action %s is not a row in the spec role table", action)
		}
	}
}
