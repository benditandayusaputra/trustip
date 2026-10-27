# DESIGN.md: Trustip

Desain antarmuka Trustip yang disetujui tim pada 9 Oktober 2026: primary biru, tema terang dan gelap, serta tata letak yang mengutamakan tindakan dan catatan dua pihak. Keputusan dan alasan dicatat pada `adr/0046-tema-terang-gelap-dan-biru.md`, penyegaran 16 Oktober 2026 pada `adr/0053-penyegaran-antarmuka-ruang-kerja.md`, dan revisi dari desain Figma 27 Oktober 2026 pada `adr/0060-revisi-tampilan-dari-figma.md`.

Dokumen ini mengatur visual. Perilaku, angka, izin, dan keadaan fitur tetap mengikuti `02-SPESIFIKASI-FITUR.md`. Implementasi menggunakan komponen bersama dan token, bukan nilai mentah di halaman. Hasil QA per halaman dan perbaikan UX dicatat di `QA-ANTARMUKA.md` dan `adr/0054-qa-navigasi-dan-formulir.md`.

## 1. Suasana dan pengguna

Trustip adalah buku catatan dagang yang bisa dipercaya: rapi, tenang, dan mudah dipakai pemilik warung serta produsen rumahan. Halaman publik memperlihatkan manfaat titip jual dan identitas usaha. Halaman operasional menonjolkan langkah berikutnya, status, dan angka yang sumbernya jelas.

HP selebar 360 piksel menjadi sasaran utama, dengan dukungan 320 piksel, tablet, dan desktop. Pengguna tetap bisa menyelesaikan pencatatan dengan satu ibu jari.

## 2. Tema dan palet

Pilihan tema hanya Terang dan Gelap, lewat tombol Mode di header yang membuka menu berlatar padat (bukan pilihan bawaan peramban), atau dua tombol di menu HP. Pengaturan menyediakan sakelar Tampilan terang/gelap. Bawaan Terang, termasuk di perangkat yang memakai mode gelap; preferensi lama Sistem dibaca sebagai Terang. Preferensi disimpan di perangkat dan diterapkan sebelum tampilan pertama tanpa melonggarkan CSP. Penggantian tema tidak mengubah akun, data, ataupun hak akses.

| Token | Terang | Gelap | Penggunaan |
|---|---|---|---|
| `--color-bg` | `#F0F9FF` | `#0D1522` | Latar halaman |
| `--color-surface` | `#FFFFFF` | `#162131` | Kartu, header, dialog |
| `--color-surface-hover` | `#EEF2F8` | `#1B293E` | Hover, permukaan sekunder |
| `--color-border` | `#E3E8F0` | `#2B394C` | Pemisah dekoratif |
| `--color-border-strong` | `#8491A5` | `#71819A` | Batas kontrol |
| `--color-primary` | `#2563EB` | `#2563EB` | Tombol utama dan ikon merek |
| `--color-primary-hover` | `#1D4ED8` | `#1D4ED8` | Hover tombol utama |
| `--color-on-accent` | `#FFFFFF` | `#FFFFFF` | Teks pada primary |
| `--color-accent` | `#1D4ED8` | `#93C5FD` | Tautan, data utama, fokus |
| `--color-accent-surface` | `#EFF6FF` | `#172D50` | Informasi, rekap, navigasi aktif |
| `--color-text` | `#17243B` | `#F3F6FC` | Isi dan angka |
| `--color-text-muted` | `#52627A` | `#A5B4CA` | Label dan keterangan |
| `--color-success` | `#166534` | `#86EFAC` | Berhasil |
| `--color-success-surface` | `#F0FDF4` | `#143324` | Permukaan berhasil |
| `--color-warning` | `#92400E` | `#FCD34D` | Menunggu |
| `--color-warning-surface` | `#FFFBEB` | `#352C17` | Permukaan menunggu |
| `--color-danger` | `#B91C1C` | `#FCA5A5` | Galat dan masalah |
| `--color-danger-surface` | `#FEF2F2` | `#3B212B` | Permukaan masalah |
| `--color-chart-fill` | `#2563EB` | `#2563EB` | Isi area grafik |

