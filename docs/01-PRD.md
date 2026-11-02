# PRD: Trustip

> Platform titip jual untuk UMKM. Pemilik produk dan toko saling menemukan, menyepakati aturan dagang secara terbuka, lalu mencatat titipan, sisa, dan setoran di satu data yang sama.

| Item | Keterangan |
|---|---|
| Kompetisi | GTNIC 2026, kategori Web Development |
| Subtema | Accelerating Digital Transformation for Business and SMEs |
| Fokus industri | UMKM dan Digital Economy |
| Versi dokumen | 2.1 |
| Tanggal | 7 Oktober 2026 |
| Status | Draf revisi untuk ditinjau tim |
| Dokumen pasangan | `02-SPESIFIKASI-FITUR.md`, `03-ATURAN-REKAYASA.md`, `04-RENCANA-IMPLEMENTASI.md`, `DESIGN.md` |

## Perubahan dari versi 1.0

| # | Perubahan | Alasan |
|---|---|---|
| 1 | Lingkup naik dari alat kelola menjadi platform: ditambah Cari Mitra dua arah, Kebutuhan Toko, serta Ajuan Kerja Sama dengan tawar-menawar | Permintaan tim: UMKM bisa mencari toko, toko bisa mencari produk |
| 2 | Titipan Masuk dicatat oleh toko saat barang ditaruh. Hitung Sisa saat tutup atau saat barang diambil menggantikan laporan penjualan terpisah, dan retur menyatu di dalamnya | Mengikuti cara kerja titip jual harian di warung dan toko |
| 3 | "Tagihan" menjadi "Setoran" yang terbit otomatis, bisa langsung setiap kali hitung | Toko perlu tahu saat itu juga berapa yang harus dibayar |
| 4 | Klaim "satu-satunya" dihapus. Tabel pesaing diganti dengan yang bersumber | Ada aplikasi dan proyek sejenis yang tidak tercantum di v1.0 |
| 5 | Bukti masalah memakai sumber konkret, bukan sebutan umum "topik penelitian" | Bagian latar belakang harus bisa diperiksa juri |
| 6 | Halaman publik produk dan profil naik menjadi prioritas utama | Tantangan lomba meminta layanan yang lebih baik bagi pelanggan |
| 7 | Arsitektur diganti: API Go, antarmuka SvelteKit sebagai PWA, pencarian lokasi dengan PostGIS, pembaruan langsung lewat SSE, berkas disimpan di luar basis data | Kriteria teknis berbobot 20%. Rancangan v1.0 menghabiskan kuota basis data gratis dan mengklaim RLS yang tidak aktif |
| 8 | Jadwal bertanggal diganti urutan pengerjaan bergerbang | Tim memutuskan tidak mengikat rencana pada tanggal |

## Perubahan dari versi 2.0

| # | Perubahan | Alasan |
|---|---|---|
| 1 | Nama produk menjadi Trustip | Keputusan tim |
| 2 | Mencatat Titipan Masuk dan Hitung Sisa saat luring naik dari P1 ke P0 | Sinyal di warung dan kantin sering putus, dan pencatatan tidak boleh menunggu |
| 3 | Segel dokumen bertanda tangan digital | Transparansi dua pihak butuh bukti bahwa catatan tidak diubah setelah terjadi |
| 4 | Captcha, kata sandi dengan salt dan pepper, tanda tangan permintaan ke API | Pengerasan keamanan. Kriteria teknis berbobot 20% |
| 5 | Tampilan mengikuti `DESIGN.md`: tema terang/gelap dengan primary biru dan angka monospace | Arah desain yang dipilih tim |
| 6 | Ditambah `03-ATURAN-REKAYASA.md` dan `04-RENCANA-IMPLEMENTASI.md`. Urutan pengerjaan menjadi 13 tahap | Aturan kode dan urutan kerja dibuat eksplisit dan bisa diperiksa |

## 1. Ringkasan

Titip jual (konsinyasi) adalah cara jual yang lazim bagi UMKM produsen: barang ditaruh di warung, kantin, toko oleh-oleh, atau kedai, dan uang baru diterima setelah barang laku. Dua hal membuat cara ini melelahkan. Pertama, mencari pasangan dagang: produsen tidak tahu toko mana yang mau menerima titipan, dan toko tidak tahu siapa yang punya produk yang dibutuhkannya. Kedua, mencatatnya: titipan, sisa, retur, dan pembayaran tersebar di buku nota dan chat.

Trustip menyatukan keduanya dalam empat bagian:

1. **Cari Mitra.** Pemilik produk mencari toko yang menerima titipan, diurutkan dari yang terdekat. Toko mencari produk, atau memasang Kebutuhan Toko ("butuh camilan untuk kantin, 50 bungkus per minggu").
2. **Kesepakatan Dagang.** Kedua pihak menyepakati aturan di web: harga jual, harga setor atau persen komisi, jadwal setor, dan nasib barang tak laku. Keduanya melihat angka yang sama.
3. **Kelola Titipan.** Toko mencatat barang yang ditaruh. Saat tutup atau saat barang diambil, toko mengisi sisa. Sistem langsung menampilkan terjual berapa, omzet, untung toko, dan yang harus disetor.
4. **Laporan dan Saran.** Laporan penjualan untuk kedua pihak, peringatan kedaluwarsa, saran jumlah titip, dan saran mitra yang cocok.

Satu kalimat nilai: **dari mencari tempat titip sampai uang disetor, kedua pihak bekerja di catatan yang sama.**

## 2. Latar belakang dan masalah

### 2.1 Alur saat ini

1. Produsen berkeliling atau bertanya lewat kenalan untuk mencari toko yang mau dititipi.
2. Aturan disepakati lisan: harga, bagian toko, kapan dibayar.
3. Barang ditaruh, dicatat di nota kertas atau buku masing-masing.
4. Produsen datang atau menelepon untuk tahu laku berapa.
5. Sisa dihitung, bagian toko dihitung manual, uang diserahkan.
6. Barang tak laku atau kedaluwarsa ditarik, kalau ingat.

