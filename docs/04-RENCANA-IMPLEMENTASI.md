# Rencana Implementasi: Trustip

| Item | Keterangan |
|---|---|
| Versi | 1.0 |
| Tanggal | 7 Oktober 2026 |
| Dokumen pasangan | `01-PRD.md`, `02-SPESIFIKASI-FITUR.md`, `03-ATURAN-REKAYASA.md`, `DESIGN.md` |

Dokumen ini memecah seluruh P0 di spesifikasi menjadi 94 fitur berurutan dalam 13 tahap. Satu baris di tabel fitur sama dengan satu commit. Setiap tahap ditutup gerbang yang harus lulus sebelum tahap berikutnya dimulai.

Tidak ada tanggal di dokumen ini. Urutan dan gerbang yang mengikat.

## Daftar isi

1. Prasyarat
2. Variabel lingkungan
3. Protokol kerja per fitur
4. Kapan berhenti dan melapor
5. Peta tahap
6. Rincian tahap dan daftar fitur
7. P1 setelah seluruh P0 lulus
8. Gerbang akhir

## 1. Prasyarat

Disiapkan tim sebelum fitur pertama dikerjakan. Pelaksana tidak membuat akun atau rahasia produksi sendiri. Rahasia untuk pengembangan lokal dibangkitkan `make setup` ke berkas `.env` yang tidak dilacak git.

| # | Yang disiapkan | Dipakai di |
|---|---|---|
| 1 | Repositori GitHub kosong dan kunci SSH yang bisa push. Remote: `git@github.com:benditandayusaputra/trustip.git` | Tahap 0 |
| 2 | Identitas git lokal (`user.name`, `user.email`) milik anggota tim | Tahap 0 |
| 3 | Go, Node.js, dan Docker terpasang di mesin kerja | Tahap 0 |
| 4 | Proyek Neon dengan dua string koneksi: pemilik skema dan (setelah dibuat di Tahap 12) peran aplikasi | Tahap 0 dan 12 |
| 5 | Widget Cloudflare Turnstile untuk host app: kunci situs dan kunci rahasia | Tahap 3 |
| 6 | Rahasia produksi acak 32 byte: kunci JWT, pepper kata sandi, kunci HMAC tepi, benih kunci segel | Tahap 12 |
| 7 | VPS dengan Docker, satu subdomain untuk API, port 80 dan 443 terbuka | Tahap 12 |
| 8 | Proyek Vercel yang terhubung ke repositori, direktori akar `app` | Tahap 12 |

Catatan basis data:

- Basis data Neon adalah basis data **aplikasi yang live**, dan juga dipakai untuk pengembangan dan semua uji (`03-ATURAN-REKAYASA.md` R-TEST-01, ADR 0009). `api/.env` memakai string koneksi yang sama dengan `DATABASE_MIGRATE_URL` di `api/.env.production`, dengan nilai diapit tanda kutip tunggal karena memuat `&`. CI tetap memakai PostgreSQL di kontainer layanan sampai CI diberi akses ke Neon.
- String koneksi dari panel Neon biasanya memakai host berakhiran `-pooler`. Untuk migrasi, pakai host langsung, yaitu nama host yang sama tanpa `-pooler`. Server API juga memakai host langsung karena ia punya kumpulan koneksinya sendiri.
- Bila driver menolak parameter `channel_binding` pada string koneksi, pakai versi driver yang mendukungnya. Jangan menurunkan `sslmode` di bawah `require`.
- String koneksi hanya ditulis di berkas `.env` yang tidak dilacak git. Repositori ini publik.

## 2. Variabel lingkungan

Hanya nama. Nilainya tidak pernah masuk repositori. `api/.env.example` dan `app/.env.example` memuat daftar yang sama tanpa nilai.

**`api`**

| Variabel | Isi |
|---|---|
| `APP_ENV` | `development`, `test`, atau `production` |
| `HTTP_ADDR` | Alamat dengar server |
| `PUBLIC_APP_ORIGIN` | Asal app, untuk pemeriksaan asal dan CORS aliran SSE |
| `DATABASE_URL` | Koneksi peran aplikasi |
| `DATABASE_MIGRATE_URL` | Koneksi pemilik skema, hanya untuk migrasi |
| `JWT_SECRET_V1`, `JWT_ACTIVE_KEY_ID` | Kunci tanda tangan token akses |
| `PASSWORD_PEPPER_V1`, `PASSWORD_PEPPER_ACTIVE` | Pepper kata sandi |
| `EDGE_HMAC_KEY_V1`, `EDGE_HMAC_ACTIVE` | Kunci tanda tangan permintaan dari proxy |
| `SEAL_PRIVATE_KEY_V1`, `SEAL_ACTIVE_KEY_ID` | Benih Ed25519 untuk segel dokumen |
| `TURNSTILE_SECRET_KEY`, `TURNSTILE_EXPECTED_HOSTNAME` | Verifikasi captcha |
| `FILE_STORAGE_DIR`, `PUBLIC_FILES_BASE_URL` | Penyimpanan dan alamat berkas |
| `LOG_LEVEL` | Tingkat log |

**`app`**

| Variabel | Isi |
|---|---|
| `API_INTERNAL_URL` | Alamat API yang dipanggil proxy |
| `EDGE_HMAC_KEY_V1`, `EDGE_HMAC_ACTIVE` | Sama dengan milik API |
| `PUBLIC_APP_ORIGIN` | Asal app |
| `PUBLIC_STREAM_URL` | Alamat aliran SSE di API |
| `PUBLIC_FILES_BASE_URL` | Alamat berkas publik di API |
| `PUBLIC_TURNSTILE_SITE_KEY` | Kunci situs captcha |

## 3. Protokol kerja per fitur

Sepuluh langkah yang sama untuk setiap baris di tabel fitur.