Teks putih hanya dipakai di atas primary biru atau token yang memang menyediakan kontras cukup. Warna status selalu disertai kata. Tidak ada cyan/hijau sebagai primary. Gradasi hanya dipakai lembut, dari permukaan biru muda ke permukaan, pada panel Beranda. Angka biasa berwarna teks utama; biru membantu tindakan dan penekanan yang perlu.

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

Angka panjang mendapat kartu selebar baris pada HP di bawah 480 px, sehingga jumlah rupiah tidak terpecah. Angka memakai tabular figures; rupiah selalu memakai `Rp` dan pemisah ribuan. Format berasal dari modul `Intl` yang sama dengan lokal `id-ID` dan zona `Asia/Jakarta`. Jarak mengikuti token kelipatan 4/8 px, dengan kartu 16 sampai 20 px. Radius kartu 20 px dan kontrol 12 px. Bayangan kartu tipis memakai token bersama; bayangan lebih tinggi hanya untuk menu mengambang dan hover kartu produk.

## 4. Kerangka seluruh halaman

- Publik: header merek, navigasi pil (Beranda, Cari toko, Kebutuhan toko, Panduan) dengan halaman aktif disorot biru, pilihan tema, masuk; footer merek, direktori, cara kerja, demo, pendaftaran, dan kebijakan. Pada HP, menu membuka tautan Cari toko, Cari produk dan penitip, serta Kebutuhan toko; menu ditutup lewat Escape, klik luar, atau setelah memilih tautan. Profil toko, profil penitip, produk, dan Kebutuhan publik punya tombol Kembali di atas isi. Halaman publik tetap dirender server.
- Anggota: sidebar desktop biru primary selebar 256 px berisi merek dengan nama ruang kerja di bawahnya, lalu kelompok Ruang kerja dan Kelola usaha; menu aktif putih dengan teks biru. Pemberitahuan hanya lewat lonceng di header, dan Setoran berada di Laporan. Header ringkas setinggi 64 px: ikon modul, nama ruang kerja, dan judul halaman. Halaman detail atau halaman di luar menu mengganti ikon dengan tombol kembali dan menampilkan jejak `Induk > Judul`; judul tidak diulang di isi halaman. Kanan header berisi tombol Mode, lonceng dengan jumlah belum dibaca, dan tombol Profil. Menu Profil hanya berisi Kelola profil dan Keluar, terbuka saat kursor atau fokus berada di tombol Profil; Keluar dari semua perangkat berada di Pengaturan. Profil dan Pemberitahuan adalah halaman fokus: tanpa sidebar, header hanya berisi tombol kembali dan judul; navigasi bawah pada HP/tablet sesuai peran. Semua rute aplikasi, direktori, dan papan kebutuhan menggunakan kerangka yang sama ketika anggota masuk.
- Penitip: Ruang kerja (Beranda, Cari, Mitra, Laporan), Kelola usaha (Produk, Ajuan, Stok, Pengaturan). Navigasi bawah: Beranda, Cari, Mitra, Laporan, Lainnya.
- Toko: Ruang kerja (Beranda, Hitung, Titipan, Cari, Laporan), Kelola usaha (Mitra, Kebutuhan, Ajuan, Stok, Antrean, Pengaturan).
- Navigasi bawah toko: Beranda, Hitung, Titipan, Cari, Lainnya.
- Lainnya membuka lembar menu yang memuat semua modul relevan. Navigasi aktif di sidebar memakai permukaan putih dan teks biru; navigasi bawah memakai permukaan biru lembut dan label; label ikon tidak dihilangkan.
- Lebar isi maksimal 1200 px, formulir 640 px, dokumen baca 720 px. Main dan kolom flex/grid boleh menyusut agar isi tidak memaksa overflow.
- Safe area perangkat dihitung pada navigasi bawah dan toast. Ringkasan formulir berada dalam alur halaman tepat sebelum tombol simpan dan tidak menutupi isian. Dialog serta lembar bawah berada di atas navigasi.