### 2.2 Masalah dan buktinya

| Kode | Masalah | Dialami | Bukti |
|---|---|---|---|
| M1 | Produsen sulit menjangkau toko, dan syarat tiap toko berbeda | Penitip | BPOM menyebut keterbatasan akses pasar sebagai salah satu tantangan UMK pangan olahan [S8]. Satu toko oleh-oleh mensyaratkan sampel, izin edar, sertifikat halal, formulir penawaran, lalu evaluasi sekitar satu minggu [S9] |
| M2 | Penitip harus datang ke toko untuk tahu jumlah terjual | Penitip | Studi kasus UMKM kerupuk kulit di Jakarta: pemilik perlu datang langsung ke lokasi mitra untuk memeriksa catatan penjualan [S1] |
| M3 | Catatan dua pihak tidak sama, lalu timbul selisih dan konflik | Keduanya | Pada studi yang sama, pembayaran bergantung pada komunikasi lisan dan hitungan manual, sehingga membuka selisih data, keterlambatan, sampai konflik [S1] |
| M4 | Pembayaran terlambat | Penitip | Studi titip jual kue tradisional di Aceh mencatat telat bayar sebagai bentuk wanprestasi yang sering terjadi [S3] |
| M5 | Barang tak laku dan kedaluwarsa tidak tercatat | Keduanya | Retur belum sistematis sehingga barang menumpuk tanpa catatan jelas [S1]. Di satu usaha roti, retur berkisar 5,81% sampai 29,72% dari produksi, dengan kerugian sekitar Rp2.353.000 per bulan [S4] |
| M6 | Jumlah titip ditentukan dengan kira-kira | Penitip | Pada usaha roti yang sama, retur terjadi karena jumlah produksi tidak sebanding dengan permintaan aktual [S4] |
| M7 | Toko yang menerima banyak penitip sulit melacak kewajiban ke tiap penitip | Toko | Belum ada sumber. Wajib diuji lewat wawancara (Lampiran A) |
| M8 | Toko tidak tahu ke mana mencari produk baru untuk mengisi rak | Toko | Belum ada sumber. Wajib diuji lewat wawancara (Lampiran A) |

Praktik kertas yang baik sebenarnya sudah punya jawabannya: kartu stok konsinyasi yang dipegang toko dan produsen, dengan kolom titip, laku, retur, sisa, setoran, dan paraf [S5]. Trustip adalah versi digital kartu itu, ditambah cara menemukan pasangan dagangnya.

### 2.3 Ukuran pasar

- Jumlah usaha industri mikro dan kecil menembus 4,46 juta unit pada 2025 [S6].
- Industri makanan adalah kelompok terbesar: sekitar 1,9 juta usaha atau 44% pada 2024 [S7].

Kedua angka ini batas atas, bukan jumlah pelaku titip jual. Kami tidak menemukan data berapa persen yang memakai skema ini. Angkanya berasal dari sumber sekunder yang mengutip BPS, jadi cek ke publikasi BPS sebelum masuk proposal.

### 2.4 Akar masalah

1. **Tidak ada tempat bertemu.** Informasi "toko ini menerima titipan" dan "produk ini mencari toko" hanya beredar lewat kenalan.
2. **Tidak ada satu catatan yang dipercaya dua pihak.** Selisih, telat bayar, dan retur tak tercatat adalah gejalanya.

### 2.5 Mengapa web dan PWA

1. Dua pihak yang berbeda usaha harus melihat data yang sama. Itu hanya bisa lewat layanan daring bersama, bukan aplikasi luring di satu HP.
2. Pencatatan terjadi di depan rak, dengan HP. PWA bisa dipasang ke layar utama tanpa toko aplikasi, ringan, dan tetap terbuka saat sinyal lemah.
3. Halaman profil dan produk yang terbuka bagi umum membuat toko, produk, dan Kebutuhan Toko bisa ditemukan lewat tautan dan mesin pencari.

## 3. Analisis pesaing

Hasil penelusuran web per 7 Oktober 2026.

| Solusi | Yang dilakukan | Yang tidak ada |
|---|---|---|
| Buku nota dan WhatsApp | Cara bawaan | Semua masalah di bagian 2.2 |
| Titip Dagang (Android) [S10] | Titik toko berbasis GPS, isi sisa lalu terjual dihitung otomatis, kalkulator HPP, akun staf, ekspor. Tercatat 50+ unduhan | Dipakai dari sisi penitip saja. Tidak ada pencarian toko baru, kesepakatan dua pihak, atau akun toko |
| Titipku versi pencatatan konsinyasi (Android) [S11] | Catat titipan banyak toko, laporan terjual dan tagihan, ekspor Excel. Tercatat 50+ unduhan | Sama: satu sisi, tanpa pencarian mitra |
| Sistem konsinyasi berbasis web, proyek Telkom University [S1][S2] | Dua sisi untuk satu UMKM: staf mitra masuk lalu mengisi penjualan, retur, dan pembayaran. Skor SUS 81,87 | Dibuat untuk satu usaha dan mitranya. Tidak ada pencarian mitra, banyak penitip, atau kesepakatan per produk |
| Modul konsinyasi di perangkat lunak akuntansi dan kasir [S13][S14] | Konsinyasi keluar, retur, piutang, faktur | Dirancang untuk admin kantor satu perusahaan. Pihak lawan tidak ikut memakai |
| Aplikasi kasir umum [S12] | Kasir dan stok milik toko | Salah satu yang populer menyatakan belum punya fitur khusus konsinyasi dan menyarankan akal-akalan lewat menu supplier dan retur |
| Jooalan [S15] | Satu perusahaan menitipkan produk pemasoknya ke reseller | Bukan tempat bertemu terbuka antara UMKM dan toko mana pun |
| PaDi UMKM [S16] | Pengadaan BUMN dari UMKM | Bukan titip jual ke toko ritel |
| Mitra Dagang [S17] | Katalog, pesanan, dan faktur untuk penjual | Tidak memodelkan titipan, sisa, dan bagi hasil |
| Repositori sumber terbuka konsinyasi sisi toko [S18] | Satu toko menampung titipan banyak UMKM, bagi hasil per periode | Satu sisi (toko), tanpa akun penitip dan tanpa pencarian |