1. **Baca.** Baris fiturnya, lalu bagian spesifikasi dan aturan yang dirujuk. Bukan seluruh dokumen.
2. **Kontrak.** Bila ada rute baru atau berubah, tulis dulu di `api/openapi/openapi.yaml`.
3. **Domain.** Tulis uji yang gagal, lalu fungsi murninya. Contoh wajib di spesifikasi menjadi vektor uji.
4. **Data.** Migrasi dan kueri, lalu bangkitkan kode.
5. **Layanan dan repositori.** Uji layanan lebih dulu (sukses, tiap galat, tiap penolakan peran), lalu kodenya, lalu uji integrasi.
6. **Penangan.** DTO, pemetaan, rute, lalu uji kontrak, uji peran, dan uji isolasi.
7. **App.** Bangkitkan tipe, lalu skema formulir, keadaan, komponen, dan enam keadaan layar. Uji komponen, dan uji ujung ke ujung bila fitur menyentuh sebuah alur.
8. **Periksa.** `make check` dan `make test`. Baca ulang perubahan terhadap daftar periksa di `03-ATURAN-REKAYASA.md` bagian 14.
9. **Commit dan push.** Pesan mengikuti kolom "Pesan commit" dan boleh disesuaikan bila isi nyatanya berbeda. Tambahkan baris `Rujukan:`. Lalu `git push origin main`.
10. **Lanjut** ke baris berikutnya.

Aturan tambahan:

- Posisi kerja ditentukan dari `git log --oneline` dibandingkan tabel fitur. Tidak ada berkas kemajuan terpisah.
- Fitur yang ternyata terlalu besar untuk satu commit yang rapi dipecah menjadi dua baris berurutan, masing-masing lulus pemeriksaan sendiri.
- Perbaikan atas fitur yang sudah di-commit dibuat sebagai commit baru (🐛), bukan dengan menulis ulang riwayat.
- P1 dan P2 tidak disentuh sebelum gerbang akhir lulus, walau terlihat mudah.
- Tidak ada fitur di luar spesifikasi. Gagasan baru dicatat sebagai usulan, tidak dikerjakan.

## 4. Kapan berhenti dan melapor

Pekerjaan berjalan terus dari satu fitur ke fitur berikutnya, kecuali pada salah satu keadaan ini. Saat itu pekerjaan berhenti, keadaannya dilaporkan, dan tidak ada jalan pintas yang diambil.

1. Ada rahasia, akun, atau akses dari bagian 1 yang belum tersedia dan dibutuhkan fitur berikutnya.
2. Dua dokumen bertentangan dan urutan otoritas di `03-ATURAN-REKAYASA.md` tidak menyelesaikannya.
3. Sebuah gerbang tahap tetap gagal setelah tiga kali perbaikan.
4. Satu-satunya cara maju adalah melanggar sebuah aturan `R-`.
5. Sebuah perintah akan menjalankan migrasi turun, `DROP`, `TRUNCATE`, atau `DELETE` massal terhadap basis data aplikasi di Neon. Basis data uji sementara `trustip_test_*` dikecualikan.
6. Layanan luar yang dibutuhkan tidak bisa dijangkau.

## 5. Peta tahap

| Tahap | Isi | Fitur | Gerbang ringkas |
|---|---|---:|---|
| 0 | Kerangka repositori dan pagar mutu | 6 | Klon baru lulus `make setup`, `make check`, `make test`. CI hijau |
| 1 | Fondasi API | 8 | Tanpa tanda tangan tepi ditolak. Semua jawaban berbentuk baku |
| 2 | Fondasi app | 8 | Halaman contoh memanggil API lewat proxy. Bisa dipasang. Tanpa pelanggaran CSP |
| 3 | Akun dan sesi | 7 | Daftar, masuk, segarkan, keluar lulus ujung ke ujung. Token tidak terbaca skrip |
| 4 | Organisasi, lokasi, hak akses | 8 | Dua organisasi beda tipe terdaftar. Uji isolasi lulus |
| 5 | Produk, Ajuan undangan, Kesepakatan, Kemitraan, Segel | 9 | Dua akun bermitra, isi sama di dua layar, segel sah |
| 6 | Titipan Masuk, Stok, Hitung Sisa | 8 | Titip 20, sisa 2, rekap benar sampai rupiah |
| 7 | Setoran, pembayaran, Laporan | 8 | Tiga jadwal menerbitkan Setoran. Laporan cocok dengan Setoran |
| 8 | Kerja luring | 6 | Catatan luring terkirim tepat satu kali. Bentrok bisa ditinjau |
| 9 | Cari Mitra, Kebutuhan Toko, tawar-menawar | 7 | Alur A dan B lulus. Cari terdekat di bawah 200 ms |
| 10 | Dasbor, saran, reputasi | 6 | Tiap saran muncul pada data pemicunya |
| 11 | Pemberitahuan langsung, halaman publik, verifikasi segel | 6 | Pemberitahuan lintas peramban di bawah 3 detik. Lighthouse memenuhi anggaran |
| 12 | Data contoh, pengerasan, deploy, dokumentasi | 7 | Situs live dengan akun demo. Daftar periksa keamanan terisi |

Alat kelola (tahap 5 sampai 7) dikerjakan sebelum direktori (tahap 9) supaya inti demo aman lebih dulu. Kerja luring (tahap 8) menyusul segera setelah alur catatnya utuh.

## 6. Rincian tahap dan daftar fitur

Kolom "Rujukan" menyebut ID di `02-SPESIFIKASI-FITUR.md` (`FR-`, `BR-`, `NFR-`, `INS-`) dan di `03-ATURAN-REKAYASA.md` (`R-`).

### Tahap 0. Kerangka repositori dan pagar mutu

Baca dulu: aturan bagian 1, 2, 3.13, 12.

