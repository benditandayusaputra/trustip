# DESIGN.md: Trustip

Desain antarmuka Trustip yang disetujui tim pada 9 Oktober 2026: primary biru, tema terang dan gelap, serta tata letak yang mengutamakan tindakan dan catatan dua pihak. Keputusan dan alasan dicatat pada `adr/0046-tema-terang-gelap-dan-biru.md`.

Dokumen ini mengatur visual. Perilaku, angka, izin, dan keadaan fitur tetap mengikuti `02-SPESIFIKASI-FITUR.md`. Implementasi menggunakan komponen bersama dan token, bukan nilai mentah di halaman.

## 1. Suasana dan pengguna

Trustip adalah buku catatan dagang yang bisa dipercaya: rapi, tenang, dan mudah dipakai pemilik warung serta produsen rumahan. Halaman publik memperlihatkan manfaat titip jual dan identitas usaha. Halaman operasional menonjolkan langkah berikutnya, status, dan angka yang sumbernya jelas.

HP selebar 360 piksel menjadi sasaran utama, dengan dukungan 320 piksel, tablet, dan desktop. Pengguna tetap bisa menyelesaikan pencatatan dengan satu ibu jari.

## 2. Tema dan palet

Pilihan tema: Sistem, Terang, Gelap. Bawaan mengikuti sistem. Preferensi disimpan di perangkat dan diterapkan sebelum tampilan pertama tanpa melonggarkan CSP. Perubahan sistem hanya mengubah tema bila pilihan Sistem aktif. Penggantian tema tidak mengubah akun, data, ataupun hak akses.

| Token | Terang | Gelap | Penggunaan |
|---|---|---|---|
| `--color-bg` | `#F6F8FC` | `#0B1220` | Latar halaman |
| `--color-surface` | `#FFFFFF` | `#131D2E` | Kartu, header, dialog |
| `--color-surface-hover` | `#EEF2F8` | `#1B293E` | Hover, permukaan sekunder |
| `--color-border` | `#DCE3EF` | `#2D3C53` | Pemisah dekoratif |
| `--color-border-strong` | `#7B879B` | `#71819A` | Batas kontrol |
| `--color-primary` | `#2563EB` | `#2563EB` | Tombol utama dan ikon merek |
| `--color-primary-hover` | `#1D4ED8` | `#1D4ED8` | Hover tombol utama |
| `--color-on-accent` | `#FFFFFF` | `#FFFFFF` | Teks pada primary |
| `--color-accent` | `#1D4ED8` | `#93C5FD` | Tautan, data utama, fokus |
| `--color-accent-surface` | `#EFF6FF` | `#172D50` | Informasi, rekap, navigasi aktif |
| `--color-text` | `#152238` | `#F3F6FC` | Isi dan angka |
| `--color-text-muted` | `#52627A` | `#A5B4CA` | Label dan keterangan |
| `--color-success` | `#166534` | `#86EFAC` | Berhasil |
| `--color-success-surface` | `#F0FDF4` | `#143324` | Permukaan berhasil |
| `--color-warning` | `#92400E` | `#FCD34D` | Menunggu |
| `--color-warning-surface` | `#FFFBEB` | `#352C17` | Permukaan menunggu |
| `--color-danger` | `#B91C1C` | `#FCA5A5` | Galat dan masalah |
| `--color-danger-surface` | `#FEF2F2` | `#3B212B` | Permukaan masalah |
| `--color-chart-fill` | `#2563EB` | `#2563EB` | Isi area grafik |

Teks putih hanya dipakai di atas primary biru atau token yang memang menyediakan kontras cukup. Warna status selalu disertai kata. Tidak ada cyan/hijau sebagai primary atau gradasi hias. Angka biasa berwarna teks utama; biru membantu tindakan dan penekanan yang perlu.

Kontras teks ditargetkan minimal 4,5:1, batas kontrol/fokus minimal 3:1. Garis dekoratif tidak menggantikan batas isian. Uji aksesibilitas kedua tema memeriksa pasangan aktual di seluruh keadaan komponen.

## 3. Tipografi dan jarak

Inter untuk isi; JetBrains Mono untuk rupiah dan jumlah barang. Keduanya disajikan dari server sendiri sebagai WOFF2 dengan `font-display: swap`, hanya Inter reguler dimuat lebih awal.