**Posisi Trustip.** Sejauh penelusuran kami, belum ada layanan yang menggabungkan tiga hal ini sekaligus:

1. Pencarian mitra titip jual dua arah berbasis lokasi, termasuk papan Kebutuhan Toko.
2. Kesepakatan dagang per produk yang tercatat dan disetujui dua pihak.
3. Pencatatan titipan, sisa, dan setoran yang dipakai bersama oleh banyak penitip dan banyak toko.

Kalimat ini hasil penelusuran web, bukan jaminan bahwa tidak ada pemain lain. Di proposal, tulis sebagai "belum kami temukan", bukan "satu-satunya".

Catatan: aplikasi "Ka TiTiPiN" yang disebut di v1.0 tidak berhasil kami temukan lagi. Cek ulang sebelum dikutip.

## 4. Tujuan dan ukuran keberhasilan

### 4.1 Tujuan produk

| Kode | Tujuan | Tuntutan tantangan GTNIC |
|---|---|---|
| T1 | Pemilik produk dan toko bisa saling menemukan tanpa perantara kenalan | Layanan lebih baik bagi pelanggan (toko adalah pelanggan penitip) |
| T2 | Setiap kerja sama punya kesepakatan tertulis yang sama di kedua pihak | Mengelola proses operasional |
| T3 | Titipan, sisa, retur, dan setoran tercatat sekali dan dipercaya dua pihak | Mengelola proses operasional |
| T4 | Hitungan bagian toko dan setoran otomatis, tanpa selisih | Meningkatkan produktivitas |
| T5 | Kerugian barang tak laku turun lewat peringatan dan saran jumlah titip | Meningkatkan produktivitas |
| T6 | Pembeli akhir bisa mengenali produk, pembuatnya, dan tempat membelinya | Layanan lebih baik bagi pelanggan |

### 4.2 Ukuran keberhasilan

| Ukuran | Target |
|---|---|
| Waktu penitip menemukan toko terdekat yang cocok lalu mengirim ajuan | Di bawah 2 menit |
| Waktu toko mencatat titipan masuk 3 produk | Di bawah 30 detik |
| Waktu toko mengisi sisa 3 produk sampai rekap setoran tampil | Di bawah 45 detik |
| Selisih hitungan setoran dengan hitungan manual pada data uji | 0 rupiah |
| Pemberitahuan sampai ke pihak lain saat keduanya daring | Di bawah 3 detik |
| Lighthouse seluler: performa dan aksesibilitas | Minimal 90 dan 90 |
| Bisa dipasang sebagai PWA di Android dan iOS | Ya |
| Galat pada alur utama saat demo | 0 |
| Uji pengguna | Minimal 3 penitip dan 3 toko mencoba alur inti, skor SUS minimal 70 |

## 5. Pengguna

| Persona | Gambaran | Yang dicari |
|---|---|---|
| **Pak Anton**, penitip harian | Membuat roti manis, menaruh di 12 warung tiap pagi dan mengambil sisa sore harinya | Tahu harus menaruh berapa di tiap warung, uang hari itu jelas, dan tambah warung di jalur antarnya |
| **Bu Rina**, penitip mingguan | Merek keripik rumahan di 8 sampai 15 toko, antar dua kali seminggu | Tahu laku di mana tanpa menelepon, tagih tanpa debat, tahu barang yang hampir kedaluwarsa |
| **Bu Sri**, pemilik warung | Rak sering kosong, ingin jajanan baru tetapi tidak tahu harus mencari ke siapa | Produk yang cocok dan dekat, hitungan yang tidak merepotkan, tahu untung dari titipan |
| **Pak Dedi**, pemilik toko oleh-oleh | Menerima titipan dari 20 sampai 40 penitip, punya syarat izin edar dan halal | Satu tempat untuk semua kewajiban ke penitip, penitip yang memenuhi syarat |
| **Sinta**, staf toko | Mencatat barang masuk dan menghitung sisa | Layar yang cepat diisi dengan satu tangan |
| **Pembeli akhir** | Memindai QR di kemasan atau membuka tautan | Info produk, pembuat, dan toko terdekat yang menjualnya |
| **Admin platform** | Anggota tim | Menindak profil atau unggahan yang dilaporkan |

Setiap usaha di Trustip adalah **organisasi** bertipe Penitip atau Toko. Satu akun pengguna menjadi anggota satu organisasi.

## 6. Ruang lingkup

Prioritas menandai urutan, bukan tanggal. **P0** harus utuh dan lulus uji sebelum P1 disentuh.

### 6.1 P0: inti platform

