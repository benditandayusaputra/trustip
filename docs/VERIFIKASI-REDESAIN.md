# Verifikasi redesain

Tanggal pengujian: 9–10 Oktober 2026.

## Hasil implementasi

Warna utama berubah menjadi biru `#2563EB`. Tema Sistem, Terang, dan Gelap tersedia pada kepala halaman dan Pengaturan. Pilihan tersimpan di perangkat, mengikuti perubahan sistem bila dipilih, serta diselaraskan antar tab sebelum isi halaman ditampilkan. Pemilih tetap bekerja pada halaman statis dan fallback luring.

Seluruh rute UI pada `RENCANA-REDESAIN.md` memakai token dan komponen bersama: navigasi sesuai peran, sidebar desktop, navigasi bawah HP dengan safe area, kontrol, tabel, dialog, pemberitahuan, grafik, peta, dan halaman galat. Beranda, direktori, profil, produk publik, dasbor, formulir transaksi, halaman akun, dokumen kebijakan, dan pengaturan mendapat susunan visual baru. Informasi publik, hak akses, rumus uang, serta alur konfirmasi dua pihak tetap mengikuti spesifikasi.

Komponen tambahan kecil dan memiliki satu tanggung jawab: identitas merek, pemilih tema, bagian pengaturan tema, susunan filter, panel tindakan formulir, serta contoh catatan beranda. Teks baru berada di katalog fitur; warna, ukuran, dan jarak memakai token. Tidak ada pengecualian lint atau tipe maupun dependensi runtime tambahan.

## Gerbang pengujian

| Pemeriksaan | Hasil |
|---|---|
| `make check` | Lulus: pemeriksaan komentar, format, lint, kode tak terpakai, tipe, token, dan dependensi |
| `make test` | Lulus: seluruh paket Go dengan race detector; 82 berkas dan 495 uji frontend |
| `make e2e` | Lulus: 65 kasus dalam lima kelompok, tanpa uji yang dilewati |
| Aksesibilitas | 102 audit halaman dalam dua tema, tanpa pelanggaran serius/kritis; 408 pemeriksaan lebar halaman pada 320, 768, 1024, dan 1440 px |
| Keyboard | Masuk dengan akun toko hingga membuka Hitung Cepat tanpa mouse |
| Tema/PWA | Pilihan tersimpan, lintas tab, mengikuti sistem, storage diblokir, sebelum parse body, cache aset, dan pergantian tema saat luring |
| Fungsi | Cari mitra, ajuan, tawar-menawar, kebutuhan, kesepakatan, titipan, hitung, bayar, konfirmasi, antrean luring, dan perubahan stok |
| Keamanan | CSP/kepala keamanan, muatan skrip di teks bebas, dan penolakan permintaan tulis lintas situs |
| Kompatibilitas | 42 uji tambahan lulus pada Firefox, WebKit desktop, dan WebKit dengan profil iPhone 13 |
| Bundel | Halaman publik memenuhi batas JavaScript awal 90 KB gzip; layar pencatatan memenuhi 120 KB |

Isi staging diperiksa pada checkout terpisah dengan dependensi dari `npm ci`: `make check` dan seluruh 495 uji frontend lulus. Seluruh berkas Go yang dilacak sama dengan workspace yang lulus suite lengkap. Uji fixture demo dijalankan ulang pada checkout tersebut dengan race detector dan lulus. Fixture membangun riwayat 30 hari sekali lalu memastikan reset menyisakan nol baris demo; pembangunan riwayat kedua pada fixture lama membuat suite mencapai batas 30 menit. Pemeriksaan lengkap tetap dijalankan oleh kait Git sebelum push.

Audit mencakup pengunjung, akun toko, dan akun penitip. Jumlah audit adalah kunjungan layar menurut peran dan tema, bukan jumlah rute unik. Halaman dinamis memakai dokumen yang ada di akun demo. Keadaan undangan, onboarding, segel sah, konflik antrean, serta pembayaran dikonfirmasi juga diperiksa melalui skenario ujung ke ujung. Uji otomatis lebar halaman memeriksa overflow dokumen; pemeriksaan visual tambahan menilai susunan isian dan angka.