## 5. Pola halaman

### Publik

Beranda menjelaskan titip jual sejak layar pertama dengan ilustrasi penitip menyerahkan produk ke toko di panel Satu alur, dua mitra, dua kartu peran di bawahnya, serta ajakan daftar/demo. Berikutnya tampil kebutuhan dua peran usaha, lima langkah bernomor (cari mitra, sepakati harga, catat titipan, hitung sisa, konfirmasi setoran) dalam panel Cara kerja, contoh hitungan Begini hasilnya, masalah beserta manfaat pencatatan dengan judul di tengah, dan ajakan penutup. Kartu peran, panel Cara kerja, contoh hitungan, dan kartu masalah memakai panel bergradasi lembut. Paragraf rata kiri. Pembayaran antar mitra dijelaskan sebagai pembayaran langsung yang dicatat dan dikonfirmasi di Trustip. Tidak memuat klaim pengguna, testimoni, atau hasil bisnis yang belum terbukti.

Ilustrasi titip jual dipakai di hero Beranda dan panel akun, dan mengikuti token kedua tema. Footer pada kerangka publik mengelompokkan tautan menjadi Jelajahi, Mulai, dan Informasi dengan judul huruf kapital kecil, disertai identitas, penjelasan singkat aplikasi, hak cipta tim, catatan, dan versi API. Semua tautan menuju halaman yang tersedia; tidak menambahkan kontak atau akun sosial rekaan. Pada HP, bagian footer tersusun satu kolom dan target sentuh minimal 44 px. Direktori mempertahankan footer ringkas agar perluasan footer tidak menambah pergeseran tata letak saat hasil dimuat.

Direktori Toko, Direktori Penitip, dan Papan Kebutuhan menampilkan pencarian dalam satu kartu selebar area isi, sejajar dengan hasil; isian cari dan tombol Filter berada satu baris mulai tablet. Kartu toko dan penitip tersusun sebagai baris (identitas, syarat dan alasan, lalu skor dan aksi); kartu produk berupa grid tiga sampai empat kolom. Filter lokasi, kategori, dan syarat dibuka melalui tombol Filter, dengan penjelasan dan jumlah pilihan aktif. Panel filter berada di atas hasil pada HP maupun desktop dan menyesuaikan jumlah kolom dengan ruang yang tersedia. Pilihan Daftar/Peta menggunakan hasil pencarian yang sama; marker lokasi publik bisa diklik atau diaktifkan lewat keyboard untuk membuka detail. Daftar nama toko tetap tersedia sebagai alternatif peta. Direktori Penitip dan Papan Kebutuhan selalu menampilkan pencarian di atas hasil. Filter lanjutan berada di samping hasil mulai 1440 px; pada layar lebih kecil, pengguna membukanya melalui tombol Filter dan urutan. Hasil dan keadaan kosong tetap terlihat tanpa harus melewati seluruh isian filter. Foto produk ditampilkan jika tersedia; bila belum ada, gunakan placeholder ikon yang tenang, bukan foto rekaan. Profil usaha memakai kepala profil, kategori, keterangan, syarat, dan aksi relevan. Produk memiliki foto utama, harga jual saran, komposisi, pembuat, dan daftar tempat membeli. Kebijakan memakai kolom baca nyaman.

Panduan (`/panduan`) ringkas: delapan kartu bernomor dua kolom berisi judul, pelaku, dan ringkasan; langkahnya terlipat di `details` Lihat langkah. FAQ (`/faq`) memakai daftar lipat `details` per topik dengan pintasan topik di atas. Keduanya diprerender dan tertaut dari header publik, menu HP, dan footer.

Daftar putih publik tidak berubah: harga setor, angka transaksi, lokasi persis penitip, nomor WhatsApp, dan jumlah stok per toko tidak ditambahkan. Segel bekerja di belakang layar dan tidak tampil di antarmuka (ADR 0058); segel tidak disebut tanda tangan tersertifikasi.

