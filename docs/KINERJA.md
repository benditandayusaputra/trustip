# Kinerja

Hasil ukur terhadap anggaran di `03-ATURAN-REKAYASA.md` bagian 10.1. Diukur 9 Oktober 2026. Halaman diukur dari mesin pengembang di Indonesia dengan API lokal. Uji beban dijalankan dari VPS API (Oracle, Melbourne) ke basis data aplikasi di Neon Sydney (`aws-ap-southeast-2`), sama dengan jalur produksi.

## Ringkasan

| Ukuran | Batas | Hasil | Status | Diperiksa oleh |
|---|---|---|---|---|
| LCP halaman publik, 4G lambat | < 2,5 detik | 1,85 sampai 1,87 detik | Tercapai | `app/tests/lighthouse.test.ts` |
| LCP dasbor, 4G lambat | < 2,5 detik | 7,1 detik di server pratinjau lokal | Belum terbukti | Lihat bagian Dasbor |
| CLS | < 0,1 | 0 di halaman publik, 0,075 di dasbor | Tercapai | `app/tests/lighthouse.test.ts` |
| Skor performa Lighthouse HP | ≥ 90 | 98 di halaman publik | Tercapai | `app/tests/lighthouse.test.ts` |
| INP | < 200 ms | TBT 0 ms di semua halaman yang diukur | Tercapai di lab | Lighthouse (TBT sebagai pengganti INP di lab) |
| JavaScript awal halaman publik | ≤ 90 KB terkompresi | 47 sampai 88 KB | Tercapai | `app/tests/performance-budget.test.ts`, `app/tests/partnership.test.ts` |
| JavaScript awal layar pencatatan | ≤ 120 KB terkompresi | di bawah 120 KB | Tercapai | `app/tests/partnership.test.ts` (`recordingScreensStayWithinBudget`) |
| API baca, p95 | < 150 ms | 17 sampai 100 ms | Tercapai | `make load` dari VPS (`TestLightLoad_NFR04_...`) |
| API tulis, p95 | < 300 ms | 787 ms (simpan Hitung Sisa) | Belum tercapai | Lihat bagian Uji beban |
| Cari terdekat pada 10.000 organisasi | < 200 ms | 100 ms | Tercapai | `make load` dari VPS (`TestNearestStores_NFR03_...`) |
| Jumlah kueri per permintaan daftar | ≤ 3 | Daftar Titipan Masuk 1, Cari toko 3 | Tercapai | Penelusur kueri, lihat Rencana kueri |

## Cara mengukur

- **Lighthouse** dijalankan dari Playwright (`make e2e`, lintasan ketiga) dengan profil HP bawaan Lighthouse dan pelambatan nyata (`throttlingMethod: 'devtools'`, RTT 150 ms, 1,6 Mbps, CPU 4 kali). Mode simulasi tidak dipakai karena di server lokal semua skrip selesai sebelum cat pertama, sehingga simulasi membebankan seluruh JavaScript ke FCP, padahal halaman publik dirender server dan sudah tampil sebelum JavaScript datang.
- **JavaScript awal** adalah jumlah ukuran gzip semua skrip yang dimuat sampai peristiwa `load`. Potongan yang dimuat saat dibutuhkan (peta, grafik, captcha, QR) tidak dihitung (R-PERF-11).
- **Uji beban ringan** (`make load`, tag build `load`) mengirim 30 permintaan baca per rute dan 15 penyimpanan Hitung Sisa berturut-turut lewat server lengkap (tanda tangan tepi, sesi, layanan, basis data), lalu menghitung p95. Uji ini dipisah dari `make test` karena hasilnya bergantung pada jarak ke basis data, dan harus dijalankan dari mesin yang dekat dengan Neon.

## JavaScript awal per halaman

| Halaman | Terkompresi |
|---|---:|
| `/` | 71,5 KB |
| `/cari/toko` | 77,6 KB |
| `/cari/penitip` | 88,4 KB |
| `/kebutuhan` | 82,4 KB |
| `/masuk` | 47,6 KB |
| `/daftar` | 74,1 KB |

Perubahan di Tahap 12: katalog teks antarmuka dipecah per fitur. Sebelumnya satu berkas `id.ts` (18 KB terkompresi) ikut dimuat di setiap halaman. Sekarang setiap komponen hanya mengimpor katalog yang dipakainya. LCP beranda turun dari 2,26 ke 2,11 detik pada pengukuran simulasi yang sama.

## Lighthouse

