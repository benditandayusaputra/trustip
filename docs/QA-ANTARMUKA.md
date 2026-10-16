# QA antarmuka dan UX

Tanggal: 16 Oktober 2026. Cakupan: seluruh 49 templat halaman SvelteKit, halaman galat 404, dan variasi peran serta parameter yang menghasilkan 78 tampilan halaman.

## Metode

Pemeriksaan menggunakan build produksi lokal, API pengembangan, serta akun demo toko dan penitip. Setiap tampilan diperiksa melalui screenshot desktop terang 1440 px dan HP gelap 390 px. Pemeriksaan lebar konten juga dilakukan pada 320, 768, dan 1024 px. Screenshot ditinjau untuk hierarki informasi, kepadatan kontrol, keterbacaan, konsistensi, dan letak tindakan utama.

Seluruh 156 keadaan terang/gelap menjalani pemeriksaan axe dengan aturan WCAG 2 A, AA, dan 2.1 AA. Hasil akhir: tidak ada pelanggaran yang dideteksi axe, gambar gagal dimuat, galat JavaScript halaman, atau overflow horizontal pada ukuran yang diperiksa. Ini merupakan hasil pemeriksaan otomatis dan visual pada keadaan yang dicatat, bukan sertifikasi aksesibilitas atau jaminan seluruh kemungkinan data.

Onboarding memakai respons sesi uji tanpa keanggotaan, dilanjutkan sampai isian profil dan lokasi tanpa menyimpan usaha. Tinjauan antrean memakai catatan lokal berstatus `NEEDS_REVIEW` yang tidak dikirim ke server. Pemeriksaan interaksi tidak menyimpan profil, produk, dokumen dagang, pembayaran, atau pengakhiran kemitraan.

## Temuan dan revisi

| Temuan | Perbaikan |
| --- | --- |
| Navigasi halaman publik sulit ditemukan di HP | Menu ikon dengan tautan direktori dan kebutuhan; dapat ditutup dengan Escape, klik di luar, atau memilih tujuan. |
| Header terlalu padat pada layar sempit | Kontrol tema dan akun memakai ikon bernama aksesibel; merek publik menyesuaikan lebar. |
| Halaman dalam menu Lainnya tidak menunjukkan lokasi aktif | Penanda halaman aktif dan status terbuka ditambahkan pada navigasi bawah. |
| Detail dan formulir kurang jelas jalan kembalinya | Tautan kembali dan jejak halaman yang dapat diklik; label riwayat Hitung Sisa diperbaiki. |
| Filter panjang mendahului hasil pencarian di HP | Pencarian tetap terlihat, filter lanjutan dapat dibuka, dan hasil memperoleh ruang tersendiri. |
| Formulir produk panjang tanpa kelompok informasi | Tiga bagian dengan judul dan petunjuk; pasangan isian memakai dua kolom pada layar yang sesuai. |
| Deskripsi panjang memakai isian satu baris | Area teks pada deskripsi, komposisi, alamat, catatan, persyaratan, dan informasi pembayaran. |
| QR mendahului pekerjaan ubah produk | Formulir didahulukan; QR tersedia dalam bagian yang dapat dibuka. |
| Metadata dokumen sulit dipindai | Kartu metadata dengan label dan nilai tersusun satu kolom di HP. |
| Pengaturan panjang dan bagian sulit dicari | Navigasi antarbagian tetap terlihat; formulir tetap terpasang agar isian belum disimpan bertahan. |
| Tindakan akhir kemitraan terlalu dekat dengan pekerjaan rutin | Bagian terpisah yang dapat dibuka, dengan konfirmasi akhir tetap berlaku. |
| Captcha membuat halaman akun melebar pada 320 px | Grid satu kolom dan kartu yang dapat menyusut; widget tetap utuh dan verifikasi tetap berjalan. |

Sidebar tetap berwarna primary biru. Seluruh perubahan berada pada presentasi dan navigasi frontend; kontrak API serta aturan bisnis tetap berlaku.

## Pemeriksaan per halaman

Setiap baris telah diperiksa pada desktop terang dan HP gelap, serta ukuran tambahan di atas. Tindakan yang disebut sebagai ditinjau tidak berarti dikirim ke server.