| Modul | Isi |
|---|---|
| Akun dan organisasi | Daftar, masuk, pilih tipe usaha, profil, lokasi di peta |
| Cari Mitra | Direktori toko dan direktori penitip, urut terdekat, saring jenis toko dan kategori produk, skor kecocokan |
| Kebutuhan Toko | Toko memasang kebutuhan produk, penitip menanggapi |
| Ajuan Kerja Sama | Ajuan dari salah satu pihak, tawar-menawar syarat, terima atau tolak. Termasuk tautan undangan untuk mitra lama |
| Kesepakatan Dagang | Syarat per produk (harga setor atau komisi persen), syarat umum, versi, persetujuan dua pihak |
| Produk | Katalog penitip, kategori, masa simpan, legalitas |
| Titipan Masuk | Dicatat toko saat barang ditaruh, dikonfirmasi penitip |
| Stok | Buku stok per toko per batch, urutan kedaluwarsa |
| Hitung Sisa | Diisi toko saat tutup atau saat barang diambil, termasuk barang yang dikembalikan. Menghasilkan rekap terjual, omzet, untung toko, dan harus setor |
| Setoran dan pembayaran | Setoran terbit otomatis per hitung, mingguan, atau bulanan. Toko mencatat bayar, penitip mengonfirmasi |
| Laporan | Laporan penjualan toko dan penitip, rekap harian, ekspor Excel |
| Dasbor dan saran | Angka kunci, daftar perlu tindakan, peringatan kedaluwarsa, stok menipis, saran jumlah titip, saran mitra |
| Pemberitahuan | Di dalam aplikasi secara langsung, ditambah tautan WhatsApp siap kirim |
| Halaman publik | Profil toko, profil penitip, halaman produk dengan QR dan daftar toko yang menjual |
| Reputasi dan pelaporan | Angka reputasi dari data transaksi, tombol laporkan |
| PWA dan kerja luring | Bisa dipasang, cangkang tersimpan, data terakhir terbaca saat luring. Titipan Masuk dan Hitung Sisa bisa dicatat tanpa sinyal lalu terkirim sendiri |
| Segel dokumen | Setiap peristiwa dagang disegel dengan tanda tangan digital berantai dan bisa diperiksa di halaman verifikasi |

### 6.2 P1: setelah P0 lulus uji

- Kiriman yang diumumkan penitip lebih dulu, lalu dikonfirmasi toko
- Sanggahan atas titipan atau hitungan, dengan koreksi yang disetujui dua pihak
- Pembayaran sebagian
- Anggota staf dengan undangan
- Tampilan peta di Cari Mitra
- Ulasan dan bintang antar mitra
- Halaman admin untuk moderasi
- Pemberitahuan dorong (web push)
- Unduh PDF kesepakatan, rekap hitung, dan setoran, dengan kode QR segel
- Saran redistribusi antar toko dan penanda produk lambat

### 6.3 P2: dicatat sebagai pengembangan lanjutan

- Pembayaran daring (QRIS dinamis) dengan konfirmasi otomatis
- Pengiriman WhatsApp otomatis lewat WhatsApp Business API
- Masuk dengan akun Google
- Satu usaha berperan ganda sebagai penitip dan toko
- Pemindaian barcode dengan kamera
- Model biaya sewa rak dan model beli putus
- Prakiraan permintaan dengan model statistik (misalnya newsvendor [S4])
- Integrasi kasir toko dan perangkat lunak akuntansi
- Verifikasi legalitas ke sumber resmi
- Persetujuan yang ditandatangani dengan kunci di perangkat pengguna

### 6.4 Di luar lingkup

- Menjual langsung ke pembeli akhir (keranjang belanja, ongkos kirim)
- Menalangi atau menjamin pembayaran antar pihak
- Pengantaran barang

## 7. Glosarium

| Istilah | Arti |
|---|---|
| Penitip | Organisasi pemilik produk yang menaruh barang di toko |
| Toko | Organisasi yang menjual barang titipan |
| Cari Mitra | Direktori dua arah untuk mencari toko atau penitip |
| Kebutuhan Toko | Unggahan toko tentang produk yang sedang dicarinya |
| Ajuan Kerja Sama | Usulan syarat dari satu pihak ke pihak lain, bisa ditawar |
| Kesepakatan Dagang | Syarat yang sudah disetujui dua pihak. Punya nomor versi |
| Kemitraan | Hubungan aktif antara satu penitip dan satu toko |
| Harga jual | Harga satuan ke pembeli di toko |
| Model **Harga Setor** | Toko menyetor nilai tetap per unit terjual. Selisih dengan harga jual adalah untung toko |
| Model **Komisi** | Toko mengambil persen dari harga jual. Sisanya hak penitip |
| Titipan Masuk | Catatan barang yang ditaruh di toko pada satu waktu |
| Batch | Unit dari satu baris Titipan Masuk dengan tanggal kedaluwarsa yang sama |
| Hitung Sisa | Toko menghitung sisa fisik. Selisih dengan stok tercatat dianggap terjual |
| Retur | Sisa yang dikembalikan ke penitip. Tidak dibayar |
| Setoran | Jumlah yang wajib dibayar toko ke penitip untuk barang terjual |
| Omzet | Jumlah unit terjual dikali harga jual |
| Untung toko | Omzet dikurangi hak penitip |
| Hak penitip | Bagian penitip dari omzet. Sama dengan nilai Setoran |
| Skor kecocokan | Angka 0 sampai 100 yang merangkum jarak, kategori, jenis toko, reputasi, dan keaktifan |
| Segel | Tanda tangan digital platform atas satu peristiwa dagang, berantai per kemitraan, yang membuktikan catatan tidak berubah |
| Antrean luring | Catatan yang dibuat tanpa sinyal dan menunggu terkirim dari perangkat |

## 8. Alur utama

### 8.1 Alur A: penitip mencari toko

1. Pak Anton membuka Cari Mitra. Daftar toko muncul terurut dari yang terdekat dengan lokasi usahanya.
2. Ia menyaring "Warung" dan "Kantin", lalu melihat Warung Bu Sri berjarak 1,2 km dengan tanda "Butuh roti dan kue basah".
3. Ia membuka profil warung, melihat syarat dan reputasi bayar, lalu menekan **Ajukan kerja sama**.
4. Ia memilih Roti Manis, mengusulkan harga jual Rp2.000 dan harga setor Rp1.800, setor setiap hitung.
5. Bu Sri menerima pemberitahuan, menawar harga setor Rp1.750. Pak Anton menerima.
6. Kemitraan aktif. Kesepakatan versi 1 tersimpan dan terlihat sama di kedua akun.

