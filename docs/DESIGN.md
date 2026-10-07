# DESIGN.md: Trustip

Arah desain antarmuka Trustip. Dokumen ini adalah adaptasi dari berkas acuan desain dasbor gelap yang dipilih tim (latar `#121212`, aksen cyan, angka monospace), disesuaikan untuk aplikasi titip jual yang dipakai dari HP di warung dan dari laptop di rumah produksi.

Bila dokumen ini bertentangan dengan `02-SPESIFIKASI-FITUR.md` bagian 24 soal tampilan, dokumen ini yang berlaku. Soal perilaku (keadaan memuat, pesan galat, pola formulir), bagian 24 yang berlaku.

## 1. Tema dan suasana

Trustip harus terasa seperti **buku catatan dagang yang bisa dipercaya dan dikendalikan**: tenang, rapi, angka lebih menonjol daripada hiasan. Bukan media sosial, bukan aplikasi permainan, bukan toko daring.

Tiga kesan yang dikejar:

1. **Andal.** Angka uang dan jumlah barang selalu jelas asalnya dan selalu sama di dua pihak.
2. **Terkendali.** Pengguna selalu tahu apa yang sedang menunggu tindakannya.
3. **Ringkas.** Suasananya dasbor data, tetapi disederhanakan untuk pemilik warung dan produsen rumahan, bukan untuk analis.

Perangkat sasaran: HP Android kelas bawah selebar 360 piksel lebih dulu, lalu tablet dan laptop.

## 2. Palet warna

Tema gelap adalah satu-satunya tema. Tidak ada tema terang.

### 2.1 Warna dasar (dari berkas acuan)

| Token | Nilai | Pemakaian |
|---|---|---|
| `--color-bg` | `#121212` | Latar halaman |
| `--color-surface` | `#1E1E1E` | Kartu, lembar, bilah navigasi |
| `--color-surface-hover` | `#252525` | Baris tabel dan kartu saat disorot atau ditekan |
| `--color-border` | `#2C2C2E` | Garis kartu, pemisah baris, garis kisi grafik |
| `--color-accent` | `#00E5FF` | Data utama di grafik, angka kunci, tombol aksi utama, tautan, cincin fokus |
| `--color-danger` | `#FF453A` | Lewat tempo, kedaluwarsa, ditolak, galat |
| `--color-success` | `#32D74B` | Sesuai, lunas, tersambung, nilai normal |
| `--color-chart-fill` | `#30D158` | Ujung gradasi area grafik |
| `--color-text` | `#FFFFFF` | Teks utama |
| `--color-text-muted` | `#98989D` | Label, keterangan, teks sekunder |
| `--color-danger-surface` | `#3A1C1C` | Latar kotak peringatan |

### 2.2 Warna tambahan (tidak ada di berkas acuan, ditambahkan karena dibutuhkan)

| Token | Nilai | Alasan |
|---|---|---|
| `--color-warning` | `#FFD60A` | Status "menunggu" butuh warna sendiri. Diambil dari keluarga warna yang sama dengan merah dan hijau acuan |
| `--color-border-strong` | `#6E6E73` | Garis tepi isian dan tombol sekunder. Garis `#2C2C2E` hanya berkontras 1,2:1 terhadap kartu, di bawah syarat 3:1 untuk batas komponen interaktif |
| `--color-success-surface` | `#16301C` | Latar lencana dan kotak status berhasil |
| `--color-warning-surface` | `#332B0A` | Latar lencana dan kotak status menunggu |
| `--color-accent-surface` | `#0A2E33` | Latar lencana informasi dan baris terpilih |

### 2.3 Kontras yang sudah dihitung

Semua pasangan di bawah memenuhi WCAG AA (4,5:1 untuk teks, 3:1 untuk batas komponen).