| # | Fitur | Rujukan | Pesan commit |
|---|---|---|---|
| 0.1 | Repositori dengan folder `api`, `app`, `docs`, berkas abai yang menutup semua `.env` kecuali `.env.example`, `.editorconfig`, README kerangka, dan kelima dokumen | R-REPO-01, R-SEC-73 | 🎉 Mulai repositori Trustip dengan dokumen produk dan aturan rekayasa |
| 0.2 | Kerangka API: modul Go, `cmd/server` dengan `/healthz`, semua pemeriksa di tabel 3.13, pemeriksa tanpa komentar, uji arsitektur | R-CC-27 sampai 32, R-API-01 | 🧱 Siapkan kerangka API Go dengan pemeriksa mutu |
| 0.3 | Kerangka app: SvelteKit, TypeScript ketat, Tailwind, semua pemeriksa di tabel 3.13, Vitest, Playwright | R-APP-01, NFR-15 | 🧱 Siapkan kerangka app SvelteKit dengan pemeriksa mutu |
| 0.4 | Kait git (pra-commit, bentuk pesan, pra-push) dan Makefile akar dengan `setup` (termasuk membangkitkan `.env` pengembangan), `dev`, `check`, `test`, `build`, `gen`, `e2e` | R-REPO-03, R-GIT-03, R-GIT-05, R-SEC-75 | 🔧 Pasang kait git dan perintah satu pintu di Makefile |
| 0.5 | Alur CI: pemeriksaan, uji dengan PostgreSQL ber-PostGIS, pemindai rahasia, pemindai kerentanan, pemeriksa berkas hasil pembangkit | R-REPO-04, R-SEC-76 | 👷 Tambah alur CI untuk lint, uji, dan pemindaian |
| 0.6 | PostgreSQL lokal di kontainer, alat migrasi, pembangkit kueri, migrasi pertama yang memasang PostGIS dan pg_trgm, dan perintah uji sambung ke Neon yang hanya membaca versi dan ketersediaan ekstensi | R-API-11, R-TEST-01 | 🗃️ Siapkan PostgreSQL lokal, migrasi, dan pembangkit kueri |

**Gerbang:** klon baru lulus `make setup`, `make check`, dan `make test` tanpa langkah manual. Commit berpesan salah bentuk ditolak kait. Berkas berkomentar ditolak `make check`. CI hijau. Uji sambung ke Neon berhasil dan kedua ekstensi tersedia.

### Tahap 1. Fondasi API

Baca dulu: aturan bagian 4, 5, 7.6, 7.9, 9.

| # | Fitur | Rujukan | Pesan commit |
|---|---|---|---|
| 1.1 | Pemuat konfigurasi yang memvalidasi rahasia, ditambah jam dan pembangkit ID yang bisa ditiru | R-SEC-73, R-SEC-74, R-API-03 | 🔧 Tambah pemuat konfigurasi yang menolak rahasia kosong |
| 1.2 | Pencatat log terstruktur dengan penyamaran bidang sensitif dan ujinya | R-LOG-01 sampai 04, NFR-10 | 🔊 Tambah pencatat log terstruktur dengan penyamaran data sensitif |
| 1.3 | Jenis galat, katalog kode dan pesan, penulis jawaban baku, penangkap `panic`, nomor permintaan | R-CC-43, R-RES-08 sampai 14 | 🥅 Tambah jenis galat dan bentuk jawaban baku API |
| 1.4 | Server HTTP dengan rantai perantara: kepala keamanan, batas isi, batas waktu, log akses, henti tertib, rute `GET /api/v1/meta` | R-API-07, R-LOG-03, R-LOG-05, R-CC-72 | 🏗️ Rakit server HTTP dengan rantai perantara dan henti tertib |
| 1.5 | Kumpulan koneksi, pengelola transaksi, pelacak kueri lambat, `/readyz` | R-API-04, R-PERF-03, R-PERF-04, R-LOG-07, R-LOG-10 | 🗃️ Tambah kumpulan koneksi, pengelola transaksi, pelacak kueri lambat |
| 1.6 | Verifikasi tanda tangan HMAC dari proxy: string kanonik, jendela waktu, nonce sekali pakai | R-SEC-23 sampai 26 | 🔒️ Verifikasi tanda tangan HMAC dari proxy di API |
| 1.7 | Pembatas laju ember token per IP dan per rute dengan kepala baku | R-SEC-43, tabel 7.9 | 🔒️ Tambah pembatas laju per IP dan per rute |
| 1.8 | Kontrak OpenAPI dasar (amplop, galat), uji kontrak atas jawaban nyata, pembangkit tipe untuk app | R-RES-07, R-RES-14 | 📝 Tambah kontrak OpenAPI dengan uji kesesuaian jawaban |

**Gerbang:** permintaan tanpa tanda tangan, bertanda tangan kedaluwarsa, atau bernonce ulang ditolak 401. Semua jawaban, termasuk 404, 429, dan 500, lolos uji kontrak. Uji membuktikan log tidak memuat rahasia. Uji arsitektur lulus.

### Tahap 2. Fondasi app

Baca dulu: `DESIGN.md` seluruhnya, aturan bagian 6, 7.13, 7.14, spesifikasi bagian 24.

| # | Fitur | Rujukan | Pesan commit |
|---|---|---|---|
| 2.1 | Token desain, huruf yang disajikan sendiri, gaya dasar, pemeriksa warna mentah | `DESIGN.md` 2, 4, 10, R-APP-08 | 💄 Terapkan token desain gelap dan huruf dari DESIGN.md |
| 2.2 | Komponen kontrol: tombol, isian teks, isian angka, isian uang, pilihan, sakelar | `DESIGN.md` 3, NFR-11 | 💄 Tambah komponen kontrol formulir |
| 2.3 | Komponen wadah dan umpan balik: kartu, kartu angka kunci, lencana, kotak peringatan, toast, dialog, lembar bawah, kerangka muat, keadaan kosong, tab, tabel yang menjadi kartu | `DESIGN.md` 3, spesifikasi 24 | 💄 Tambah komponen wadah dan umpan balik |
| 2.4 | Kerangka tata letak publik dan aplikasi: navigasi bawah, bilah samping, halaman galat | NFR-01, `DESIGN.md` 5 | 📱 Tambah kerangka tata letak publik dan aplikasi |
| 2.5 | Kepala keamanan dan kebijakan keamanan konten berbasis hash | R-SEC-67, R-SEC-68, tabel 7.14 | 🔒️ Pasang kepala keamanan dan CSP di app |
| 2.6 | Proxy `/api` satu asal yang menandatangani permintaan dan menyaring kepala | R-APP-13 sampai 15, aturan 7.6 | 🔒️ Tambah proxy satu asal bertanda tangan HMAC |
| 2.7 | Klien API (buka amplop, galat bertipe, segarkan sesi satu jalur, kunci idempotensi), katalog teks, format rupiah, tanggal, dan jarak | R-APP-04, R-APP-09, R-APP-12, NFR-12 | 🏗️ Tambah klien API, katalog teks, dan format lokal |
| 2.8 | Manifes, ikon, pintasan, service worker untuk cangkang dan aset, halaman `/luring`, pemberitahuan versi baru | FR-PWA-01, FR-PWA-02, FR-PWA-05, BR-PWA-01 | ✨ Jadikan app bisa dipasang dengan cangkang tersimpan |