### 8.2 Alur B: toko mencari produk

1. Bu Sri memasang Kebutuhan Toko: "Camilan anak sekolah, harga jual di bawah Rp3.000, sekitar 40 bungkus per minggu".
2. Kebutuhan itu muncul di papan kebutuhan dan di hasil pencarian penitip sekitar.
3. Bu Rina menanggapi dengan ajuan yang berisi Keripik Mini dan syarat komisi 20%.
4. Bu Sri juga bisa membuka direktori penitip, melihat produk yang sedang mencari toko, lalu mengirim ajuan lebih dulu.

### 8.3 Alur C: titip harian

1. Pagi hari Pak Anton menaruh 20 roti. Bu Sri membuka **Catat titipan masuk**, memilih Pak Anton, mengisi 20. Layar menampilkan saran: "Biasanya laku 17 pada hari Rabu. Saran titip 19."
2. Pak Anton menerima pemberitahuan dan menekan **Sesuai**.
3. Saat tutup, Bu Sri membuka **Hitung sisa**, mengisi sisa 2, menandai keduanya dikembalikan.
4. Rekap tampil: terjual 18, omzet Rp36.000, untung warung Rp4.500, harus setor Rp31.500.
5. Bu Sri membayar tunai dan menekan **Catat bayar**. Pak Anton menekan **Uang diterima**. Setoran lunas.

### 8.4 Alur D: titip mingguan

1. Bu Rina menaruh 40 bungkus keripik di toko Pak Dedi. Sinta mencatat titipan masuk beserta tanggal kedaluwarsa.
2. Setiap Sabtu Sinta mengisi Hitung Sisa. Stok dan penjualan di akun Bu Rina ikut berubah.
3. Senin dini hari sistem menerbitkan Setoran minggu lalu. Pak Dedi mentransfer, mengunggah bukti, Bu Rina mengonfirmasi.
4. Bu Rina menerima peringatan: "8 bungkus di Toko C kedaluwarsa dalam 5 hari."

### 8.5 Alur E: pembeli akhir

1. Pembeli memindai QR di kemasan keripik.
2. Halaman produk menampilkan pembuat, komposisi, keterangan izin edar dan halal yang dinyatakan penitip, serta daftar toko terdekat yang sedang menjualnya.

## 9. Kesepakatan Dagang yang transparan

Kedua pihak selalu melihat tiga angka yang sama untuk setiap produk: harga jual, hak penitip per unit, dan untung toko per unit.

| Model | Yang disepakati | Contoh | Hak penitip per unit | Untung toko per unit |
|---|---|---|---|---|
| Harga Setor | Harga jual dan harga setor | Roti dijual Rp2.000, setor Rp1.750 | Rp1.750 | Rp250 |
| Komisi | Harga jual dan persen komisi toko | Keripik dijual Rp15.000, komisi 20% | Rp12.000 | Rp3.000 |

Syarat umum yang ikut disepakati:

- Jadwal setor: setiap hitung, mingguan, atau bulanan
- Tempo bayar setelah setoran terbit
- Irama hitung yang dijanjikan toko: harian, setiap kunjungan, atau mingguan
- Barang tak laku dan kedaluwarsa dikembalikan ke penitip dan tidak dibayar
- Barang yang hilang di toko dianggap terjual
- Catatan bebas, misalnya letak pajang

Aturan mainnya:

1. Kesepakatan baru berlaku setelah disetujui pihak yang tidak mengusulkannya. Nama penyetuju dan waktunya tercatat.
2. Setiap perubahan menjadi versi baru. Versi lama tetap bisa dibaca.
3. Harga dan bagian toko **dikunci ke barang saat barang ditaruh**. Perubahan kesepakatan hanya berlaku untuk titipan berikutnya.

## 10. Kebutuhan non-fungsional (ringkas)

Rinciannya ada di `02-SPESIFIKASI-FITUR.md` bagian 20.

| Aspek | Kebutuhan |
|---|---|
| Responsif | 360 sampai 1440 piksel, dirancang dari layar HP. Chrome, Firefox, Safari, Edge |
| PWA | Bisa dipasang, cangkang aplikasi tersimpan, data terakhir terbaca saat luring. Titipan Masuk dan Hitung Sisa bisa dicatat saat luring |
| Performa | LCP halaman publik di bawah 2,5 detik pada 4G lambat. API baca di bawah 150 ms (p95) |
| Keamanan | Data antar organisasi terpisah dan diuji, hak peran dan hak data diperiksa di server, kata sandi Argon2id dengan salt dan pepper, captcha, sesi di cookie HttpOnly, pembatasan laju, API hanya menerima permintaan bertanda tangan |
| Integritas | Semua perubahan stok dan uang berjalan dalam transaksi. Uang disimpan sebagai bilangan bulat rupiah |
| Jejak | Stok dihitung dari buku mutasi, bukan angka yang ditimpa. Setiap persetujuan tercatat dan disegel dengan tanda tangan digital |
| Privasi | Lokasi persis usaha rumahan tidak ditampilkan ke umum. Nilai uang hanya terlihat oleh pihak dalam kemitraan |
| Aksesibilitas | Kontras WCAG AA, label lengkap, target sentuh minimal 44 piksel |
| Bahasa | Bahasa Indonesia, format Rupiah, zona waktu Asia/Jakarta |

## 11. Arsitektur dan teknologi (ringkas)

Rinciannya ada di `02-SPESIFIKASI-FITUR.md` bagian 21. Tabel ini bisa disalin ke proposal. Kode berada di dua folder, `api` dan `app`, dan ditulis mengikuti `03-ATURAN-REKAYASA.md`.

