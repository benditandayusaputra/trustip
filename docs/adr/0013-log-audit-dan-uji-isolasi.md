# 0013. Log audit di transaksi peristiwa dan daftar uji isolasi

## Konteks

NFR-14 meminta log audit untuk masuk, perubahan profil, keputusan ajuan, perubahan kesepakatan, pembatalan, konfirmasi, dan pembayaran. Bagian 9 menetapkan bahwa jejak bisnis ditulis ke `audit_logs` di dalam transaksi peristiwanya. Peristiwa itu tersebar di banyak modul, sedangkan modul tidak boleh saling mengimpor (R-API-02). R-SEC-41 meminta setiap rute ber-ID punya uji isolasi, dengan uji pendaftar yang gagal bila ada yang terlewat. R-LOG-06 meminta akses lintas organisasi yang dibalas 404 dicatat sebagai peristiwa keamanan.

## Keputusan

1. Nilai `audit.Entry` beserta daftar aksinya tinggal di `kernel/audit`. Setiap modul mendefinisikan port `AuditTrail` di lapisan layanannya sendiri dan memanggilnya di dalam transaksi yang sama dengan perubahan. Modul `audit` hanya punya repositori.
2. ID entri dan IP diisi adaptor `auditTrail` di akar komposisi. IP diambil dari konteks permintaan yang diisi perantara `CarryClientIP` dengan IP klien yang sudah ikut ditandatangani proxy (R-SEC-24). Layanan tidak tahu soal HTTP.
3. `organization_id` boleh kosong, karena pengguna bisa masuk sebelum punya organisasi.
4. Penolakan karena data milik organisasi lain dibuat dengan `apperror.OutsideScope()`: jawabannya 404 yang sama persis dengan data yang tidak ada, tetapi penangan galat mencatat `cross_organization_access` di log keamanan. Kueri yang lingkupnya ditegakkan SQL tidak bisa membedakan keduanya dan tetap 404 biasa.
5. `isolationCases` di `cmd/server/isolation_test.go` memetakan setiap rute ber-ID ke permintaan dari organisasi lain bertipe sama. Uji tanpa basis data gagal bila ada rute ber-ID tanpa kasus, atau kasus tanpa rute. Uji basis data menjalankan semua kasus dan mengharapkan 404.

## Akibat

- Fitur baru yang menambah rute ber-ID tidak bisa lulus `make test` sebelum menambah kasus isolasinya.
- Daftar tampilan log audit untuk OWNER belum ada, karena belum ada kebutuhan yang memintanya.
