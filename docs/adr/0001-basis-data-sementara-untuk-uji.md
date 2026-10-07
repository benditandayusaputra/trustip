# 0001. Basis data sementara untuk uji migrasi

## Konteks

R-API-08 dan R-SEC-56 menetapkan bahwa SQL hanya ada di `db/queries` dan `db/migrations`. Uji migrasi naik lalu turun memasang dan mencabut ekstensi yang berlaku untuk seluruh basis data, sehingga tidak boleh dijalankan di basis data uji bersama tempat uji integrasi lain berjalan bersamaan (R-TEST-06).

Nama basis data tidak bisa dikirim sebagai parameter kueri, sehingga `CREATE DATABASE` dan `DROP DATABASE` tidak bisa ditulis sebagai kueri sqlc.

Citra `postgis/postgis` tidak menyediakan varian arm64. Di laptop berprosesor Apple, kontainer pengembangan dijalankan dengan `platform: linux/amd64` lewat emulasi. CI berjalan di amd64 tanpa emulasi.

## Keputusan

1. Uji yang butuh basis data sendiri membuat basis data sementara bernama acak `trustip_test_<heksadesimal>` lalu menghapusnya di akhir uji. Hanya dua pernyataan ini yang ditulis di berkas Go, hanya di paket pembantu uji `internal/platform/database/dbtest` yang tidak diimpor kode produksi, dan nama basis data dilewatkan `pgx.Identifier.Sanitize`.
2. Semua kueri lain di uji, termasuk pemeriksaan ekstensi terpasang, memakai kueri sqlc.
3. Uji menolak berjalan bila `TEST_DATABASE_URL` kosong atau menunjuk ke basis data Neon (R-TEST-01).

## Akibat

- Uji migrasi bisa dijalankan bersamaan dengan uji integrasi lain tanpa saling mengganggu.
- Peran basis data uji lokal butuh hak `CREATEDB`. Kontainer pengembangan dan layanan CI memakai pengguna pemilik, jadi syarat ini terpenuhi.
- Pengecualian ini tidak berlaku untuk kode produksi.