| Lapisan | Pilihan | Alasan |
|---|---|---|
| API | Go dengan Fiber v3 | Cepat, hemat memori, satu berkas biner, cocok untuk koneksi langsung (SSE) dan tugas terjadwal |
| Basis data | PostgreSQL di Neon dengan PostGIS dan pg_trgm | Pencarian "terdekat" dengan indeks spasial, pencarian teks cepat, layanan terkelola [S19] |
| Akses data | pgx dan sqlc | Kueri ditulis tangan dan diperiksa saat kompilasi, tanpa ORM |
| Antarmuka | SvelteKit 2, Svelte 5, TypeScript, Tailwind CSS 4 | Bundel kecil untuk HP kelas bawah, dukungan service worker bawaan untuk PWA |
| Pembaruan langsung | Server-Sent Events | Pemberitahuan sampai dalam hitungan detik tanpa polling yang membangunkan basis data terus-menerus |
| Berkas | Penyimpanan di luar basis data lewat satu antarmuka penyimpanan | Foto tidak memakan kuota basis data |
| Keutuhan catatan | Segel Ed25519 berantai per kemitraan | Kedua pihak bisa memeriksa bahwa catatan tidak diubah |
| Kerja luring | Service worker, IndexedDB, antrean catat dengan kunci idempotensi | Toko tetap bisa mencatat tanpa sinyal, tanpa dokumen ganda |
| Keamanan tepi | Captcha Turnstile, tanda tangan HMAC dari proxy ke API | Menahan bot dan mencegah API dipanggil tanpa lewat app |
| Hosting | Antarmuka di Vercel, API di VPS di belakang Caddy | Halaman publik dekat pengguna, API tanpa jeda mulai dingin |
| Uji | go test, Vitest, Playwright, GitHub Actions | Fungsi uang dan stok diuji otomatis di setiap perubahan |

## 12. Kesesuaian dengan rubrik GTNIC

| Kriteria | Bobot | Jawaban Trustip |
|---|---|---|
| Relevansi masalah dan inovasi | 20% | Masalah M1 sampai M6 bersumber. Inovasi: tempat bertemu dua arah, kesepakatan yang dikunci ke barang, saran jumlah titip, skor kecocokan |
| Fungsionalitas dan kualitas implementasi | 25% | Alur utuh: cari, ajukan, sepakat, titip, hitung, setor, lapor. Data contoh 30 hari |
| Kualitas teknis dan arsitektur | 20% | Buku stok, fungsi uang murni dengan uji, pencarian spasial, isolasi data teruji, segel bertanda tangan digital, kerja luring tanpa dokumen ganda, aturan kode yang diperiksa mesin, CI |
| UX dan desain antarmuka | 15% | Dua angka per produk untuk titip dan hitung, rekap seketika, PWA yang tetap bisa mencatat saat luring, sistem desain bertoken, umpan balik konsisten |
| Dampak bisnis dan skalabilitas | 10% | Banyak penitip dan banyak toko sejak awal. Makin banyak pengguna, direktori makin berguna |
| Dokumentasi, demo, presentasi | 10% | README, diagram, akun demo, skenario video |

Tiga tuntutan tantangan lomba:

| Tuntutan | Jawaban |
|---|---|
| Mengelola proses operasional | Kesepakatan, titipan masuk, stok, hitung sisa, setoran |
| Meningkatkan produktivitas | Tidak perlu datang untuk tahu laku berapa, hitungan otomatis, saran jumlah titip |
| Layanan lebih baik bagi pelanggan | Toko (pelanggan penitip) mendapat syarat jelas dan pasokan yang pas. Pembeli akhir mendapat halaman produk dan lokasi toko |

## 13. Risiko dan mitigasi

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Direktori kosong di awal (masalah ayam dan telur) | Cari Mitra tidak berguna | Alat kelola tetap berguna tanpa direktori lewat tautan undangan untuk mitra lama. Mulai dari satu wilayah. Kebutuhan Toko memberi alasan bagi toko untuk mendaftar |
| Toko enggan mencatat | Data berhenti mengalir | Input dibuat dua angka per produk (masuk, sisa). Toko mendapat nilai sendiri: untung dan kewajiban per penitip. Penitip bisa mengingatkan lewat WhatsApp |
| Sengketa jumlah | Kepercayaan turun | Setiap titipan dan hitungan dikonfirmasi pihak lain atau dianggap sesuai setelah 48 jam. Buku stok bisa ditelusuri |
| Profil atau kebutuhan palsu | Direktori tidak dipercaya | Reputasi dihitung dari transaksi nyata, tombol laporkan, batas jumlah ajuan per hari |
| Lokasi rumah produsen tersebar | Risiko privasi | Lokasi penitip dibulatkan untuk umum, alamat lengkap hanya untuk mitra |
| Lingkup besar | Banyak fitur setengah jadi | P0 bergerbang: tiap tahap harus lulus uji ujung ke ujung sebelum tahap berikutnya. P1 tidak disentuh sebelum P0 lulus |
| Kuota basis data gratis habis | Aplikasi mati saat dinilai | Berkas di luar basis data, tanpa polling, penjaga hidup hanya pada jam penjurian |
| Angka hasil hitung salah | Kehilangan kepercayaan juri dan pengguna | Fungsi uang dan stok adalah fungsi murni dengan kasus uji tetap, termasuk pembulatan |
| Catatan luring bentrok dengan data terbaru | Angka stok salah | Server menolak catatan yang dasarnya sudah berubah dan pengguna meninjaunya di samping keadaan terbaru. Urutan kirim dijaga per kemitraan |

## 14. Batasan sistem

Bisa disalin ke subbab "Batasan sistem aplikasi" di proposal.