**Gerbang:** halaman contoh menampilkan hasil `GET /api/v1/meta` lewat proxy. Aplikasi bisa dipasang. Tidak ada pelanggaran CSP di konsol. Semua komponen lulus pemeriksa aksesibilitas. JavaScript awal di bawah anggaran.

### Tahap 3. Akun dan sesi

Baca dulu: spesifikasi bagian 2, aturan 7.2 sampai 7.5.

| # | Fitur | Rujukan | Pesan commit |
|---|---|---|---|
| 3.1 | Tabel pengguna dan sesi. Hash kata sandi Argon2id dengan salt per kata sandi dan pepper server, daftar kata sandi umum, batas hash bersamaan | BR-AUTH-04, R-SEC-01 sampai 07, NFR-05 | 🔒️ Tambah hash kata sandi Argon2id dengan salt dan pepper |
| 3.2 | Verifikasi captcha di server di balik port, dan komponen isiannya di app | FR-AUTH-12, BR-AUTH-05, R-SEC-09 sampai 13 | 🔒️ Tambah verifikasi captcha di server dan komponen isiannya |
| 3.3 | Daftar: validasi, persetujuan privasi, halaman `/daftar` | FR-AUTH-01, FR-AUTH-08 | ✨ Tambah pendaftaran akun dengan persetujuan privasi |
| 3.4 | Masuk: token akses dan penyegar di cookie server, penguncian setelah gagal berulang, pesan seragam, halaman `/masuk` | FR-AUTH-02, BR-AUTH-01, R-SEC-08, R-SEC-14 sampai 16, R-SEC-45 | ✨ Tambah masuk dengan sesi di cookie HttpOnly |
| 3.5 | Segarkan sesi dengan rotasi dan deteksi pakai ulang, keluar, keluar dari semua perangkat | FR-AUTH-03, R-SEC-17 sampai 20 | 🔒️ Tambah rotasi token penyegar, keluar, dan keluar semua perangkat |
| 3.6 | Perantara autentikasi, pemeriksa asal dan kepala klien, penjaga area `/app` dengan parameter `lanjut` yang aman | FR-AUTH-06, R-SEC-21, R-SEC-22, R-SEC-49 | 🛂 Jaga area aplikasi dan tolak permintaan lintas situs |
| 3.7 | Halaman kebijakan privasi dan ketentuan | FR-PUB-09, R-SEC-13 | 📝 Tambah halaman kebijakan privasi dan ketentuan |

**Gerbang:** daftar, masuk, segarkan, dan keluar lulus uji ujung ke ujung. Uji membuktikan token tidak terbaca skrip dan tidak ada di penyimpanan peramban. Lima kali gagal mengunci akun. Pemakaian ulang token penyegar mencabut seluruh keluarga sesi. Permintaan tulis dari asal lain ditolak.

### Tahap 4. Organisasi, lokasi, hak akses

Baca dulu: spesifikasi bagian 1, 3, 16, aturan 7.8, 7.15.

| # | Fitur | Rujukan | Pesan commit |
|---|---|---|---|
| 4.1 | Tabel organisasi, keanggotaan, kategori, dan profil. Pelaku, tabel izin peran, dan uji yang membandingkannya dengan spesifikasi | BR-AUTHZ-01, BR-AUTHZ-03, BR-AUTHZ-05, R-SEC-36, R-SEC-37 | 🛂 Tambah organisasi, keanggotaan, dan tabel izin peran |
| 4.2 | Onboarding tiga langkah yang membuat organisasi dan keanggotaan dalam satu transaksi, dengan slug otomatis | FR-AUTH-04, FR-AUTH-05, BR-AUTH-02, BR-AUTH-03, FR-ORG-02, BR-ORG-01, BR-ORG-02 | ✨ Tambah onboarding tipe usaha, profil, dan lokasi |
| 4.3 | Lokasi: "Gunakan lokasi saya", geser pin di peta, lokasi publik yang dibulatkan untuk penitip | FR-ORG-03, BR-ORG-03, BR-ORG-04 | ✨ Tambah penentuan lokasi dengan pembulatan lokasi publik |
| 4.4 | Unggah gambar: dikecilkan di peramban, dikode ulang di server, penyimpanan di luar basis data, sajian publik dan privat | BR-ORG-06, NFR-09, R-SEC-54, R-SEC-70 sampai 72 | ✨ Tambah unggah gambar yang dikode ulang di server |
| 4.5 | Pengaturan profil dasar, profil toko, profil penitip, dan info pembayaran penitip | FR-ORG-01, FR-ORG-04, FR-ORG-05, FR-ORG-07, BR-ORG-07 | ✨ Tambah pengaturan profil toko dan penitip |
| 4.6 | Sakelar "Tampil di Cari Mitra" dan meter kelengkapan profil | FR-ORG-06, FR-ORG-08, BR-ORG-05 | ✨ Tambah sakelar tampil di Cari Mitra dan meter kelengkapan profil |
| 4.7 | Log audit, dan kerangka uji isolasi yang gagal bila ada rute ber-ID tanpa uji | NFR-07, NFR-14, BR-AUTHZ-02, R-SEC-38 sampai 41 | 🛂 Tambah log audit dan kerangka uji isolasi antar organisasi |
| 4.8 | Pemberitahuan tersimpan: lonceng, jumlah belum dibaca, tandai dibaca, tautan ke dokumen | FR-NTF-01, FR-NTF-04, FR-NTF-05, BR-NTF-01 sampai 03 | ✨ Tambah pemberitahuan di dalam aplikasi |

**Gerbang:** satu penitip dan satu toko bisa mendaftar sampai dasbor kosong. Uji isolasi lulus untuk semua rute ber-ID. Tabel izin di kode sama dengan spesifikasi bagian 1. Foto yang diunggah tidak lagi memuat data lokasi.