Pengujian sistem Firefox menggunakan preferensi gelap native pada profil browser uji. Diagnosis lokal menunjukkan emulasi preferensi awal hilang saat navigasi dengan COOP: tanpa perubahan aplikasi, preferensi native tetap menghasilkan tema yang benar. Header COOP/CSP tidak dihapus dari aplikasi maupun suite pengujian. Pemeriksaan responsivitas memakai pilihan terang/gelap eksplisit; perilaku mengikuti perubahan sistem diperiksa tersendiri. Pengaturan preferensi mengikuti [konfigurasi Firefox Playwright](https://github.com/microsoft/playwright/blob/main/browser_patches/firefox/preferences/playwright.cfg).

## Perbaikan yang ditemukan saat pengujian

- Manifest service worker memakai jalur relatif. Jalur tersebut kini dinormalkan terhadap scope sebelum pencocokan cache, sehingga aset aplikasi dapat dibuka luring. Halaman yang belum tersimpan diarahkan ke fallback `/luring`.
- Pemilih tema menggunakan penangan perubahan native dari skrip asal sendiri, sehingga tetap berfungsi pada halaman luring yang menonaktifkan hidrasi. CSP tetap berlaku.
- Ringkasan pencatatan berada di alur formulir dekat tombol simpan, sehingga tidak menutupi isian saat layar kecil atau keyboard muncul.
- Isian jumlah memperbarui pratinjau saat diketik, sebelum fokus berpindah. Uji FR-CNT-04 yang menahan fokus pada isian membuktikan kegagalan lama dan kini lulus; batas minimum/maksimum serta perhitungan yang disimpan tetap diperiksa.
- Kerangka halaman mempertahankan instance isi ketika sesi selesai dimuat. Teks pencarian yang diketik sebelum navigasi anggota muncul tidak hilang. Uji regresi sengaja menahan jawaban akun untuk memeriksa keadaan ini.
- Menu Lainnya menutup setelah tautan dipilih, sehingga dialog tidak menutupi halaman tujuan pada navigasi aplikasi. Uji regresi membuka menu dari dasbor HP lalu berpindah ke Produk.
- Widget verifikasi mengikuti tema halaman dan mengganti token ketika tampilannya berubah. Kontainer sempit memakai ukuran compact; kontainer cukup lebar memakai flexible. Ukuran dan tinggi cadangan mengikuti [konfigurasi resmi Turnstile](https://developers.cloudflare.com/turnstile/get-started/client-side-rendering/widget-configurations/). Uji responsif menunggu widget menghasilkan token agar pemeriksaan lebar mencakup widget yang selesai dimuat.
- Nilai rupiah panjang mendapat kartu selebar baris pada HP sempit. Tabel desktop menggulir di wadahnya dan menjadi kartu berlabel di HP.
- Hitungan toko per radius menerima pengunjung tanpa organisasi melalui UUID nullable; uji repositori yang sebelumnya gagal kini lulus.
- Skenario transaksi dan aksesibilitas dijalankan terpisah dari kelompok umum agar lalu lintas uji tidak saling menghabiskan batas permintaan API. Batas aplikasi tidak dilonggarkan dan tidak ada uji yang dilewati.
- Audit mencoba ulang respons sementara `429`/`5xx` secara terbatas, termasuk kegagalan pada permintaan API setelah dokumen dimuat. Semua halaman, lebar layar, dan aturan aksesibilitas tetap wajib lulus.

## Batas bukti

Pengujian lokal memakai build produksi frontend dan API lokal yang terhubung ke Neon sesuai aturan proyek. Hasil lab tidak membuktikan performa deployment, INP lapangan, atau waktu tugas pengguna. Target pencarian dua menit, titipan tiga produk tiga puluh detik, dan hitung tiga produk empat puluh lima detik tetap membutuhkan uji pengguna.

Suite utama memakai Chromium desktop dan profil Pixel 7. Pemeriksaan tambahan memakai engine Firefox serta WebKit dari Playwright untuk tema, halaman publik responsif, dan header keamanan. Ini bukan pengujian Safari/Edge terpasang maupun iPhone fisik; pemeriksaan manual browser dan perangkat lomba tetap perlu dilakukan.

## Pengukuran performa

Lighthouse memakai konfigurasi proyek yang sama sebelum dan sesudah preload skrip tema, dengan pelambatan HP melalui DevTools. Skrip tema semula berada setelah daftar modulepreload; waterfall menunjukkan unduhannya selesai pada 4,20 detik di Beranda dan 7,63 detik di Cari Toko. Preload di awal dokumen membuatnya selesai sekitar 1,19 detik, sambil mempertahankan penerapan tema sebelum isi diparse dan CSP asal sendiri. Pola preload mengikuti [dokumentasi MDN](https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Attributes/rel/preload).

| Halaman | Skor sebelum preload | LCP sebelum | Skor sesudah | LCP sesudah | CLS sesudah |
|---|---:|---:|---:|---:|---:|
| Beranda | 75 | 4,24 s | 98 | 1,85 s | 0 |
| Cari Toko | 58 | 7,66 s | 98 | 1,87 s | 0,0028 |
| Profil Toko | 69 | 4,95 s | 98 | 1,86 s | 0 |
| Dasbor Toko | 58 | 8,49 s | 61 | 8,47 s | 0,0047 |

Ketiga halaman publik memenuhi skor minimal 90, LCP di bawah 2,5 detik, dan CLS di bawah 0,1. Dasbor memenuhi batas CLS, tetapi skor dan LCP belum memenuhi target pada server pratinjau HTTP/1.1. Batas dasbor sudah tercatat sebelum redesain pada `KINERJA.md`; area aplikasi memakai rendering peramban dan banyak potongan modul. HTTP/2 produksi perlu diukur tersendiri sebelum menyatakan target dasbor tercapai. Pengujian performa tidak melonggarkan batas untuk halaman publik.

Panduan lomba ditindaklanjuti pada desain responsif, orisinalitas, fungsi, dokumentasi, dan aksesibilitas. Proposal PDF, video demo, serta pembuktian website/repository publik merupakan serahan lomba tersendiri. Font Times New Roman dan margin proposal tidak diterapkan ke website. Inspirasi desain dicatat pada rencana; tidak ada aset atau kode referensi yang disalin.

## Pemeriksaan visual

72 tangkapan halaman dan 72 tangkapan area layar berasal dari build akhir yang berjalan. Dua tema dan lebar 320, 390, serta 1440 px diperiksa pada 12 layar: Beranda, Masuk, Daftar, Cari Toko, Produk Publik, Dasbor Toko/Penitip, Hitung Sisa, Detail Setoran, Verifikasi Segel, dan Pengaturan Toko/Penitip. Preferensi tema dibandingkan dengan atribut tema yang diterapkan; tidak ada overflow dokumen, alert aplikasi, atau galat JavaScript dalam kunjungan tersebut.

Pratinjau Hitung Sisa diperiksa setelah seluruh isian diedit sambil fokus tetap berada pada isian terakhir. Pemilih tema di Pengaturan diperiksa keselarasan nilainya dengan pemilih pada kepala halaman. Tangkapan layar memakai locale Indonesia dan zona waktu Asia/Jakarta. Enam gambar implementasi baru ditautkan dari README; gambar sebelumnya tetap disimpan.

- [Dasbor Toko](screenshots/redesign/dasbor-toko.png)
- [Hitung Sisa](screenshots/redesign/hitung-sisa.png)
- [Cari Toko](screenshots/redesign/cari-toko.png)
- [Dasbor Penitip](screenshots/redesign/dasbor-penitip.png)
- [Produk Publik](screenshots/redesign/produk-publik.png)
- [Verifikasi Segel](screenshots/redesign/verifikasi-segel.png)

## Dokumen terkait

- [Desain implementasi](DESIGN.md)
- [Rencana yang disetujui](RENCANA-REDESAIN.md)
- [Keputusan tema](adr/0046-tema-terang-gelap-dan-biru.md)
- [Kinerja sebelum redesain dan batas yang sudah diketahui](KINERJA.md)
- [Lisensi pihak ketiga](THIRD_PARTY.md)
