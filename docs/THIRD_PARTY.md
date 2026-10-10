# Pustaka dan Layanan Pihak Ketiga

Setiap pustaka atau layanan pihak ketiga dicatat di sini dalam commit yang sama dengan pemakaian pertamanya (`03-ATURAN-REKAYASA.md` R-CC-52). Versi mengikuti berkas kunci (`api/go.sum`, `app/package-lock.json`).

## API (Go)

| Pustaka | Versi | Lisensi | Ukuran | Alasan dipakai |
|---|---|---|---|---|
| `github.com/gofiber/fiber/v3` | 3.5.0 | MIT | Biner server 8,3 MB setelah `-s -w` | Kerangka HTTP yang cepat dengan uji tanpa jaringan lewat `app.Test` |
| `github.com/jackc/pgx/v5` | 5.11.0 | MIT | Bersama goose, biner migrasi 9,8 MB setelah `-s -w` | Driver PostgreSQL dengan kumpulan koneksi dan dukungan `channel_binding` untuk Neon |
| `github.com/google/uuid` | 1.6.0 | BSD-3-Clause | Sudah ikut lewat Fiber | Pembangkit UUIDv7 yang terurut waktu untuk semua ID |
| `github.com/getkin/kin-openapi` | 0.149.0 | MIT | Hanya di uji | Memvalidasi jawaban nyata terhadap `openapi.yaml` di uji kontrak (R-RES-07) |
| `go.uber.org/goleak` | 1.3.0 | MIT | Hanya di uji | Memastikan tidak ada goroutine yang bocor setelah server berhenti (R-CC-72) |
| `github.com/pressly/goose/v3` | 3.28.0 | MIT | Lihat baris pgx | Migrasi bernomor dengan arah naik dan turun, dibaca dari berkas yang ditanam ke biner |
| `golang.org/x/crypto` | 0.57.0 | BSD-3-Clause | Hanya paket `argon2`, menambah 34 KB ke biner setelah `-s -w` | Hash kata sandi Argon2id (R-SEC-03). Pustaka standar Go belum menyediakan Argon2 |
| `golang.org/x/text` | 0.42.0 | BSD-3-Clause | Hanya paket `unicode/norm`, menambah 115 KB ke biner setelah `-s -w` | Normalisasi NFC untuk teks bebas dan NFKC untuk kata sandi (R-SEC-01, R-SEC-50). Pustaka standar Go belum menyediakannya |

## Alat pengembangan

Dipasang `make setup` ke `.bin/` dengan versi terkunci di `Makefile`. Tidak ikut ke biner atau bundel.

| Alat | Versi | Lisensi | Alasan dipakai |
|---|---|---|---|
| `golangci-lint` | 2.14.0 | GPL-3.0 | Menjalankan semua pemeriksa Go di tabel 3.13 dan format `gofumpt` |
| `gitleaks` | 8.30.1 | MIT | Pemindai rahasia di kait pra-commit dan di CI |
| `lefthook` | 2.1.17 | MIT | Pengelola kait git |
| `govulncheck` | 1.8.0 | BSD-3-Clause | Pemindai kerentanan dependensi Go |
| `deadcode` | 0.51.0 | BSD-3-Clause | Pencari kode mati |
| `sqlc` | 1.31.1 | MIT | Membangkitkan kode Go bertipe dari berkas SQL |

## Data yang dibundel

| Data | Sumber | Lisensi | Ukuran | Alasan dipakai |
|---|---|---|---|---|
| Daftar kata sandi umum (`api/internal/modules/auth/domain/common_passwords.txt`) | `100k-most-used-passwords-NCSC.txt` dari SecLists, disaring ke 10 sampai 128 karakter ASCII, huruf kecil, tanpa duplikat: 9.113 baris | MIT | 104 KB, ditanam ke biner | Menolak kata sandi yang paling sering bocor saat mendaftar (R-SEC-01) |
| Kabupaten dan kota (`api/internal/db/migrations/00004_cities.sql`) | Berkas `db/wilayah_level_1_2.sql` dari `github.com/cahyadsn/wilayah`, mengikuti Kepmendagri Nomor 300.2.2-2430 Tahun 2025. Diambil kode, nama, provinsi, dan titik tengah 514 kabupaten dan kota, tanpa poligon batas. Bujur Kabupaten Wakatobi (74.07) di sumber tertulis 23,53890 dan dikoreksi menjadi 123,53890 sesuai poligon batasnya | MIT | 61 KB, sebagai migrasi | Kode wilayah resmi untuk kolom kota organisasi dan titik tengah kota untuk Cari Mitra (FR-ORG-01, FR-DIR-04) |

## Huruf

Disajikan dari server sendiri, subset Latin, format WOFF2, diambil dari paket Fontsource 5.3.0. Berkas ada di `app/src/lib/assets/fonts/`.

| Huruf | Berkas | Lisensi | Ukuran | Alasan dipakai |
|---|---|---|---|---|
| Inter | 400, 600 | OFL-1.1 | 23 KB dan 24 KB | Huruf teks, mudah dibaca di layar kecil |
| JetBrains Mono | 500, 700 | OFL-1.1 | 21 KB dan 21 KB | Angka uang dan jumlah dengan lebar tetap |