| Peran | Ukuran |
|---|---|
| Hero publik | 34 px di HP, 48 px mulai tablet |
| Judul halaman | 24 px, 600 |
| Judul bagian | 18 px, 600 |
| Isi dan isian | 16 px di HP |
| Label dan tabel desktop | 14 px |
| Keterangan | 12 px |
| Angka kunci | 24 px, 28 px pada layar lebar; angka utama 32 px |

Angka panjang mendapat kartu selebar baris pada HP di bawah 480 px, sehingga jumlah rupiah tidak terpecah. Angka memakai tabular figures; rupiah selalu memakai `Rp` dan pemisah ribuan. Format berasal dari modul `Intl` yang sama dengan lokal `id-ID` dan zona `Asia/Jakarta`. Jarak mengikuti token kelipatan 4/8 px, dengan kartu 16 sampai 20 px. Radius kartu 16 px dan kontrol 12 px.

## 4. Kerangka seluruh halaman

- Publik: header merek, navigasi, pilihan tema, masuk; footer kebijakan. Halaman publik tetap dirender server.
- Anggota: sidebar desktop dan header tindakan; navigasi bawah pada HP/tablet sesuai peran. Semua rute aplikasi, direktori, dan papan kebutuhan menggunakan kerangka yang sama ketika anggota masuk.
- Penitip: Beranda, Cari, Mitra, Setoran, Lainnya.
- Toko: Beranda, Hitung, Titipan, Cari, Lainnya.
- Lainnya membuka lembar menu yang memuat semua modul relevan. Navigasi aktif memakai permukaan biru lembut dan label; label ikon tidak dihilangkan.
- Lebar isi maksimal 1200 px, formulir 640 px, dokumen baca 720 px. Main dan kolom flex/grid boleh menyusut agar isi tidak memaksa overflow.
- Safe area perangkat dihitung pada navigasi bawah dan toast. Ringkasan formulir berada dalam alur halaman tepat sebelum tombol simpan dan tidak menutupi isian. Dialog serta lembar bawah berada di atas navigasi.

## 5. Pola halaman

### Publik

Beranda menampilkan nilai utama, ajakan daftar/demo, contoh hitungan yang diberi label contoh, masalah titip jual, empat langkah, dan ajakan penutup. Tidak memuat klaim pengguna, testimoni, atau hasil bisnis yang belum terbukti.

Direktori Toko selalu menampilkan pencarian dan urutan. Filter lokasi, kategori, dan syarat dibuka melalui tombol Filter, dengan penjelasan dan jumlah pilihan aktif. Panel filter berada di atas hasil pada HP maupun desktop dan menyesuaikan jumlah kolom dengan ruang yang tersedia. Pilihan Daftar/Peta menggunakan hasil pencarian yang sama; marker lokasi publik bisa diklik atau diaktifkan lewat keyboard untuk membuka detail. Daftar nama toko tetap tersedia sebagai alternatif peta. Direktori Penitip memakai panel filter di samping hasil mulai 1440 px, dan sebelum hasil pada layar lebih kecil. Foto produk ditampilkan jika tersedia; bila belum ada, gunakan placeholder ikon yang tenang, bukan foto rekaan. Profil usaha memakai kepala profil, kategori, keterangan, syarat, dan aksi relevan. Produk memiliki foto utama, harga jual saran, komposisi, pembuat, dan daftar tempat membeli. Kebijakan memakai kolom baca nyaman.

Daftar putih publik tidak berubah: harga setor, angka transaksi, lokasi persis penitip, nomor WhatsApp, dan jumlah stok per toko tidak ditambahkan. Verifikasi segel hanya menampilkan metadata yang diperbolehkan; segel tidak disebut tanda tangan tersertifikasi.

### Operasional

Dasbor toko: lima angka spesifikasi, Hitung cepat dan Perlu tindakan, lalu Saran selebar area isi, diikuti Produk terlaris dan Peluang. Saran menampilkan tiga kartu awal; daftar lain menampilkan lima baris awal. Tombol buka/ringkas tetap menyediakan seluruh data. Dasbor penitip: empat angka spesifikasi, Perlu tindakan berdampingan dengan grafik mulai desktop, lalu Saran selebar isi, dua peringkat, dan Peluang selebar isi. Filter toko/produk berada di dalam kartu grafik dan mempertahankan fokus saat pilihan berubah. Pada HP, Perlu tindakan mendahului grafik. Saran menampilkan tiga kartu awal; peringkat dan Peluang menampilkan lima baris awal, dengan tombol buka/ringkas untuk sisanya. Kolom menyesuaikan ruang setelah sidebar; angka tidak diperkecil sampai sulit dibaca. Informasi gabungan tetap hanya untuk OWNER.