1. Pembayaran dicatat dan dikonfirmasi manual. Belum ada pembayaran daring.
2. Pesan WhatsApp berupa tautan dengan pesan terisi, belum terkirim otomatis.
3. Legalitas (NIB, izin edar, halal) adalah pernyataan pengguna, belum diverifikasi ke sumber resmi.
4. Jumlah terjual berasal dari hitung sisa. Barang hilang di toko terhitung sebagai terjual.
5. Jarak dihitung garis lurus, bukan jarak tempuh.
6. Saran jumlah titip dan skor kecocokan berbasis aturan sederhana, bukan model pembelajaran mesin.
7. Saat luring hanya Titipan Masuk dan Hitung Sisa yang bisa dicatat. Aksi lain memerlukan koneksi. Di iPhone, catatan luring terkirim saat aplikasi dibuka kembali.
8. Satu usaha hanya bisa berperan sebagai penitip atau toko, tidak keduanya.
9. Segel digital membuktikan catatan tidak berubah sejak dicatat platform. Segel bukan tanda tangan elektronik tersertifikasi.

## 15. Rencana pengembangan lanjutan

Lihat bagian 6.3. Urutan yang disarankan: pembayaran QRIS, WhatsApp otomatis, peran ganda, prakiraan permintaan, integrasi kasir.

## 16. Urutan pengerjaan

Tanpa tanggal. Setiap tahap punya gerbang yang harus lulus sebelum lanjut. Daftar fitur per tahap ada di `04-RENCANA-IMPLEMENTASI.md`.

| Tahap | Isi | Gerbang |
|---|---|---|
| 0 | Kerangka repositori dan pagar mutu | Klon baru lulus pemeriksaan dan uji dengan satu perintah. CI hijau |
| 1 | Fondasi API | Permintaan tanpa tanda tangan ditolak. Semua jawaban berbentuk baku |
| 2 | Fondasi app | Halaman contoh memanggil API lewat proxy. Aplikasi bisa dipasang |
| 3 | Akun dan sesi | Daftar, masuk, segarkan, keluar lulus uji. Token tidak terbaca skrip |
| 4 | Organisasi, lokasi, hak akses | Dua organisasi berbeda tipe bisa mendaftar. Uji isolasi data lulus |
| 5 | Produk, Ajuan lewat undangan, Kesepakatan, Kemitraan, Segel | Dua akun bermitra dengan syarat yang sama di kedua layar dan segel sah |
| 6 | Titipan Masuk, Stok, Hitung Sisa | Titip 20, sisa 2, rekap benar sampai rupiah. Uji fungsi murni lulus |
| 7 | Setoran, pembayaran, Laporan | Setoran terbit otomatis untuk tiga jadwal. Angka laporan cocok dengan Setoran |
| 8 | Kerja luring | Catatan luring terkirim tepat satu kali. Bentrok bisa ditinjau |
| 9 | Cari Mitra, Kebutuhan Toko, tawar-menawar Ajuan | Alur A dan B di bagian 8 lulus uji ujung ke ujung |
| 10 | Dasbor, saran, reputasi | Setiap saran muncul pada data contoh yang memicunya |
| 11 | Pemberitahuan langsung, halaman publik, verifikasi segel | Pemberitahuan lintas peramban di bawah 3 detik. Lighthouse memenuhi target |
| 12 | Data contoh, pengerasan, deploy, dokumentasi | Situs live dengan akun demo. Daftar periksa keamanan terisi. Orang baru bisa menjalankan proyek dari README |

Alat kelola (tahap 5 sampai 7) dikerjakan sebelum direktori (tahap 9) supaya inti demo aman lebih dulu.

## 17. Syarat pengumpulan lomba

Diringkas dari guidebook GTNIC 2026. Tanggalnya tetap berlaku walau rencana kerja tidak bertanggal.

| Hal | Ketentuan |
|---|---|
| Pendaftaran | Paling lambat 16 Oktober 2026 |
| Masa pengerjaan | 19 Oktober sampai 10 November 2026 |
| Batas pengumpulan | 10 November 2026 pukul 23.59 WIB lewat Google Form resmi |
| Tautan website live | Di-hosting di platform publik, bisa diakses juri sampai kompetisi berakhir |
| Repositori kode | Akses publik |
| Proposal | PDF, maksimal 15 halaman, A4, Times New Roman 12 pt, jarak baris 1,5, margin kiri 4 cm dan lainnya 3 cm. Nama berkas `GTNIC2026_WebDev_NamaTim.pdf`. Logo GTNIC 2026 di sampul |
| Video demo | 3 sampai 5 menit, di YouTube atau Google Drive dengan akses publik |
| Surat pernyataan keaslian karya | Wajib |

Isi wajib proposal: judul dan nama tim; latar belakang masalah; tujuan, manfaat, dan solusi; target pengguna; arsitektur, diagram basis data, dan teknologi; fitur utama dan keunikan dibanding sistem serupa; batasan sistem; tangkapan layar; rencana pengembangan lanjutan.

Dua aturan yang memengaruhi cara kerja:

1. Karya harus orisinal, buatan sendiri, dan belum pernah dipublikasikan atau memenangkan perlombaan lain.
2. Framework, library, dan aset pihak ketiga boleh dipakai selama sesuai lisensinya dan didokumentasikan dalam proposal.

## 18. Pertanyaan terbuka

1. Apakah tiga anggota tim harus dari perguruan tinggi yang sama? Tata cara pendaftaran menulis wajib sama, syarat dan ketentuan menulis boleh berbeda.
2. Bolehkah mulai menulis kode sebelum 19 Oktober?
3. Berapa biaya pendaftaran? Nominalnya kosong di guidebook.
4. Apakah nama "Trustip" sudah dipakai produk lain, dan apakah domainnya tersedia? Ada layanan transaksi bernama mirip, Trustap.
5. Di wilayah mana data contoh dan uji pengguna dilakukan?