## Layanan dan citra

| Layanan atau citra | Versi | Lisensi | Alasan dipakai |
|---|---|---|---|
| PostgreSQL di Neon | 18.6, PostGIS 3.6.4, pg_trgm 1.6 | PostgreSQL License | Basis data aplikasi yang live, dipakai juga untuk pengembangan dan uji (ADR 0009) |
| `postgis/postgis` | 17-3.5 | GPL-2.0 (PostGIS), PostgreSQL License | Basis data uji di CI selama CI belum punya akses ke Neon |
| Ubin peta OpenStreetMap (`tile.openstreetmap.org`) | Layanan ubin standar | Data ODbL, atribusi wajib tampil di peta | Latar peta di `MapView`. Tunduk pada kebijakan pemakaian ubin OSM. Untuk pemakaian besar, ganti ke penyedia ubin berbayar (spesifikasi 21.2) |
| Cloudflare Turnstile | API `v0` | Layanan gratis, tunduk pada ketentuan Cloudflare | Captcha pada daftar, masuk, dan akun demo. Widget dimuat hanya di halaman yang memakainya, token diverifikasi di server (R-SEC-09 sampai 13) |

## App (SvelteKit)

Ukuran untuk pustaka yang ikut ke peramban diukur dari bundel terkompresi. Alat pengembangan tidak ikut ke bundel.

| Pustaka | Versi | Lisensi | Ukuran | Alasan dipakai |
|---|---|---|---|---|
| `svelte` | 5.57 | MIT | Runtime kecil, terkompilasi per komponen | Kerangka antarmuka dengan rune untuk keadaan |
| `@sveltejs/kit` | 3.0 | MIT | Ikut bundel per rute | Rute, render server, proxy, service worker |
| `@sveltejs/adapter-vercel` | 7.0 | MIT | Alat bangun | Deploy app ke Vercel |
| `vite` | 8.3 | MIT | Alat bangun | Pembangun dan server pengembangan |
| `@lucide/svelte` | 1.52 | ISC | Hanya ikon yang diimpor, sekitar 1 KB per ikon | Satu keluarga ikon garis berlisensi terbuka (`DESIGN.md` bagian 7) |
| `leaflet`, `@types/leaflet` | 1.9.4, 1.9.22 | BSD-2-Clause, MIT | 42 KB JavaScript dan 4 KB CSS terkompresi, dimuat hanya saat peta tampil | Peta untuk menandai lokasi usaha. Dipakai hanya lewat komponen `MapView` (R-CC-50, R-APP-11) |
| `uqr` | 0.1.3 | MIT | 8 KB JavaScript terkompresi tanpa dependensi, dimuat hanya saat kode QR tampil | Membangkitkan kode QR produk sebagai SVG di peramban untuk ditempel di kemasan (FR-PRD-07). Dipakai hanya lewat komponen `QrCode` (R-CC-50, R-APP-11) |
| `tailwindcss`, `@tailwindcss/vite` | 4.3 | MIT | Hanya kelas yang dipakai | Gaya dari token desain |
| `typescript`, `svelte-check` | 6.0, 4.6 | Apache-2.0, MIT | Alat pengembangan | Pemeriksaan tipe ketat |
| `eslint`, `typescript-eslint`, `eslint-plugin-svelte`, `svelte-eslint-parser`, `@eslint/js`, `eslint-config-prettier`, `globals` | 10.12, 8.60, 3.23, 1.8, 10.0, 10.1, 17.6 | MIT | Alat pengembangan | Lint mutu, keamanan, dan batas lapisan |
| `prettier`, `prettier-plugin-svelte`, `prettier-plugin-tailwindcss` | 3.9, 4.1, 0.8 | MIT | Alat pengembangan | Format otomatis |
| `vitest`, `@vitest/browser-playwright` | 4.1 | MIT | Alat pengembangan | Uji unit dan komponen |
| `vitest-browser-svelte` | 3.1 | MIT | Alat pengembangan | Merender komponen Svelte di peramban nyata saat uji |
| `axe-core` | 4.14 | MPL-2.0 | Alat pengembangan | Pemeriksa aksesibilitas otomatis di uji komponen |
| `@playwright/test`, `playwright` | 1.63 | Apache-2.0 | Alat pengembangan | Uji ujung ke ujung di HP dan desktop |
| `lighthouse` | 13.5 | Apache-2.0 | Alat pengembangan | Mengukur skor performa, LCP, dan CLS di uji ujung ke ujung terhadap anggaran 10.1 |
| `knip` | 6.40 | ISC | Alat pengembangan | Pencari ekspor dan dependensi tak terpakai |
| `openapi-typescript` | 7.13 | MIT | Alat pengembangan | Membangkitkan `src/lib/api/schema.d.ts` dari `openapi.yaml`. Dipasang dengan `overrides` karena versi ini masih menulis TypeScript 5 sebagai rekan, padahal hanya memakai API pembangun sintaks yang sama di TypeScript 6 |
| `@types/node` | 24 | MIT | Alat pengembangan | Tipe Node untuk berkas konfigurasi |