### Tahap 5. Produk, Ajuan undangan, Kesepakatan, Kemitraan, Segel

Baca dulu: spesifikasi bagian 6 sampai 9 dan 18.1, aturan 7.7.

| # | Fitur | Rujukan | Pesan commit |
|---|---|---|---|
| 5.1 | Katalog produk: buat, ubah, arsip, SKU otomatis, cari dan saring | FR-PRD-01, FR-PRD-02, FR-PRD-03, FR-PRD-05, BR-PRD-01, BR-PRD-03, BR-PRD-04 | ✨ Tambah katalog produk penitip |
| 5.2 | Kode QR per produk dan sakelar "Sedang mencari toko" | FR-PRD-04, FR-PRD-07 | ✨ Tambah kode QR produk dan tanda sedang mencari toko |
| 5.3 | Fungsi murni model harga: Harga Setor, Komisi, pembulatan, contoh hitungan untuk pratinjau, dengan vektor uji bersama | FR-AGR-02, BR-AGR-01 sampai 04, R-TEST-03, R-TEST-04 | ✨ Tambah hitungan Harga Setor dan Komisi sebagai fungsi murni |
| 5.4 | Ajuan lewat undangan: tabel ajuan, revisi, kemitraan, dan kesepakatan. Formulir dengan syarat per produk dan contoh hitungan, tautan `/undangan/<kode>`, klaim oleh tipe berlawanan | FR-PRP-02, FR-PRP-07, FR-PRP-09, BR-PRP-01, BR-PRP-07, BR-PRP-09 | ✨ Tambah Ajuan Kerja Sama lewat tautan undangan |
| 5.5 | Segel: bentuk kanonik, hash isi, entri berantai per kemitraan, tanda tangan Ed25519, tabel yang hanya bisa ditambah | FR-SIG-01, BR-SIG-01 sampai 06, R-SEC-27 sampai 32 | 🔒️ Tambah segel dokumen bertanda tangan Ed25519 |
| 5.6 | Keputusan Ajuan: terima, tolak, tarik. Menerima membentuk Kemitraan dan Kesepakatan versi 1 dalam satu transaksi yang tersegel. Daftar Ajuan masuk dan keluar | FR-PRP-03, FR-PRP-05, FR-PRP-06, BR-PRP-02 sampai 05, BR-PRT-01 | ✨ Tambah keputusan Ajuan yang membentuk Kemitraan dan Kesepakatan |
| 5.7 | Halaman Kesepakatan: versi aktif, tiga angka per produk, jejak persetujuan, tanda segel | FR-AGR-01, FR-AGR-03, FR-SIG-02, R-SEC-33 | ✨ Tambah halaman Kesepakatan dengan jejak persetujuan tersegel |
| 5.8 | Daftar mitra, halaman mitra bertab, katalog mitra beserta syaratnya untuk toko | FR-PRT-01, FR-PRT-02, FR-PRT-05, FR-PRD-06 | ✨ Tambah daftar mitra dan halaman kemitraan |
| 5.9 | Ajuan perubahan, riwayat versi dengan beda antar versi, jalur cepat tambah produk | FR-PRP-08, FR-AGR-04, FR-AGR-05, FR-AGR-06, BR-AGR-05, BR-AGR-07 | ✨ Tambah perubahan Kesepakatan berversi |

**Gerbang:** dua akun bermitra lewat tautan undangan dan melihat isi Kesepakatan yang sama. Segel sah. Uji integrasi yang mengubah satu harga langsung di basis data membuat halaman menampilkan "Segel tidak sah". Dua keputusan bersamaan atas satu Ajuan hanya satu yang lolos.

### Tahap 6. Titipan Masuk, Stok, Hitung Sisa

Baca dulu: spesifikasi bagian 10 sampai 13 (rumus), 26.

| # | Fitur | Rujukan | Pesan commit |
|---|---|---|---|
| 6.1 | Fungsi murni stok: saldo batch dari mutasi, alokasi FEFO, tanda batch | BR-STK-01 sampai 04, vektor FEFO | ✨ Tambah alokasi stok FEFO sebagai fungsi murni |
| 6.2 | Fungsi murni hitung: terjual, tetap dititip, dikembalikan, nilai per alokasi dengan pembulatan | BR-CNT-01, BR-CNT-02, BR-CNT-08, BR-CNT-09, BR-STL-04, vektor bagian 12 dan 13 | ✨ Tambah hitungan Hitung Sisa dan nilai per alokasi |
| 6.3 | Catat Titipan Masuk: formulir per produk dalam Kesepakatan, kedaluwarsa otomatis, pratinjau, nomor dokumen, stok bertambah dalam transaksi yang sama, salinan syarat ke batch, kunci idempotensi, pemberitahuan, segel | FR-DRP-01, FR-DRP-02, FR-DRP-04, FR-DRP-05, BR-DRP-01 sampai 07, BR-PRD-02, BR-AGR-06, BR-STK-03 | ✨ Tambah pencatatan Titipan Masuk dengan kunci idempotensi |
| 6.4 | Penitip menekan Sesuai, dianggap sesuai setelah 48 jam, toko membatalkan dalam 15 menit | FR-DRP-06, FR-DRP-07, FR-DRP-10, BR-DRP-08 | ✨ Tambah konfirmasi dan pembatalan Titipan Masuk |
| 6.5 | Riwayat Titipan Masuk bersaring dan "Ulangi titipan terakhir" | FR-DRP-08, FR-DRP-09 | ✨ Tambah riwayat Titipan Masuk dan ulangi titipan terakhir |
| 6.6 | Stok per mitra per batch, gabungan penitip, gabungan toko, kartu stok dengan tautan dokumen | FR-STK-01 sampai 05 | ✨ Tambah tampilan stok dan kartu stok per batch |
| 6.7 | Hitung Sisa: isi sisa atau terjual, dikembalikan beserta alasan, keterangan waktu, pratinjau saat mengetik, tombol cepat, rekap bernomor, kunci baris stok, segel | FR-CNT-01 sampai 07, BR-CNT-03 sampai 07 | ✨ Tambah Hitung Sisa dengan rekap omzet, untung toko, dan harus setor |
| 6.8 | Penitip menyatakan hitungan sesuai, dianggap sesuai setelah 48 jam, riwayat Hitung Sisa | FR-CNT-08, FR-CNT-10 | ✨ Tambah konfirmasi dan riwayat Hitung Sisa |