### Akun dan identitas

Masuk, registrasi, dan demo memakai kerangka akun khusus. Pada desktop, panel biru penuh setengah layar di kiri memuat tombol kembali ke beranda dan merek di pojok kiri atas, lalu label Untuk penitip dan toko, judul, penjelasan, ilustrasi, dan keterangan di tengah; formulir dalam kartu di kanan. Pada HP, panel menjadi bilah biru berisi tombol kembali dan merek di atas formulir. Kerangka akun tidak memuat pilihan tema dan footer; tema mengikuti preferensi tersimpan atau sistem. Semua teks tetap terbaca pada kedua tema, kontrol dapat dijangkau dengan keyboard, dan kata sandi dapat diperlihatkan melalui tombol berlabel. Formulir masuk dan registrasi tidak menampilkan akun contoh. Demo dari beranda diarahkan ke `/demo`, dengan captcha dan alur masuk yang sama. Halaman demo menampilkan judul di atas kartu, lalu satu kartu berisi tiga akun contoh: Bendi (penitip), Zara (toko), dan Jessica (toko yang menunggu jawaban Ajuan), masing-masing satu baris dengan inisial, peran, email, dan tombol Coba sebagai di kanan, diikuti formulir masuk.

Logo aplikasi menggunakan gambar dua mitra biru dan paket jingga yang diberikan pengguna pada 10 Oktober 2026. Gambar dipakai pada merek, favicon, ikon instalasi dan ikon Apple; ekspor ukuran tidak mengubah bentuk atau warna logo. Biru tetap menjadi warna utama, sedangkan jingga digunakan sebagai aksen ilustrasi.

### Operasional

Judul modul memakai ikon Lucide yang sesuai dalam bidang persegi membulat. Katalog produk berupa grid kartu dengan foto asli, kategori, status, nama, SKU, dan harga; produk tanpa foto memakai placeholder ikon. Filter berada dalam panel permukaan terpisah. Semua kartu angka memakai permukaan tema (putih pada tema terang) dengan ikon berwarna semantik; tidak ada kartu angka gelap. Tombol tambah di kepala halaman (Hitung sisa, Catat titipan, Pasang Kebutuhan, Tambah produk, Undang mitra) memakai gaya tombol utama yang sama; tindakan pendamping memakai tombol sekunder. Kepala kartu memisahkan judul dan isi dengan garis halus dan ikon fitur. Formulir operasional dikelompokkan dalam kartu; kontrol memiliki latar permukaan, fokus eksplisit dan label permanen. Tab memakai pilihan bersegmen. Tidak ada emoji, statistik rekaan, grafik dekoratif yang menyamar sebagai data, atau dependensi baru.

Dasbor toko: lima angka spesifikasi, Hitung cepat dan Perlu tindakan, lalu grafik Penjualan harian berdampingan dengan donat Pembagian omzet (khusus OWNER), Pengingat selebar area isi, diikuti Produk terlaris selebar isi. Hitung cepat dan Perlu tindakan memiliki tinggi sama pada desktop; seluruh kartu Pengingat memiliki tinggi sama. Pengingat menampilkan tiga kartu awal; daftar lain menampilkan lima baris awal. Tombol buka/ringkas tetap menyediakan seluruh data. Dasbor penitip: empat angka spesifikasi, Perlu tindakan dan donat Pembagian omzet di kolom kanan berdampingan dengan grafik mulai desktop, lalu Pengingat selebar isi, lalu dua peringkat yang masing-masing selebar isi. Tombol di salam pembuka berbunyi Cari toko. Filter toko/produk berada di dalam kartu grafik dan mempertahankan fokus saat pilihan berubah. Pada HP, Perlu tindakan mendahului grafik. Pengingat menampilkan tiga kartu awal; peringkat menampilkan lima baris awal, dengan tombol buka/ringkas untuk sisanya. Kolom menyesuaikan ruang setelah sidebar; angka tidak diperkecil sampai sulit dibaca. Produk terlaris menampilkan foto asli atau placeholder bila belum ada foto, serta tautan detail produk. Informasi gabungan tetap hanya untuk OWNER.