| No. | Rute | Hasil tinjauan dan tindakan |
| --- | --- | --- |
| 1 | `/` | Hierarki hero, pilihan peran, alur kerja, dan tindakan pendaftaran terbaca; menu publik HP diperbaiki. |
| 2 | `/masuk` | Label, tombol sandi, verifikasi, dan tautan daftar diperiksa; overflow 320 px diperbaiki. |
| 3 | `/daftar` | Urutan isian, persetujuan, dan verifikasi diperiksa; overflow 320 px diperbaiki. |
| 4 | `/demo` | Pilihan kedua peran dan isian masuk diperiksa; overflow 320 px diperbaiki. |
| 5 | `/onboarding` | Pemilihan peran, profil, serta lokasi ditinjau tanpa penyimpanan; deskripsi dan alamat memakai area teks. |
| 6 | `/privasi` | Lebar baca, judul bagian, dan daftar isi teks tetap terbaca pada HP. |
| 7 | `/ketentuan` | Hierarki dokumen serta pembungkusan teks diperiksa. |
| 8 | `/luring` | Pesan keadaan tanpa koneksi dan tombol coba lagi mudah ditemukan. |
| 9 | `/cari/toko` | Pencarian, urutan, pemilih daftar/peta, filter, dan kartu ditinjau sebagai pengunjung dan penitip. |
| 10 | `/cari/penitip` | Tampilan pengunjung dan toko diperiksa; pencarian tetap terlihat dan filter lanjutan dapat ditutup. |
| 11 | `/kebutuhan` | Pencarian selalu terlihat; filter dapat ditutup agar kartu kebutuhan lebih cepat dijangkau. |
| 12 | `/t/[slug]` | Profil, kategori, persyaratan, dan tindakan kerja sama terbaca; teks uji panjang tetap dibungkus. |
| 13 | `/u/[slug]` | Profil penitip, katalog, legalitas, dan tindakan kontak diperiksa. |
| 14 | `/p/[slug]/[sku]` | Harga, komposisi, dan ketersediaan diperiksa; ditambahkan tautan kembali ke katalog penitip. |
| 15 | `/kebutuhan/[id]` | Rincian serta tindakan menanggapi dan membuka profil toko mudah ditemukan. |
| 16 | `/undangan/[kode]` | Keadaan pengunjung dengan ajakan masuk/daftar diperiksa; penerimaan undangan tidak dikirim. |
| 17 | `/verifikasi/[kode]` | Kode sah dan kode tidak ditemukan diperiksa; hasil serta batas arti segel jelas. |
| 18 | `/app` | Pengalihan akun toko menuju dasbor yang sesuai diverifikasi. |
| 19 | `/app/toko` | Ringkasan, prioritas, dan hitung cepat diperiksa; header HP diringkas. |
| 20 | `/app/penitip` | Ringkasan, grafik, prioritas, dan pencarian mitra diperiksa. |
| 21 | `/app/produk` | Pencarian, kategori, status, kartu produk, dan tindakan tambah diperiksa. |
| 22 | `/app/produk/baru` | Formulir dikelompokkan; validasi kosong memfokuskan isian pertama dan mempertahankan isian lain. |
| 23 | `/app/produk/[id]` | Formulir didahulukan; QR dibuka melalui disclosure dan isian panjang memakai area teks. |
| 24 | `/app/mitra` | Tab status, daftar mitra, dan undang mitra diperiksa pada kedua peran. |
| 25 | `/app/mitra/[id]` | Ringkasan diberi judul; tindakan titip/hitung toko ditonjolkan dan pengakhiran dipisahkan. |
| 26 | `/app/mitra/[id]/kesepakatan` | Metadata dikelompokkan; segel, produk, dan jejak persetujuan diperiksa. |
| 27 | `/app/mitra/[id]/kesepakatan/[versi]` | Versi historis diperiksa; tautan kembali menuju kesepakatan aktif. |
| 28 | `/app/mitra/[id]/perubahan` | Pemilihan perubahan produk, syarat, catatan, dan jalan kembali diperiksa. |
| 29 | `/app/mitra/[id]/stok/[produk]` | Saldo serta riwayat pergerakan stok diperiksa; tersedia jalan kembali ke mitra. |
| 30 | `/app/ajuan` | Tampilan masuk/keluar dan status diperiksa pada kedua peran. |
| 31 | `/app/ajuan/baru` | Undangan umum serta ajuan dari produk, toko, dan kebutuhan diperiksa; catatan memakai area teks. |
| 32 | `/app/ajuan/[id]` | Metadata dikelompokkan; syarat dan pilihan tanggapan diperiksa tanpa mengirim tanggapan. |
| 33 | `/app/titipan` | Filter dan daftar dokumen diperiksa pada kedua peran. |
| 34 | `/app/titipan/baru` | Pemilih mitra dan formulir produk diperiksa; terdapat jalan kembali. |
| 35 | `/app/titipan/[id]` | Metadata, jumlah, tabel responsif, dan segel diperiksa. |
| 36 | `/app/hitung` | Filter, daftar hasil, dan keadaan peran penitip diperiksa; label jejak halaman diperbaiki. |
| 37 | `/app/hitung/baru` | Pemilih mitra, isian sisa, dan tindakan cepat diperiksa tanpa menyimpan hasil. |
| 38 | `/app/hitung/[id]` | Kartu ringkasan, metadata, hasil per produk, serta segel diperiksa. |
| 39 | `/app/setoran` | Filter status/umur dan daftar setoran diperiksa pada kedua peran. |
| 40 | `/app/setoran/[id]` | Metadata dikelompokkan; ringkasan, pembayaran, dan rincian produk diperiksa. |
| 41 | `/app/stok` | Daftar berisi data toko dan keadaan kosong penitip diperiksa. |
| 42 | `/app/kebutuhan` | Status serta daftar kebutuhan toko diperiksa; Lainnya menunjukkan halaman aktif di HP. |
| 43 | `/app/kebutuhan/baru` | Deskripsi memakai area teks; papan ketik angka disesuaikan dan terdapat jalan kembali. |
| 44 | `/app/kebutuhan/[id]` | Rincian, status, dan tanggapan diperiksa; perubahan status tidak dikirim. |
| 45 | `/app/laporan` | Filter dikelompokkan; tab, ringkasan, dan keadaan tanpa hasil diperiksa pada kedua peran. |
| 46 | `/app/pemberitahuan` | Keterbacaan daftar, penanda belum dibaca, dan tautan dokumen diperiksa. |
| 47 | `/app/antrean` | Keadaan kosong, petunjuk pengiriman, dan tindakan kirim diperiksa. |
| 48 | `/app/antrean/[id]` | Konflik stok dan perbandingan nilai ditinjau melalui catatan lokal; tidak dikirim ulang atau dibuang. |
| 49 | `/app/pengaturan` | Kedua peran diperiksa; navigasi bagian, area teks, pilihan tema, dan ketahanan isian belum disimpan diverifikasi. |
| Tambahan | Halaman tidak ditemukan | Status 404, pesan, kontrol tema, dan tindakan kembali diperiksa. |