**Gerbang:** titip 20 roti, sisa 2 dan dikembalikan 2, rekap menampilkan terjual 18, omzet Rp36.000, harus setor Rp31.500, untung toko Rp4.500. Semua vektor uji lulus di Go dan TypeScript. Cakupan domain 90% ke atas. Dua Hitung Sisa bersamaan hanya satu yang lolos. Simpan dua kali hanya membuat satu dokumen.

### Tahap 7. Setoran, pembayaran, Laporan

Baca dulu: spesifikasi bagian 13, 14, 21.5.

| # | Fitur | Rujukan | Pesan commit |
|---|---|---|---|
| 7.1 | Fungsi murni Setoran: batas periode mingguan dan bulanan di Asia/Jakarta, ringkasan, status turunan | BR-STL-01 sampai 03, BR-STL-07, BR-STL-09 | ✨ Tambah periode dan ringkasan Setoran sebagai fungsi murni |
| 7.2 | Setoran terbit bersama Hitung Sisa untuk jadwal `ON_COUNT`. Halaman Setoran, daftar bersaring umur, penjualan "berjalan", tombol lanjutan setelah rekap | FR-STL-01, FR-STL-03, FR-STL-04, FR-STL-09, FR-STL-10, FR-CNT-09 | ✨ Terbitkan Setoran otomatis saat Hitung Sisa |
| 7.3 | Penjadwal di dalam proses: Setoran mingguan dan bulanan, penetapan "dianggap sesuai", penanda kedaluwarsa | FR-STL-02, spesifikasi 21.5, R-CC-70, R-CC-72 | ✨ Tambah penjadwal untuk Setoran berkala dan tugas harian |
| 7.4 | Pembayaran: catat bayar dengan bukti, uang diterima atau tolak, tandai lunas oleh penitip, ingatkan lewat WhatsApp | FR-STL-05 sampai 08, BR-STL-05, BR-STL-06, BR-STL-08, FR-NTF-06, R-SEC-35 | ✨ Tambah pencatatan dan konfirmasi pembayaran Setoran |
| 7.5 | Mengakhiri kemitraan dengan daftar penghalang, dan sakelar tampil sebagai tempat membeli | FR-PRT-03, FR-PRT-04, BR-PRT-02 sampai 04 | ✨ Tambah pengakhiran kemitraan dengan daftar penghalang |
| 7.6 | Laporan penjualan toko dan penitip: pengelompokan, ringkasan, rentang cepat | FR-RPT-02, FR-RPT-03, FR-RPT-04, FR-RPT-07, BR-RPT-01 sampai 04 | ✨ Tambah laporan penjualan toko dan penitip |
| 7.7 | Rekap harian toko dan laporan Setoran | FR-RPT-01, FR-RPT-06 | ✨ Tambah rekap harian toko dan laporan Setoran |
| 7.8 | Grafik area penjualan harian dengan tampilan tabel, dan ekspor CSV yang aman dari rumus | FR-RPT-05, FR-RPT-08, R-SEC-55, `DESIGN.md` 3 | ✨ Tambah grafik penjualan harian dan ekspor CSV |

**Gerbang:** ketiga jadwal menerbitkan Setoran, yang berkala lewat penjadwal dengan jam tiruan. Tabel contoh spesifikasi bagian 13 lulus sampai rupiah. Total laporan sama dengan total Setoran pada rentang yang sama. Alur C lulus ujung ke ujung.

### Tahap 8. Kerja luring

Baca dulu: spesifikasi bagian 19, aturan bagian 8.

| # | Fitur | Rujukan | Pesan commit |
|---|---|---|---|
| 8.1 | Simpanan baca di perangkat untuk daftar dan dokumen terakhir, pita luring, hapus saat keluar | FR-PWA-04, BR-PWA-01, BR-PWA-02 | ✨ Tampilkan data terakhir saat luring |
| 8.2 | Paket kerja per kemitraan aktif tersimpan di perangkat saat daring | FR-PWA-08, R-OFF-01, R-OFF-04 | ✨ Simpan paket kerja kemitraan di perangkat |
| 8.3 | API menerima catatan luring: waktu catat perangkat dan waktu terima, tanda luring, bentrok versi kesepakatan, jendela waktu dihitung dari waktu terima | FR-PWA-10, BR-PWA-06, BR-PWA-07, R-OFF-09, R-OFF-11 | ✨ Terima catatan luring di API dengan waktu catat dan waktu terima |
| 8.4 | Antrean luring dan mesin sinkron: satu jalur tulis, urut per kemitraan, satu pengirim antar tab, coba ulang berjeda, terikat ke pengguna | FR-PWA-06, BR-PWA-03 sampai 05, R-OFF-02, R-OFF-03, R-OFF-05 sampai 08, R-OFF-13, R-OFF-14 | ✨ Tambah antrean catat luring untuk Titipan Masuk dan Hitung Sisa |
| 8.5 | Halaman Antrean: status, kirim sekarang, tinjau bentrok, buang, pilihan saat keluar, tanda "dicatat luring" di dokumen | FR-PWA-09, FR-PWA-10, BR-PWA-08, R-OFF-10, R-OFF-12, R-OFF-16 | ✨ Tambah halaman Antrean dan tinjauan catatan yang bentrok |
| 8.6 | Ajakan pasang setelah aksi pertama, petunjuk untuk iOS, permintaan penyimpanan persisten, peringatan entri lama | FR-PWA-03, R-OFF-15 | 🚸 Tambah ajakan pasang aplikasi setelah aksi pertama |

**Gerbang:** uji ujung ke ujung dengan jaringan dimatikan: toko mencatat satu Titipan Masuk dan satu Hitung Sisa, jaringan dinyalakan, tepat satu dokumen masing-masing tersimpan di server. Pengiriman ganda tidak menggandakan. Stok yang berubah di perangkat lain membuat entri berstatus perlu ditinjau. Masuk sebagai pengguna lain tidak mengirim entri pengguna sebelumnya.