Tab Penjualan dan Setoran di Laporan berada di tengah. Tab Setoran menampilkan empat kartu Terbit, Lunas, Belum lunas, dan Lewat tempo (jumlah · rupiah, berwarna sesuai status; Lewat tempo bertanda Perhatian bila ada), satu panel saringan (status, umur, rentang, cari mitra atau nomor Setoran), tabel Ringkasan per Mitra, lalu Daftar Tagihan Setoran. Setoran lewat tempo memakai kartu berbingkai merah. Laporan tab Penjualan menampilkan angka ringkasan, grafik kolom omzet harian, donat pembagian omzet, batang omzet per mitra, dan batang bertumpuk terjual serta dikembalikan per produk, lalu tabel dan unduhan CSV. Stok memakai grid dua kolom mulai tablet.

Riwayat, katalog, Ajuan, Kemitraan, Kebutuhan dan Setoran memakai filter di atas isi, status berlabel dan tindakan sesuai hak akses. Detail dokumen membedakan nomor/status, isi, jejak konfirmasi, dan langkah berikutnya. Tabel desktop berubah menjadi daftar berlabel di HP; tabel padat hanya boleh memiliki overflow lokal.

Titipan Masuk dan Hitung Sisa memakai baris produk dan kontrol jumlah besar. Ringkasan beserta simpan mudah dicapai, di atas navigasi. Rekap uang selalu berurutan omzet, untung toko, harus setor. Formulir lain dikelompokkan dalam permukaan yang jelas. Formulir produk memisahkan Informasi produk, Harga dan masa simpan, serta Deskripsi dan foto. Isian berpasangan berbagi dua kolom mulai tablet; catatan dan deskripsi memakai area teks beberapa baris. Pada halaman ubah produk, formulir tampil sebelum disclosure QR. Detail dan formulir memiliki tautan kembali ke daftar. Ringkasan dokumen memakai kartu label/nilai yang menjadi satu kolom pada HP. Profil (`/app/profil`) memuat kartu akun (foto atau ikon, nama, peran, Ubah foto profil dan Hapus foto; nama lengkap, email, nomor WhatsApp, dan kata sandi dengan tombol Ubah sandi yang membuka dialog; tombol Edit akun untuk nama dan WhatsApp) dan kartu Informasi profil usaha yang terkunci sampai tombol Edit profil ditekan, dengan Batal dan Simpan; profil toko atau penitip menyusul. Pengaturan memuat Tampilan (sakelar matahari/bulan), Tampil di Cari Mitra (bilah kelengkapan hijau saat lengkap), Info pembayaran untuk penitip yang terkunci sampai Edit, lalu Keluar dari semua perangkat. Pemberitahuan memuat judul dengan jumlah belum dibaca, tombol Tandai semua telah dibaca, pilihan Semua dan Belum dibaca, kartu belum dibaca bergaris biru dengan titik, waktu, dan tombol Tandai dibaca, serta Muat pemberitahuan lebih lama. Tindakan mengakhiri kemitraan berada dalam disclosure terpisah dan tetap membutuhkan dialog konfirmasi. Kontrol tema dan profil pada header HP menggunakan ikon dengan nama aksesibel agar header tetap ringkas; mulai tablet tampil label Mode dan Profil. Preferensi tema tersedia di header dan pengaturan, termasuk untuk staf yang tidak boleh mengubah profil.

## 6. Komponen dan keadaan