Riwayat, katalog, Ajuan, Kemitraan, Kebutuhan dan Setoran memakai filter di atas isi, status berlabel dan tindakan sesuai hak akses. Detail dokumen membedakan nomor/status, isi, jejak konfirmasi, segel dan langkah berikutnya. Tabel desktop berubah menjadi daftar berlabel di HP; tabel padat hanya boleh memiliki overflow lokal.

Titipan Masuk dan Hitung Sisa memakai baris produk dan kontrol jumlah besar. Ringkasan beserta simpan mudah dicapai, di atas navigasi. Rekap uang selalu berurutan omzet, untung toko, harus setor. Formulir lain dikelompokkan dalam permukaan yang jelas. Preferensi tema tersedia di header dan pengaturan, termasuk untuk staf yang tidak boleh mengubah profil.

## 6. Komponen dan keadaan

| Komponen | Aturan |
|---|---|
| Kartu | Permukaan sesuai tema, border dekoratif, radius 16 px, tanpa bayangan tebal |
| Angka kunci | Label sekunder, angka utama monospace, keterangan opsional |
| Tombol utama | Primary biru, teks putih, tinggi minimal 48 px; satu aksi utama per konteks |
| Sekunder | Permukaan sesuai tema, border kuat, teks utama |
| Bahaya | Border/teks bahaya, akibat dijelaskan di konfirmasi |
| Isian | Tinggi minimal 48 px, teks 16 px, border kuat dan fokus terlihat |
| Jumlah | Tombol tambah/kurang 48 px, papan angka, label yang jelas |
| Badge | Teks status, warna semantik dan latar yang sesuai |
| Grafik | Garis biru, area biru transparan, grid halus, alternatif tabel |
| Peta | Terang pada light, filter gelap hanya pada dark; kontrol dan atribusi selalu terbaca |
| Dialog, lembar bawah, toast | Token tema; fokus dan tindakan sesuai keadaan |
| Skeleton | Ukuran menyerupai isi, tidak menggeser halaman saat data datang |
| Kosong | Penjelasan singkat dan satu tindakan yang relevan |
| Luring | Pita status, jumlah antrean, peringatan usia lama, batas aksi jelas |

Memuat, kosong, berhasil, validasi gagal, jaringan gagal, dan luring wajib ditangani. Galat isian tampil dekat isian, fokus berpindah ke yang pertama salah, dan nilai tidak hilang. Menyimpan menonaktifkan kiriman ganda. Konflik antrean memperlihatkan catatan perangkat dan keadaan server dengan label jelas. Seluruh aturan bisnis dan tenggat konfirmasi tetap berlaku.

## 7. Aksesibilitas, gerak, dan performa

Target sentuh minimal 44 px; kontrol utama 48 px. Fokus selalu terlihat; heading berurutan; landmark utama tunggal per kerangka; label dan nama tindakan jelas. Uang penting tidak disampaikan lewat warna saja. Tabel memiliki caption, gambar memiliki alternatif, dan perubahan penting memakai status/alert yang tepat.

Gerak tidak mengganggu pekerjaan. Transisi kontrol singkat 150 sampai 200 ms; reduced motion mematikan gerak nonesensial. Angka uang langsung berubah tanpa animasi menghitung naik.

Peta, grafik, QR dan captcha tetap dimuat sesuai kebutuhan. Preferensi tema berupa skrip kecil dari asal sendiri, tanpa dependensi baru. JS awal publik maksimal 90 KB terkompresi dan pencatatan 120 KB. Sasaran Lighthouse HP performa/aksesibilitas minimal 90, LCP di bawah 2,5 detik, CLS di bawah 0,1. Ketercapaian dicatat setelah pengukuran.

## 8. Token dan pemeriksaan

Warna, ukuran, jarak, radius, lapisan dan safe area didefinisikan pada `app/src/lib/styles/tokens.css`. Tailwind memetakan token tersebut; komponen tidak menulis nilai mentah. Teks pengguna hanya dari katalog `i18n`.

Pemeriksaan wajib: format, lint, tipe, batas lapisan, komentar, token; uji unit/komponen, alur kedua peran, luring, preferensi tersimpan, tema sistem, aksesibilitas kedua tema dan tata letak HP/desktop. Ikon instalasi PWA memakai biru yang sama dengan merek layar.

Cetak dan dokumen proposal tetap memakai hitam di atas putih. Aturan font dan margin proposal lomba tidak diterapkan ke font website.