### Tahap 9. Cari Mitra, Kebutuhan Toko, tawar-menawar

Baca dulu: spesifikasi bagian 4, 5, 6.

| # | Fitur | Rujukan | Pesan commit |
|---|---|---|---|
| 9.1 | Fungsi murni skor kecocokan dengan alasan singkat | FR-DIR-10, vektor skor 77 | ✨ Tambah skor kecocokan mitra sebagai fungsi murni |
| 9.2 | Direktori toko: kartu, urut terdekat dengan indeks spasial, titik asal, saring jarak dengan perluasan, saringan toko, urutan lain, cari teks, saringan di alamat halaman | FR-DIR-01, FR-DIR-03 sampai 06, FR-DIR-08, FR-DIR-09, FR-DIR-12, BR-DIR-01 sampai 06 | ✨ Tambah direktori toko terurut dari yang terdekat |
| 9.3 | Direktori penitip dengan tab Produk dan Penitip serta saringannya | FR-DIR-02, FR-DIR-07 | ✨ Tambah direktori penitip dan produk yang mencari toko |
| 9.4 | "Ajukan kerja sama" dan "Tanya lewat WhatsApp" dari kartu, profil, dan produk. Batas ajuan harian dan jeda setelah ditolak | FR-DIR-11, FR-PRP-01, BR-PRP-06, R-SEC-44 | ✨ Tambah Ajuan langsung dari Cari Mitra |
| 9.5 | Tawar-menawar: revisi, perpindahan giliran, penanda baris berubah, masa berlaku dihitung ulang | FR-PRP-04, BR-PRP-08 | ✨ Tambah tawar-menawar Ajuan dengan penanda perubahan |
| 9.6 | Kebutuhan Toko: pasang, masa tayang, perpanjang, batas lima terbuka, tanggapan, tandai terpenuhi atau tutup | FR-NED-01, FR-NED-02, FR-NED-06, BR-NED-01 sampai 04 | ✨ Tambah Kebutuhan Toko |
| 9.7 | Papan Kebutuhan, tanda di kartu dan profil toko, tanggapan penitip lewat Ajuan, pemberitahuan ke penitip yang cocok, pengingat masa tayang | FR-NED-03, FR-NED-04, FR-NED-05, FR-NED-07 | ✨ Tambah Papan Kebutuhan dan tanggapan penitip |

**Gerbang:** Alur A dan Alur B lulus ujung ke ujung. Pencarian terdekat di bawah 200 ms pada 10.000 organisasi buatan dan rencana kuerinya memakai indeks spasial. Contoh skor 77 lulus.

### Tahap 10. Dasbor, saran, reputasi

Baca dulu: spesifikasi bagian 15, 18.

| # | Fitur | Rujukan | Pesan commit |
|---|---|---|---|
| 10.1 | Mesin saran sebagai fungsi murni untuk kedaluwarsa, mendekati kedaluwarsa, stok segera habis, Setoran lewat tempo, belum dihitung, dan mitra yang cocok | INS-01, INS-02, INS-03, INS-06, INS-07, INS-10 | ✨ Tambah mesin saran sebagai fungsi murni |
| 10.2 | Saran jumlah titip dan tampilannya di formulir Titipan Masuk | INS-09, FR-DRP-03 | ✨ Tambah saran jumlah titip di formulir Titipan Masuk |
| 10.3 | Dasbor penitip: angka kunci, perlu tindakan, grafik, peringkat, saran, peluang | Spesifikasi 15.1, `DESIGN.md` 5 | ✨ Tambah dasbor penitip |
| 10.4 | Dasbor toko dengan Hitung cepat | Spesifikasi 15.2, FR-CNT-11 | ✨ Tambah dasbor toko dengan Hitung cepat |
| 10.5 | Reputasi: ringkasan harian, angka toko dan penitip, tanda "Baru" | FR-TRS-01, FR-TRS-02, FR-TRS-03, BR-TRS-01, BR-TRS-02 | ✨ Tambah angka reputasi dari data transaksi |
| 10.6 | Tombol Laporkan pada profil, produk, dan Kebutuhan. Perilaku organisasi yang ditangguhkan | FR-TRS-04, BR-TRS-04, BR-TRS-05, BR-AUTHZ-06 | ✨ Tambah pelaporan profil, produk, dan Kebutuhan |

**Gerbang:** setiap saran P0 punya satu uji yang memicunya dan satu yang tidak. Contoh saran jumlah titip 19 lulus. Dasbor memenuhi anggaran LCP pada data contoh.

### Tahap 11. Pemberitahuan langsung, halaman publik, verifikasi segel

Baca dulu: spesifikasi bagian 16, 17, 18.1.

| # | Fitur | Rujukan | Pesan commit |
|---|---|---|---|
| 11.1 | Aliran SSE bertiket sekali pakai, hub, denyut, putus saat tab tersembunyi, periksa berkala sebagai cadangan | FR-NTF-02, FR-NTF-03, BR-NTF-04, BR-NTF-05, R-SEC-26, R-PERF-07 | ✨ Tambah pemberitahuan langsung lewat SSE bertiket |
| 11.2 | Beranda yang tidak bergantung pada basis data | FR-PUB-01, NFR-17 | ✨ Tambah beranda |
| 11.3 | Profil publik toko dan penitip dengan daftar putih data | FR-PUB-02, FR-PUB-03, BR-PUB-02, BR-AUTHZ-04, R-SEC-40 | ✨ Tambah profil publik toko dan penitip |
| 11.4 | Halaman produk publik dan "Tersedia di" | FR-PUB-04, FR-PUB-05, BR-PUB-01 | ✨ Tambah halaman produk publik dengan daftar toko yang menjual |
| 11.5 | Halaman Kebutuhan publik, direktori dan Papan Kebutuhan tanpa masuk, tombol bagikan, gambar pratinjau, cache tepi | FR-PUB-06, FR-PUB-07, FR-PUB-08, FR-NED-08, BR-PUB-03, R-PERF-08 | ✨ Buka direktori dan Kebutuhan untuk pengunjung |
| 11.6 | Halaman verifikasi segel publik, rute kunci publik, pemeriksaan rantai harian | FR-SIG-03, FR-SIG-04, FR-SIG-05, R-SEC-34 | ✨ Tambah halaman verifikasi segel publik |