| Komponen | Aturan |
|---|---|
| Kartu | Permukaan sesuai tema, border dekoratif, radius 20 px dan bayangan tipis dari token |
| Angka kunci | Ikon sesuai data pada permukaan berwarna semantik, label sekunder, angka utama monospace, keterangan opsional |
| Tombol utama | Primary biru, teks putih, tinggi minimal 48 px; satu aksi utama per konteks |
| Sekunder | Permukaan sesuai tema, border kuat, teks utama |
| Bahaya | Border/teks bahaya, akibat dijelaskan di konfirmasi |
| Isian | Tinggi minimal 48 px, teks 16 px, border kuat dan fokus terlihat |
| Jumlah | Tombol tambah/kurang 48 px, papan angka, label yang jelas |
| Badge | Teks status, warna semantik dan latar yang sesuai |
| Grafik | SVG buatan sendiri tanpa pustaka: kolom biru untuk omzet harian, donat untuk pembagian omzet, batang mendatar (bisa bertumpuk) untuk perbandingan mitra dan produk. Grid halus, tooltip per kolom atau batang, nilai tertulis sebagai teks, alternatif tabel untuk grafik harian |
| Peta | Terang pada light, filter gelap hanya pada dark; kontrol dan atribusi selalu terbaca |
| Dialog, lembar bawah, toast | Token tema; fokus dan tindakan sesuai keadaan |
| Skeleton | Pola daftar, kartu, detail, formulir, laporan, profil, dan dasbor mengikuti susunan isi; warna kedua tema dan animasi menghormati preferensi pengurangan gerak |
| Kosong | Penjelasan singkat dan satu tindakan yang relevan |
| Luring | Pita status, jumlah antrean, peringatan usia lama, batas aksi jelas |

Memuat, kosong, berhasil, validasi gagal, jaringan gagal, dan luring wajib ditangani. Galat isian tampil dekat isian, fokus berpindah ke yang pertama salah, dan nilai tidak hilang. Menyimpan menonaktifkan kiriman ganda. Konflik antrean memperlihatkan catatan perangkat dan keadaan server dengan label jelas. Seluruh aturan bisnis dan tenggat konfirmasi tetap berlaku.

Saat berpindah halaman, cangkang navigasi tetap tampil dan isi tujuan memakai skeleton sesuai jenis halaman selama rute menunggu data atau kode. Isi sebelumnya tetap terpasang agar pembatalan navigasi tidak menghapus isian. Pemuatan akun dan data memakai pola yang sama; halaman yang sudah siap tidak diberi jeda buatan. Profil publik tetap dirender di server.

## 7. Aksesibilitas, gerak, dan performa

Target sentuh minimal 44 px; kontrol utama 48 px. Fokus selalu terlihat; heading berurutan; landmark utama tunggal per kerangka; label dan nama tindakan jelas. Uang penting tidak disampaikan lewat warna saja. Tabel memiliki caption, gambar memiliki alternatif, dan perubahan penting memakai status/alert yang tepat.

Gerak tidak mengganggu pekerjaan. Transisi kontrol singkat 150 sampai 200 ms; reduced motion mematikan gerak nonesensial. Angka uang langsung berubah tanpa animasi menghitung naik.

Peta, grafik, QR dan captcha tetap dimuat sesuai kebutuhan. Preferensi tema berupa skrip kecil dari asal sendiri, tanpa dependensi baru. JS awal publik maksimal 90 KB terkompresi dan pencatatan 120 KB. Sasaran Lighthouse HP performa/aksesibilitas minimal 90, LCP di bawah 2,5 detik, CLS di bawah 0,1. Ketercapaian dicatat setelah pengukuran.

## 8. Token dan pemeriksaan

Warna, ukuran, jarak, radius, lapisan dan safe area didefinisikan pada `app/src/lib/styles/tokens.css`. Tailwind memetakan token tersebut; komponen tidak menulis nilai mentah. Teks pengguna hanya dari katalog `i18n`.

Pemeriksaan wajib: format, lint, tipe, batas lapisan, komentar, token; uji unit/komponen, alur kedua peran, luring, preferensi tersimpan, tema sistem, aksesibilitas kedua tema dan tata letak HP/desktop. Ikon instalasi PWA memakai biru yang sama dengan merek layar.

Cetak dan dokumen proposal tetap memakai hitam di atas putih. Aturan font dan margin proposal lomba tidak diterapkan ke font website.