| Depan | Belakang | Rasio | Catatan |
|---|---|---:|---|
| `#FFFFFF` | `#121212` | 18,7 | Teks utama di halaman |
| `#FFFFFF` | `#1E1E1E` | 16,7 | Teks utama di kartu |
| `#98989D` | `#121212` | 6,5 | Teks sekunder di halaman |
| `#98989D` | `#1E1E1E` | 5,8 | Teks sekunder di kartu |
| `#98989D` | `#252525` | 5,3 | Teks sekunder di baris tersorot |
| `#00E5FF` | `#1E1E1E` | 10,8 | Angka kunci dan tautan di kartu |
| `#121212` | `#00E5FF` | 12,2 | Teks di atas tombol utama |
| `#FF453A` | `#1E1E1E` | 4,9 | Teks merah di kartu |
| `#FF453A` | `#3A1C1C` | 4,5 | Pas di batas. Hanya untuk judul dan ikon peringatan |
| `#FFFFFF` | `#3A1C1C` | 15,4 | Isi kotak peringatan |
| `#32D74B` | `#1E1E1E` | 8,7 | Teks hijau di kartu |
| `#FFD60A` | `#1E1E1E` | 11,8 | Teks kuning di kartu |
| `#6E6E73` | `#1E1E1E` | 3,3 | Garis tepi isian di kartu |
| `#6E6E73` | `#121212` | 3,7 | Garis tepi isian di halaman |

Aturan turunan:

- Teks di atas tombol berwarna cerah (cyan, hijau, kuning, merah) selalu `#121212`, tidak pernah putih.
- Isi kotak peringatan memakai teks putih. Merah hanya untuk judul, ikon, dan garis kiri.
- `#2C2C2E` hanya untuk garis hias. Batas yang harus terlihat supaya komponen bisa dipakai memakai `--color-border-strong`.

### 2.4 Warna status

Status selalu berupa warna ditambah teks, tidak pernah warna saja.

| Arti | Warna | Contoh status |
|---|---|---|
| Netral | `--color-text-muted` | Draf, berakhir, ditarik, dibatalkan |
| Informasi | `--color-accent` | Dikirim, terbit, giliran pihak lain |
| Menunggu | `--color-warning` | Menunggu konfirmasi, menunggu dikirim (luring), perlu ditinjau |
| Berhasil | `--color-success` | Sesuai, lunas, diterima, segel sah, tersambung |
| Bahaya | `--color-danger` | Ditolak, lewat tempo, kedaluwarsa, segel tidak sah |

## 3. Gaya komponen

| Komponen | Aturan |
|---|---|
| Kartu | Sudut 16 px. Latar `--color-surface`. Jarak dalam 20 px (16 px di layar di bawah 480 px). Garis `1px solid --color-border`. Tanpa bayangan |
| Kartu angka kunci | Label 14 px warna sekunder di atas, angka monospace tebal di bawah, satu baris keterangan kecil. Angka uang selalu diawali `Rp` |
| Tombol utama | Latar `--color-accent`, teks `#121212`, sudut 12 px, tinggi minimal 48 px. Satu tombol utama per layar |
| Tombol sekunder | Latar transparan, garis `--color-border-strong`, teks putih |
| Tombol bahaya | Latar transparan, garis dan teks `--color-danger`. Berlatar merah hanya di dalam dialog konfirmasi |
| Isian | Latar `--color-bg` di dalam kartu, garis `--color-border-strong`, sudut 12 px, tinggi minimal 48 px, teks 16 px. Saat fokus garis menjadi `--color-accent` |
| Isian angka dan uang | Huruf monospace, rata kanan, papan angka di HP. Tombol kurang dan tambah 48 px di kiri dan kanan |
| Lencana status | Bentuk pil, latar warna permukaan status, teks warna status, 12 px tebal sedang, selalu berisi kata |
| Kotak peringatan | Latar `#3A1C1C`, garis kiri `4px solid #FF453A`, judul merah, isi putih, satu tombol aksi |
| Tabel | Garis bawah baris `1px solid --color-border`. Baris tersorot `--color-surface-hover`. Kolom angka rata kanan dan monospace. Di bawah 768 px tabel berubah menjadi daftar kartu |
| Grafik | Grafik area. Garis data `#00E5FF` setebal 2 px. Isi area bergradasi dari `#00E5FF` 24% ke `#30D158` 0%. Garis kisi `#2C2C2E`. Tanpa garis tepi tebal, tanpa bayangan. Selalu ada tombol "Lihat sebagai tabel" |
| Peta | Ubin OpenStreetMap digelapkan dengan saringan CSS supaya menyatu dengan latar. Penanda `--color-accent`, penanda terpilih putih. Atribusi tetap tampil |
| Lembar bawah dan dialog | Latar `--color-surface`, sudut atas 16 px, tirai `#000000` 60% |
| Toast | Latar `--color-surface`, garis kiri 4 px sesuai warna status, tampil di atas navigasi bawah |
| Kerangka muat | Blok `--color-surface-hover` berbentuk menyerupai isi, dengan kilau pelan |
| Cincin fokus | `2px solid --color-accent` dengan jarak 2 px. Tidak pernah dihilangkan |
| Pita luring | Pita `--color-warning-surface` dengan teks `--color-warning` di bawah bilah atas |
| Tanda segel | Ikon perisai dan teks "Segel sah" berwarna hijau, atau "Segel tidak sah" berwarna merah, di kepala dokumen |