| Halaman | Skor | LCP | CLS | Respons server |
|---|---:|---:|---:|---:|
| `/` | 98 | 1,85 s | 0 | 7 ms |
| `/cari/toko` | 98 | 1,85 s | 0 | 29 ms |
| `/t/<toko>` | 98 | 1,87 s | 0 | 111 ms |
| `/app/toko` | 59 | 7,1 s | 0,075 | 2 ms |

## Dasbor

Area `/app` dirender di peramban (`ssr = false`) supaya bisa dipakai luring. Dasbor menunggu JavaScript, lalu `/api/v1/me`, lalu data dasbor.

- **CLS** turun dari 0,19 ke 0,075 setelah dua perbaikan di Tahap 12: tombol keluar dipindah ke menu "Akun" sehingga kepala aplikasi di HP tetap satu baris, dan huruf Inter dimuat lebih awal sehingga teks tidak bergeser saat huruf berganti.
- **LCP** 7,1 detik diukur di server pratinjau lokal yang memakai HTTP/1.1. Dengan enam koneksi per host dan RTT 560 ms dari pelambatan, sekitar 50 potongan skrip area `/app` dimuat dalam delapan gelombang berurutan. Vercel menyajikan HTTP/2, sehingga semua potongan diminta sekaligus. Angka produksi diukur setelah app live dan ditulis di sini.
- Kerja latar luring (paket kerja dan pengiriman antrean) kini menunggu peramban senggang, supaya tidak berebut koneksi dengan permintaan dasbor.

## Uji beban

Dijalankan dari VPS API dengan biner uji yang sama (`go test -c -tags load`), sehingga jarak jaringan sama dengan produksi.

| Permintaan | Basis data di Singapura | Basis data di Sydney |
|---|---:|---:|
| `GET /api/v1/drops` | 165 ms | 17 ms |
| `GET /api/v1/directory/stores` | 513 ms | 55 ms |
| `GET /api/v1/dashboard` | 330 ms | 100 ms |
| Cari terdekat, 10.000 organisasi | 349 ms | 100 ms |
| `POST /api/v1/counts` | 6,9 s | 787 ms |

VPS berada di Melbourne, sehingga basis data aplikasi dipindah dari Neon Singapura ke Neon Sydney (ADR 0045). Satu kueri dari VPS ke Sydney memakan sekitar 12 ms pulang-pergi.

Penyimpanan Hitung Sisa belum memenuhi anggaran tulis 300 ms. Satu penyimpanan menjalankan sekitar 30 kueri berurutan dalam satu transaksi (kunci kemitraan, stok, dokumen, baris, alokasi, mutasi, segel, Setoran, pemberitahuan, audit), lalu membaca ulang dokumennya dengan 6 kueri lagi. Rencana perbaikan: mengirim sisipan yang tidak saling bergantung sebagai satu batch pgx dan membangun jawaban dari data yang baru ditulis tanpa membaca ulang.

## Rencana kueri

Diukur dengan `EXPLAIN (ANALYZE)` di basis data aplikasi berisi data contoh (228 organisasi). Pada ukuran ini perencana memilih pemindaian berurutan untuk tabel kecil. Indeks yang mendukung setiap daftar sudah ada untuk saat data bertambah.

| Kueri | Waktu eksekusi | Indeks pendukung |
|---|---:|---|
| `ListDrops` (toko demo) | 0,53 ms | `drops_partnership_occurred_idx`, `drop_items_line_key` |
| `ListCounts` | 0,34 ms | `counts_partnership_counted_idx` |
| `ListSettlements` | 0,54 ms | `settlements_store_issued_idx`, `settlements_consignor_issued_idx` |
| `ListNotifications` | 3,07 ms | `notifications_user_id_idx` (dipakai) |
| `ListPartnerships` | < 1 ms | `partnerships_store_org_id_idx`, `stock_movements_drop_item_idx` (dipakai) |
| `GetDashboardFigures` | 1,65 ms | `settlements_partnership_status_due_idx` |
| `ListDirectoryNeeds` (radius 10 km) | < 1 ms | `needs_status_category_idx` (dipakai), `organizations_pkey` |
| Cari toko terdekat, 10.000 organisasi | 114 ms | `organizations_location_public_idx` (GiST, dipakai) |

Jumlah kueri per permintaan diukur dengan menurunkan ambang penelusur kueri lambat (`internal/platform/database/tracer.go`) ke nol: daftar Titipan Masuk menjalankan 1 kueri, Cari toko menjalankan 3 (pemirsa, hitungan per radius, halaman daftar).