**Gerbang:** pemberitahuan sampai ke peramban lain dalam 3 detik. Lighthouse memenuhi anggaran pada beranda, direktori, profil, dan halaman produk. Uji membuktikan jawaban publik tidak memuat kolom di luar daftar putih. Alur E lulus.

### Tahap 12. Data contoh, pengerasan, deploy, dokumentasi

Baca dulu: spesifikasi bagian 21.8, 25, 26, aturan bagian 7, 10, 13.

| # | Fitur | Rujukan | Pesan commit |
|---|---|---|---|
| 12.1 | Data contoh 30 hari, akun demo, tombol "Coba akun demo", setel ulang malam | FR-AUTH-07, spesifikasi 25 | 🌱 Tambah data contoh dan akun demo |
| 12.2 | Uji keamanan menyeluruh (XSS, suntikan, lintas situs, isolasi) dan daftar periksa keamanan terisi | R-SEC-61, R-SEC-69, NFR-06 | 🔒️ Lengkapi uji keamanan dan daftar periksa |
| 12.3 | Uji beban ringan, rencana kueri, Lighthouse dan anggaran bundel di CI, hasilnya dicatat | NFR-03, NFR-04, R-PERF-02, aturan 10.1 | ⚡️ Ukur performa terhadap anggaran dan catat hasilnya |
| 12.4 | Pemeriksaan aksesibilitas semua halaman dan uji papan ketik | NFR-11 | ♿️ Perbaiki temuan aksesibilitas di semua halaman |
| 12.5 | Deploy API: citra minimal, compose, Caddy, peran basis data aplikasi, migrasi dan data contoh ke Neon | Spesifikasi 21.8, R-API-13, R-SEC-60, R-SEC-76, R-LOG-09 | 🚀 Tambah berkas deploy API dengan Docker dan Caddy |
| 12.6 | Deploy app: direktori akar `app`, variabel lingkungan, wilayah Singapura | Spesifikasi 21.8 | 🚀 Siapkan deploy app di Vercel |
| 12.7 | README lengkap, daftar pustaka pihak ketiga, panduan verifikasi segel, diagram arsitektur dan basis data | NFR-16, aturan bagian 13 | 📝 Lengkapi README dan dokumen pendukung |

**Gerbang:** lihat bagian 8.

Deploy boleh dimajukan ke akhir Tahap 2 sebagai kerangka berjalan bila prasyarat 7 dan 8 sudah siap, supaya setiap push langsung terlihat di alamat live.

## 7. P1 setelah seluruh P0 lulus

Dikerjakan hanya setelah gerbang akhir lulus, dengan protokol yang sama. Satu baris satu commit.

| # | Fitur | Rujukan |
|---|---|---|
| P1.1 | Lupa kata sandi lewat email (butuh layanan surel, dicatat sebagai pihak ketiga) | FR-AUTH-09 |
| P1.2 | Anggota staf lewat undangan dan izin mencatat pembayaran | FR-ORG-09, FR-ORG-10 |
| P1.3 | Sanggahan atas titipan dan hitungan dengan koreksi dua pihak | FR-DRP-11, FR-CNT-12 |
| P1.4 | Kiriman yang diumumkan penitip lalu dikonfirmasi toko | FR-DRP-12 |
| P1.5 | Pembayaran sebagian dan terbit Setoran lebih awal | FR-STL-11, FR-STL-13 |
| P1.6 | Tampilan peta dan daftar favorit di Cari Mitra | FR-DIR-13, FR-DIR-14 |
| P1.7 | Halaman admin moderasi dan ulasan antar mitra | FR-TRS-05, FR-TRS-06 |
| P1.8 | Pemberitahuan dorong dan pengaturan jenis pemberitahuan | FR-NTF-07, FR-NTF-08, FR-PWA-07 |
| P1.9 | Unduh PDF Kesepakatan, Setoran, dan laporan, dengan QR segel | FR-AGR-07, FR-STL-12, FR-RPT-09, FR-SIG-06 |
| P1.10 | Saran produk lambat, pindah stok, dan toko terbaik | INS-04, INS-05, INS-08 |
| P1.11 | Hitung per batch, layar tukar barang, foto barang, percakapan di Ajuan | FR-CNT-13, FR-CNT-14, FR-DRP-13, FR-PRP-10 |
| P1.12 | Data terstruktur untuk mesin pencari dan hapus akun | FR-PUB-10, FR-AUTH-10 |

## 8. Gerbang akhir

Semua butir harus terpenuhi sebelum karya diserahkan.

**Fungsi**

1. Lima alur utama (A sampai E) lulus uji ujung ke ujung di HP dan desktop.
2. Uji luring Tahap 8 lulus.
3. Semua vektor uji lulus di Go dan TypeScript.
4. Tidak ada uji yang dilewati.

**Mutu**

5. `make check` dan `make test` lulus dari klon baru.
6. Cakupan domain 90% ke atas.
7. `docs/SECURITY_CHECKLIST.md` terisi untuk setiap aturan `R-SEC` dengan tautan ke ujinya.
8. `docs/KINERJA.md` menunjukkan semua anggaran tercapai.
9. Tidak ada temuan terbuka dari pemindai rahasia dan pemindai kerentanan.

**Live**

10. App dan API live. Akun demo bisa dipakai lewat tombol "Coba akun demo".
11. Uji asap terhadap alamat live lulus: masuk demo, buka dasbor, catat satu Hitung Sisa, periksa satu segel di halaman verifikasi.
12. Server API tersambung sebagai peran aplikasi, bukan pemilik skema.

**Serahan lomba** (PRD bagian 17)

13. Repositori publik dengan README yang membuat orang baru bisa menjalankan proyek.
14. `docs/THIRD_PARTY.md` lengkap untuk disalin ke proposal.
15. Tangkapan layar, diagram arsitektur, dan diagram basis data siap untuk proposal.
16. Skenario video di spesifikasi bagian 25 bisa dijalankan pada data contoh.