## Verifikasi interaksi dan regresi

Delapan skenario browser tambahan lulus: menu publik melalui Escape dan navigasi; pencarian dengan filter tertutup; header HP dan status Lainnya; validasi produk tanpa kehilangan isian; QR produk; perpindahan bagian pengaturan tanpa kehilangan isian; pembukaan dan pembatalan konfirmasi akhir kemitraan; serta pengisian cepat Hitung Sisa tanpa menyimpan.

Regresi browser: 14 pengujian navigasi, akun, beranda, dan tema lulus; 18 pengujian loading halaman lulus. Regresi unit/komponen frontend: 590 pengujian lulus. `make check test` lulus, mencakup format, lint, tipe, token desain, arsitektur, komentar, rahasia, serta suite API dan frontend.

## Bukti visual pilihan

| Keadaan | Screenshot |
| --- | --- |
| Formulir produk desktop | [Produk baru](screenshots/qa/produk-baru.webp) |
| Pencarian HP dengan filter tertutup | [Cari penitip](screenshots/qa/cari-penitip-hp.webp) |
| Pengaturan HP | [Profil usaha](screenshots/qa/pengaturan-hp.webp) |
| Ringkasan toko dengan sidebar biru | [Dasbor toko](screenshots/qa/dasbor-toko.webp) |

## Batas pemeriksaan

Pemeriksaan memakai Chromium pada ukuran viewport yang disebutkan. Safari, Firefox, perangkat fisik, dan pembaca layar belum diperiksa dalam putaran ini. Screenshot akun dapat memperlihatkan keadaan verifikasi captcha yang sedang berjalan; uji regresi ukuran menunggu widget benar-benar terpasang. Foto produk demo tidak tersedia pada sebagian data sehingga placeholder ikon dipertahankan. Konten uji lama yang menyerupai markup ditampilkan sebagai teks; data tersebut tidak diubah melalui QA ini.

Pembatasan laju API sempat aktif saat pemindaian berulang. Halaman yang terdampak diulang sesudah jendela pembatasan pulih, dan hasil akhir tidak mengandung respons 429. Respons 401 pada percobaan sesi pengunjung tanpa login merupakan keadaan yang diharapkan.
