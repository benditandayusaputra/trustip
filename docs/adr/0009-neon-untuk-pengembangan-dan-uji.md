# 0009. Neon dipakai untuk pengembangan dan uji

## Konteks

R-TEST-01 semula mewajibkan pengembangan dan semua uji memakai PostgreSQL lokal di kontainer, dan melarang uji menyentuh Neon. Tim memutuskan pada 7 Oktober 2026 bahwa pengembangan dan uji langsung memakai basis data Neon yang live. Data uji yang tercampur dengan data aplikasi diterima, karena masa pengerjaan masih panjang dan data contoh akan disiapkan di Tahap 12.

Neon menjalankan PostgreSQL 18.6 dengan PostGIS 3.6.4 dan pg_trgm 1.6. Kontainer lokal sebelumnya memakai PostgreSQL 17, sehingga uji di Neon juga lebih dekat ke keadaan produksi.

## Keputusan

1. `api/.env` memakai string koneksi Neon untuk `DATABASE_URL` dan `DATABASE_MIGRATE_URL`. `make dev`, `make migrate`, `make test`, dan `make e2e` membaca berkas itu. Kontainer PostgreSQL lokal dan `api/deploy/compose.dev.yaml` dihapus.
2. Uji integrasi tetap membuat basis data sementara `trustip_test_<acak>`, memigrasikannya, lalu menghapusnya, di server Neon yang sama. Basis data aplikasi tidak pernah dihapus atau dimigrasi turun oleh uji.
3. Uji ujung ke ujung dan `make dev` memakai basis data aplikasi. Akun dan sesi yang dibuat uji tersimpan di sana.
4. Pemuat konfigurasi API dan pembantu uji tidak lagi menolak host `neon.tech`. Aturan produksi tetap: `sslmode` minimal `require`, dan server menolak tersambung sebagai pemilik skema bila `APP_ENV=production`.
5. `cmd/migrate down` hanya boleh ke basis data yang namanya diawali `trustip_test_`. Migrasi turun ke basis data aplikasi ditolak di kode.
6. CI tetap memakai PostgreSQL di kontainer layanan sampai string koneksi Neon disimpan sebagai rahasia repositori.

## Akibat

- Uji lebih lambat karena setiap basis data sementara dibuat lewat jaringan: `make test` di API sekitar 30 detik.
- Basis data aplikasi akan berisi akun uji. Tahap 12 perlu membersihkannya sebelum data contoh dan akun demo diisi, dengan keputusan tim karena itu menghapus data live.
- Kata sandi pemilik skema dipakai di mesin pengembangan. Bila bocor, rotasi di panel Neon lalu perbarui `api/.env.production` dan `api/.env`.
- ADR 0001 tetap berlaku untuk cara membuat dan menghapus basis data sementara, tetapi servernya kini Neon, bukan kontainer lokal.