## Lampiran A. Panduan wawancara (15 sampai 20 menit)

Sasaran: minimal 3 penitip dan 3 toko. Tujuannya menguji masalah M7 dan M8 serta tiga asumsi terbesar.

**Asumsi yang diuji**

1. Toko mau mencatat titipan masuk dan sisa di HP bila hanya dua angka per produk.
2. Toko mau memasang kebutuhan produk secara terbuka.
3. Penitip mau mencari toko baru lewat direktori, bukan hanya lewat kenalan.

**Untuk penitip**

1. Menitip ke berapa toko? Bagaimana dulu menemukan toko-toko itu?
2. Pernah ditolak atau bingung dengan syarat toko? Ceritakan.
3. Bagaimana aturan bagi hasilnya disepakati? Pernah berbeda ingatan dengan toko?
4. Bagaimana tahu laku berapa? Seberapa sering harus datang atau menelepon?
5. Pernah selisih hitung atau telat dibayar? Ceritakan kejadian terakhir.
6. Berapa kira-kira barang yang kembali atau kedaluwarsa tiap bulan?
7. Bagaimana menentukan jumlah yang ditaruh di tiap toko?
8. Kalau ada daftar toko terdekat yang menerima titipan, apa yang ingin dilihat sebelum menghubungi?

**Untuk toko**

1. Menerima titipan dari berapa penitip? Bagaimana mereka datang pertama kali?
2. Pernah ingin produk tertentu tetapi tidak tahu mencarinya ke siapa?
3. Apa syarat untuk penitip baru (izin, halal, sampel)?
4. Bagaimana mencatat barang masuk dan sisa sekarang? Berapa menit sehari?
5. Bagaimana tahu harus bayar berapa ke tiap penitip? Pernah keliru?
6. Tahu berapa untung dari barang titipan bulan lalu?
7. Maukah mengisi dua angka per produk di HP setiap tutup? Apa yang membuat enggan?

Catat nama usaha, kota, kutipan langsung, foto buku catatan (dengan izin), dan izin pengutipan di proposal.

## Sumber

- [S1] Pengembangan Modul Pembayaran dan Pengembalian Produk pada Sistem Informasi Berbasis Website untuk Mitra Konsinyasi UMKM Kerupuk Kulit Pakde Waginu, e-Proceeding of Engineering Vol. 13 No. 1, Februari 2026, Telkom University. https://openlibrarypublications.telkomuniversity.ac.id/index.php/engineering/article/download/29404/27756/58065
- [S2] Pengembangan Modul Manajemen Konsinyasi dan Pengiriman Produk untuk UMKM Kerupuk Kulit Sapi Pakde Waginu, Telkom University. https://openlibrarypublications.telkomuniversity.ac.id/index.php/engineering/article/download/29402/27753/58061
- [S3] Perjanjian konsinyasi kue tradisional Aceh, JIM Bidang Hukum Keperdataan Universitas Syiah Kuala, Vol. 5(3), 2021. https://jim.usk.ac.id/perdata/article/download/20256/9333
- [S4] Penerapan model newsvendor di Glaze Bakery, Universitas Katolik Parahyangan. https://repository.unpar.ac.id/handle/123456789/22234
- [S5] Ukirama, Sistem Consignment (Konsinyasi): contoh kartu stok konsinyasi. https://ukirama.com/blogs/konsinyasi-consignment
- [S6] GoodStats, peta persebaran industri mikro kecil 2025 (mengutip BPS). https://data.goodstats.id/statistic/jawa-tengah-memimpin-ini-peta-persebaran-5-raksasa-industri-mikro-kecil-2025-AUaS5
- [S7] GoodStats, jumlah usaha industri mikro dan kecil 2024 (mengutip BPS). https://data.goodstats.id/statistic/5-provinsi-dengan-jumlah-usaha-industri-mikro-dan-kecil-terbanyak-2024-JW1dP
- [S8] RRI, BPOM Sebut Hampir Separuh Industri Mikro Indonesia Bergerak di Sektor Pangan, 31 Juli 2026. https://rri.co.id/batam/regional/2617893/bpom-sebut-hampir-separuh-industri-mikro-indonesia-bergerak-di-sektor-pangan
- [S9] Program titip jual produk UMKM di Mubarokfood. https://wikiexport.ai/ingin-titip-jual-produk-anda-di-mubarokfood-begini-caranya/
- [S10] Titip Dagang di Google Play. https://play.google.com/store/apps/details?id=com.titipdagang.myapp&hl=en_US
- [S11] Titipku (pencatatan konsinyasi) di Google Play. https://play.google.com/store/apps/details?id=dev.isdev.titipku&hl=en_US
- [S12] Kasir Pintar, cara transaksi konsinyasi untuk pihak penerima barang. https://help.kasirpintar.co.id/knowledge-base/bagaimana-cara-transaksi-konsinyasi-untuk-pihak-penerima-barang/
- [S13] Transaksi konsinyasi di Accurate Online. https://penjualanonline.id/cara-membuat-transaksi-konsinyasi-di-accurate-online-update-2026-tips-khusus-tim-cawang/
- [S14] Fitur iPOS 5 Ultimate. https://trigonal.id/blog/artikel/aplikasi-kasir-untuk-minimarket-20-fitur-canggihnya/
- [S15] Jooalan, layanan konsinyasi. https://www.jooalan.id/konsinyasi
- [S16] PaDi UMKM di Google Play. https://play.google.com/store/apps/details?id=com.telkom.padiumkm&hl=en_US
- [S17] Mitra Dagang. https://mitradagang.co.id/
- [S18] Repositori konsinyasi-titip-jual. https://github.com/BenyRonald77/konsinyasi-titip-jual
- [S19] Neon, daftar ekstensi PostgreSQL yang didukung. https://neon.com/docs/extensions/pg-extensions