## 4. Tipografi

| Peran | Huruf | Ukuran dan ketebalan |
|---|---|---|
| Judul halaman | Inter | 24 px, 600 |
| Judul bagian dan kartu | Inter | 18 px, 600 |
| Isi | Inter | 16 px di bawah 768 px, 14 px untuk tabel dan panel padat di layar lebar, 400 |
| Label dan keterangan | Inter | 14 px, 500. Keterangan kecil 12 px |
| Angka kunci (rupiah, unit) | JetBrains Mono | 32 px, 700. 28 px di layar di bawah 480 px |
| Angka di tabel, isian, dan rekap | JetBrains Mono | Seukuran teks sekitarnya, 500 |

Aturan:

- Semua angka uang dan jumlah barang memakai huruf monospace dengan angka selebar tetap, supaya kolom lurus dan perubahan nilai tidak menggeser tata letak.
- Berkas acuan menetapkan isi 14 px. Trustip memakai 16 px di HP dan di semua isian karena dua alasan: iOS memperbesar halaman bila isian di bawah 16 px, dan pengguna membaca sambil berdiri di toko.
- Huruf disajikan dari server sendiri dalam format WOFF2, subset Latin, dengan `font-display: swap`. Empat berkas: Inter 400 dan 600, JetBrains Mono 500 dan 700. Hanya Inter 400 yang dimuat lebih awal.
- Tinggi baris 1,5 untuk isi dan 1,2 untuk judul dan angka kunci.

## 5. Tata letak

- **Kisi:** 12 kolom dari 1024 px ke atas, 8 kolom di 768 sampai 1023 px, 4 kolom di bawahnya.
- **Jarak:** kelipatan 8 px (8, 16, 24, 32, 48). Jarak 4 px hanya untuk celah ikon dan teks. Jarak dalam kartu 20 px mengikuti berkas acuan.
- **Lebar isi maksimal:** 1200 px, rata tengah.
- **Navigasi:** bilah bawah lima butir di HP, bilah samping di 1024 px ke atas.
- **Target sentuh:** minimal 44 px, tombol utama 48 px.

Dasbor di layar lebar mengikuti pola berkas acuan:

| Baris | Isi | Lebar |
|---|---|---|
| Atas | Kartu angka kunci sejajar | 3 sampai 5 kartu selebar sama |
| Tengah kiri | Grafik utama penjualan harian | 8 kolom |
| Tengah kanan | Panel "Perlu tindakan" dan kartu saran | 4 kolom |
| Bawah | Peringkat dan peluang | 6 dan 6 kolom |

Di HP urutannya menurun menurut kepentingan: angka kunci (dua kolom), Hitung cepat (khusus toko), Perlu tindakan, grafik, peringkat, peluang.

Layar pencatatan (Titipan Masuk dan Hitung Sisa) memakai satu kolom penuh. Rekap tiga angka (omzet, untung toko, harus setor) menempel di bawah layar di atas tombol simpan.

## 6. Gerak

- Durasi 150 sampai 200 ms, kurva keluar melambat. Hanya `opacity` dan `transform` yang dianimasikan.
- Angka rekap yang berubah tidak dianimasikan menghitung naik. Nilainya langsung berganti.
- `prefers-reduced-motion` dihormati: semua gerak dimatikan kecuali pudar singkat.

## 7. Ikon dan gambar

- Satu keluarga ikon garis berlisensi terbuka, tebal garis 1,75 px, ukuran 20 dan 24 px. Ikon selalu ditemani teks, kecuali di bilah navigasi yang sudah berlabel.
- Ikon mengikuti benda di dunia titip jual: rak atau etalase untuk toko, kotak untuk titipan, papan hitung untuk hitung sisa, jabat tangan untuk kesepakatan, dompet untuk setoran, penanda peta untuk lokasi, perisai untuk segel.
- Foto produk dan toko tampil dengan sudut 12 px dan rasio tetap (1:1 untuk kartu, 4:3 untuk halaman), dengan ukuran eksplisit supaya tata letak tidak bergeser.

## 8. Lakukan dan jangan

**Lakukan**

- Pakai grafik area untuk penjualan harian supaya lonjakan akhir pekan terlihat halus.
- Selalu tampilkan nilai rupiah di samping jumlah unit, misalnya "18 terjual, Rp36.000".
- Tampilkan tiga angka rekap dalam urutan tetap: omzet, untung toko, harus setor.
- Beri setiap status sebuah kata, bukan hanya warna.
- Tandai catatan yang belum terkirim dengan jelas dan sebutkan jumlahnya.
- Pakai warna cyan hanya untuk data utama dan satu aksi utama per layar, supaya tetap berarti.

**Jangan**

- Jangan pakai warna pastel atau latar putih di antarmuka.
- Jangan pakai ikon umum yang tidak berhubungan dengan dagang, misalnya bola lampu atau roket.
- Jangan pakai bayangan tebal atau gradasi hias.
- Jangan pakai merah untuk hal selain masalah yang perlu ditangani.
- Jangan menaruh teks putih di atas tombol cyan, hijau, atau kuning.
- Jangan mengandalkan garis `#2C2C2E` sebagai satu-satunya batas isian.
- Jangan pakai tanda pisah panjang di teks antarmuka.

## 9. Pengecualian dan batasan yang disadari

1. **Cetak dan PDF** memakai hitam di atas putih supaya hemat tinta. Larangan latar putih berlaku untuk layar.
2. **Keterbacaan di luar ruang.** Layar gelap lebih sulit dibaca di bawah sinar matahari langsung. Penangkalnya: teks utama putih berkontras 16:1 ke atas, angka besar, dan tidak ada teks penting berwarna abu. Uji pengguna di warung wajib mencatat apakah layar terbaca di meja kasir yang terang.
3. **Peta gelap** dibuat dengan saringan CSS di atas ubin OpenStreetMap, bukan ubin gelap dari penyedia lain, supaya tidak menambah layanan pihak ketiga.

## 10. Token untuk implementasi

Semua nilai di atas ditulis sekali sebagai variabel CSS di `app/src/lib/styles/tokens.css` lalu dipetakan ke tema Tailwind. Komponen tidak boleh menulis nilai heksadesimal, ukuran huruf, atau jarak secara langsung. Pemeriksa gaya menolak nilai warna mentah di luar berkas token.

| Kelompok | Token |
|---|---|
| Warna | Tabel 2.1 dan 2.2 |
| Sudut | `--radius-card: 16px`, `--radius-control: 12px`, `--radius-pill: 999px` |
| Jarak | `--space-1: 4px`, `--space-2: 8px`, `--space-4: 16px`, `--space-5: 20px`, `--space-6: 24px`, `--space-8: 32px`, `--space-12: 48px` |
| Huruf | `--font-sans: Inter`, `--font-mono: JetBrains Mono` |
| Ukuran huruf | `--text-xs: 12px`, `--text-sm: 14px`, `--text-base: 16px`, `--text-lg: 18px`, `--text-xl: 24px`, `--text-kpi: 32px` |
| Gerak | `--duration-fast: 150ms`, `--duration-base: 200ms` |
| Lapisan | `--z-nav: 10`, `--z-sheet: 20`, `--z-toast: 30` |
