# Rencana Redesain Trustip

Status: disetujui pada 9 Oktober 2026; menjadi acuan implementasi.
Tanggal: 9 Oktober 2026.

## 1. Arah desain yang diusulkan

Trustip menjadi ruang kerja dagang yang rapi, mudah dipakai, dan terasa dekat dengan UMKM. Warna utama biru biasa, `#2563EB`. Halaman publik mengutamakan cerita usaha dan produk; halaman operasional mengutamakan tindakan, angka, dan kejelasan catatan dua pihak.

Keputusan yang perlu direview: biru, palet light/dark, susunan navigasi, kepadatan halaman operasional, serta penampilan beranda dan direktori. Implementasi telah diizinkan setelah review.

Dasar produk: PRD bagian 4, 5, 8, dan 12; spesifikasi bagian 1, 4, 10 sampai 19, 23, dan 24. Panduan lomba diperiksa pada halaman 4, 7 sampai 10, dan 12 sampai 14.

## 2. Referensi dan penerapannya

| Referensi resmi | Pola yang diambil | Adaptasi untuk Trustip |
|---|---|---|
| [Linear: A calmer interface for a product in motion](https://linear.app/now/behind-the-latest-design-refresh) | Hierarki perhatian, navigasi yang tenang, posisi tindakan konsisten, garis pemisah halus | Sidebar tidak mengalahkan isi; tindakan dan jumlah uang mudah ditemukan |
| [Linear: Preferences](https://linear.app/docs/account-preferences) | Pilihan light, dark, dan mengikuti sistem | Tema berlaku pada seluruh halaman dan disimpan di perangkat |
| [Shopify Polaris: Table](https://shopify.dev/docs/api/app-home/latest/web-components/layout-and-structure/table) | Filter di atas data dan tabel yang menjadi daftar pada layar kecil | Riwayat titipan, stok, setoran, laporan, dan katalog |
| [Faire](https://www.faire.com/) | Penemuan produk dan usaha dengan foto, kategori, serta informasi mitra | Direktori dua arah dan profil usaha; alur tetap titip jual sesuai spesifikasi |

Referensi adalah inspirasi pola, bukan template yang disalin. Tidak ada kode, logo, atau aset dari referensi yang dipakai. Aset pihak ketiga yang kelak dipakai harus memiliki lisensi yang sesuai dan dicatat pada `docs/THIRD_PARTY.md` serta proposal. Foto usaha/produk sebaiknya memakai aset tim atau pemilik usaha.

## 3. Palet light dan dark

| Peran | Light | Dark |
|---|---|---|
| Biru utama, tombol utama | `#2563EB` | `#2563EB` |
| Hover tombol utama | `#1D4ED8` | `#1D4ED8` |
| Teks di tombol biru | `#FFFFFF` | `#FFFFFF` |
| Tautan dan fokus | `#1D4ED8` | `#93C5FD` |
| Latar halaman | `#F6F8FC` | `#0B1220` |
| Kartu dan dialog | `#FFFFFF` | `#131D2E` |
| Permukaan sekunder | `#EEF2F8` | `#1B293E` |
| Garis dekoratif | `#DCE3EF` | `#2D3C53` |
| Garis kontrol | `#7B879B` | `#71819A` |
| Teks utama | `#152238` | `#F3F6FC` |
| Teks sekunder | `#52627A` | `#A5B4CA` |
| Permukaan informasi | `#EFF6FF` | `#172D50` |
| Sukses, teks | `#166534` | `#86EFAC` |
| Menunggu, teks | `#92400E` | `#FCD34D` |
| Bahaya, teks | `#B91C1C` | `#FCA5A5` |

Biru dipakai pada tindakan utama, tautan, navigasi aktif, dan data utama grafik. Angka biasa memakai warna teks utama. Hijau hanya untuk status berhasil; kuning untuk menunggu; merah untuk masalah. Status selalu menyertakan kata.

Tema dark memakai navy netral dan permukaan bertingkat. Tema light menggunakan putih dan abu terang dengan ruang yang cukup. Identitas biru tetap sama pada keduanya; tidak ada gradasi cyan-hijau.

Semua nilai menjadi token semantik pada implementasi. Setiap pasangan teks/latar ditargetkan memenuhi WCAG AA; batas isian dan fokus minimal 3:1. Palet usulan perlu diuji pada setiap keadaan aktual, bukan hanya dua warna utama.

## 4. Tipografi, komponen, dan tata letak

- Pertahankan Inter dan JetBrains Mono yang sudah ditetapkan proyek; font disajikan dari server sendiri. Angka uang dan jumlah barang memakai angka selebar tetap.
- Beranda: judul besar sekitar 40 sampai 48 px di desktop dan 30 sampai 34 px di HP. Halaman aplikasi: judul 24 sampai 28 px. Isi dan seluruh isian di HP 16 px.
- Sudut kartu sekitar 16 px, kontrol 10 sampai 12 px. Garis halus dan bayangan hanya sangat ringan pada tema light bila membantu pemisahan.
- Tombol utama tinggi 48 px. Ikon satu keluarga Lucide dengan label yang jelas. Satu tindakan utama dalam setiap konteks.
- Halaman operasional lebar maksimal 1200 px; formulir satu kolom sekitar 640 sampai 720 px. Jarak konsisten mengikuti token kelipatan 4/8 px.
- Desktop memakai sidebar dengan kelompok kerja. Tablet memakai menu ringkas. HP memakai lima navigasi bawah sesuai peran di spesifikasi bagian 23.
- Penitip: Beranda, Cari, Mitra, Setoran, Lainnya. Toko: Beranda, Hitung, Titipan, Cari, Lainnya. Isi Lainnya tetap memuat semua modul yang relevan dengan hak akses.
- Ringkasan dan CTA formulir tetap mudah dicapai di HP, berada di atas navigasi bawah dan safe area. Tidak menutup isian, pesan validasi, atau keyboard.
- Daftar transaksi menjadi kartu berlabel pada layar kecil. Nama, status, tanggal, dan nilai tetap terlihat; rincian sekunder dibuka melalui detail.

## 5. Desain setiap halaman

Daftar di bawah mencakup seluruh rute UI yang tersedia pada saat review. Setiap baris berlaku pada light dan dark, serta HP, tablet, dan desktop.

| Halaman/rute | Rencana susunan dan tindakan |
|---|---|
| `/` | Hero dengan nilai utama titip jual, CTA Mulai, Coba demo, contoh catatan dua pihak; masalah UMKM, empat langkah, manfaat per peran, dan CTA penutup. Tidak memakai klaim jumlah pengguna atau testimoni yang belum terbukti |
| `/masuk` | Formulir fokus satu kolom; akses demo dua peran jelas; verifikasi keamanan, galat, dan status proses tidak menggeser layar |
| `/daftar` | Formulir ringkas, bantuan kata sandi dekat isian, persetujuan kebijakan, lanjut onboarding |
| `/onboarding` | Tiga langkah jelas: tipe usaha, profil, lokasi; ringkasan kemajuan dan bantuan lokasi yang aman |
| `/cari/toko` | Judul, pencarian, lokasi asal, radius, jenis/kategori, urutan; kartu berisi jarak, syarat, komisi, reputasi, skor dan alasannya, kebutuhan terbuka, serta aksi ajuan |
| `/cari/penitip` | Tab Produk/Penitip, pencarian dan filter; kartu foto, kategori, rentang harga jual saran, legalitas yang dinyatakan, jangkauan, dan informasi kecocokan |
| `/kebutuhan` | Papan kebutuhan dengan kategori, lokasi, dan masa tayang; kartu jumlah/frekuensi dan syarat, tanpa gaya percakapan sosial |
| `/kebutuhan/[id]` | Rincian kebutuhan, profil toko, kategori, jumlah/frekuensi, syarat, batas tayang, dan tanggapi melalui Ajuan |
| `/t/[slug]` | Foto usaha, jenis toko, alamat, jam, kategori diterima, syarat, reputasi, kebutuhan terbuka, peta dan atribusi; CTA kerja sama |
| `/u/[slug]` | Identitas produsen, kecamatan/kota, cerita usaha, legalitas yang dinyatakan, jumlah toko mitra, katalog; alamat lengkap tetap privat |
| `/p/[slug]/[sku]` | Foto utama, nama, harga jual saran, komposisi/deskripsi, pembuat dan legalitas; daftar Tersedia di dengan nama/alamat toko dan petunjuk arah; tanpa jumlah stok publik |
| `/undangan/[kode]` | Nama pengundang, tujuan undangan, ringkasan langkah, status undangan, dan tindakan sesuai tipe organisasi |
| `/verifikasi/[kode]` | Status segel dominan, nomor/jenis dokumen, nama pihak, peristiwa dan waktu; penjelasan sederhana. Tidak memuat uang, jumlah barang, atau syarat dagang |
| `/privasi` | Dokumen baca dengan daftar isi dan lebar teks nyaman, hierarki judul, tautan relevan |
| `/ketentuan` | Pola dokumen baca yang sama dengan privasi |
| `/luring` | Penjelasan koneksi, data terakhir, aksi kembali/coba lagi, serta batas operasi luring yang tersedia |
| `+error.svelte` | Status kesalahan, kalimat yang bisa dipahami, kembali ke halaman aman, coba lagi bila relevan, kode permintaan untuk galat tak terduga |
| `/app` | Pengarahan ke dasbor sesuai peran; keadaan perpindahan ringan dan konsisten |
| `/app/penitip` | Empat angka sesuai spesifikasi: unit di toko, terjual 30 hari, hak penitip 30 hari, setoran belum lunas; perlu tindakan di awal, grafik, peringkat, saran dan peluang |
| `/app/toko` | Lima angka sesuai spesifikasi; Hitung cepat menonjol, perlu tindakan, produk terlaris dan peluang. Staf hanya melihat informasi yang diizinkan |
| `/app/ajuan` | Tab masuk/keluar, filter status, giliran tindakan, pihak lawan dan ringkasan syarat |
| `/app/ajuan/baru` | Pilih sasaran/produk, isi syarat, contoh hitungan, ringkasan sebelum kirim |
| `/app/ajuan/[id]` | Status dan giliran jelas, revisi berurutan, baris berubah diberi label, ringkasan syarat, terima/tawar/tolak sesuai peran |
| `/app/mitra` | Daftar mitra, ringkasan stok dan kewajiban yang diizinkan, status, akses ke detail |
| `/app/mitra/[id]` | Kepala mitra dan tab lokal; stok, catatan, setoran dan kesepakatan mengikuti fitur yang tersedia |
| `/app/mitra/[id]/kesepakatan` | Versi aktif, syarat umum/per produk, tiga angka, persetujuan dua pihak, segel dan riwayat |
| `/app/mitra/[id]/kesepakatan/[versi]` | Versi historis jelas bertanggal dan dibedakan dari versi aktif; perubahan syarat mudah dibandingkan |
| `/app/mitra/[id]/perubahan` | Formulir perubahan syarat, nilai lama/baru berlabel, contoh hitungan dan ringkasan sebelum ajuan |
| `/app/mitra/[id]/stok/[produk]` | Kartu stok per produk/batch; mutasi tertanggal, saldo, kedaluwarsa, tautan dokumen sumber |
| `/app/stok` | Ringkasan stok, filter mitra/produk, tanda kedaluwarsa/menipis berlabel, batch dan aksi relevan |
| `/app/titipan` | Riwayat dengan tanggal, mitra, status konfirmasi, jumlah, penanda dicatat luring |
| `/app/titipan/baru` | Pilih mitra, baris produk dan jumlah dengan tombol besar, saran jumlah bila tersedia, kedaluwarsa, ringkasan simpan |
| `/app/titipan/[id]` | Nomor/status, produk dan batch, jejak konfirmasi, segel, waktu perangkat/server bila luring; aksi sesuai hak dan jendela waktu |
| `/app/hitung` | Riwayat hitungan dengan mitra, tanggal, status dan ringkasan nilai |
| `/app/hitung/baru` | Pilih mitra dan waktu hitung, stok tercatat, input sisa/terjual, retur dan alasan, tombol cepat; rekap langsung berurutan omzet, untung toko, harus setor |
| `/app/hitung/[id]` | Nomor/status, rincian hitungan dan rekap, konfirmasi penitip, segel, lanjut catat bayar/titipan baru sesuai jadwal |
| `/app/setoran` | Total dan filter status/umur, pihak lawan, jatuh tempo, sisa setoran, tindakan paling relevan |
| `/app/setoran/[id]` | Nilai utama, periode, rincian sumber hitungan, status, bukti bayar dan riwayat; Catat bayar untuk toko dan Uang diterima untuk penitip |
| `/app/laporan` | Rentang tanggal, kelompok/per mitra/produk, grafik dengan alternatif tabel, ringkasan omzet/untung/hak, ekspor Excel; hanya untuk peran berhak |
| `/app/produk` | Katalog dengan foto, SKU, status, kategori, dan pencarian; akses tambah dan detail |
| `/app/produk/baru` | Formulir berkelompok: identitas/foto, harga jual saran, masa simpan, komposisi/legalitas; validasi dekat isian |
| `/app/produk/[id]` | Detail/edit sesuai fitur yang ada, QR dan tautan publik, sedang mencari toko, status arsip |
| `/app/kebutuhan` | Kebutuhan toko sendiri, status/masa tayang, tanggapan baru, tombol pasang kebutuhan |
| `/app/kebutuhan/baru` | Formulir kategori, jumlah/frekuensi, syarat dan masa tayang, ringkasan publik sebelum simpan |
| `/app/kebutuhan/[id]` | Rincian dan tanggapan, status, perpanjang/terpenuhi/tutup sesuai aturan |
| `/app/pemberitahuan` | Daftar aktivitas berkelompok waktu, tanda belum dibaca, tautan dokumen, tandai dibaca |
| `/app/antrean` | Jumlah menunggu, status pengiriman per catatan, usia catatan, Kirim sekarang, penjelasan hambatan |
| `/app/antrean/[id]` | Catatan pengguna dibanding keadaan terbaru dengan label, galat spesifik, perbaiki/kirim ulang/buang dengan akibat yang jelas |
| `/app/pengaturan` | Bagian profil, lokasi, tampil di direktori, info pembayaran sesuai peran, dan tema; kontrol simpan dekat bagian terkait |

`/api/[...path]` adalah proxy API dan tidak memiliki UI. Peta direktori/favorit, admin moderasi, staf baru, PDF unduhan, chat ajuan, dan fitur P1/P2 lain tetap mengikuti gerbang prioritas. `/app/laporan/harian` tercantum pada peta spesifikasi tetapi belum berupa rute tersendiri; desain rekap mengikuti Laporan yang tersedia, tanpa membuat fitur baru pada tahap visual.

## 6. Tujuh contoh visual untuk review

Pratinjau terpisah memperlihatkan Beranda, Dasbor Toko, Dasbor Penitip, Cari Mitra, Hitung Sisa, Detail Setoran, dan Produk Publik. Tema dapat diganti. Data di dalamnya adalah contoh ilustratif untuk menilai desain, bukan data live atau bukti hasil implementasi. Tempat foto diberi label dan tidak mengimpor aset referensi.

Navigasi pratinjau hanya mengganti contoh tampilan. Formulir Hitung Sisa memperlihatkan rekap lokal dari contoh spesifikasi: stok 20, sisa 2, harga jual Rp2.000, harga setor Rp1.750 menghasilkan 18 terjual, omzet Rp36.000, untung toko Rp4.500, harus setor Rp31.500. Aksi simpan/bayar di pratinjau tidak menulis data.

## 7. Tema, feedback, dan aksesibilitas

Pilihan awal mengikuti sistem, dengan Terang/Gelap/Sistem yang tersimpan di perangkat. Muat pilihan sebelum tampilan pertama untuk mencegah kilatan tema. Implementasi mengikuti CSP yang berlaku; jangan melonggarkan keamanan skrip. Semua komponen, dialog, toast, grafik, peta, captcha, dan halaman galat memakai token sesuai tema. Peta memakai tampilan terang dalam light dan penyesuaian gelap hanya dalam dark, dengan atribusi tetap terbaca.

Semua fitur tetap mencakup memuat, kosong, berhasil, validasi gagal, jaringan gagal, dan luring. Isian tidak hilang ketika gagal. Status proses mencegah klik ganda. Fokus keyboard terlihat; urutan fokus mengikuti dokumen; status tidak mengandalkan warna. Gerak 150 sampai 200 ms, tanpa animasi angka uang, dan mengikuti reduced motion.

Untuk catatan luring, tampilkan jumlah yang belum terkirim dan peringatan usia lebih dari 24 jam. Detail konflik membedakan catatan pengguna dari keadaan server. Tidak mengesankan catatan sudah diterima pihak lain sebelum server menerima.

## 8. Hubungan dengan ketentuan lomba

| Kriteria | Bobot | Keputusan desain dan bukti setelah implementasi |
|---|---:|---|
| Relevansi dan inovasi | 20% | Beranda dan alur inti memperlihatkan masalah titip jual serta catatan bersama dua pihak |
| Fungsionalitas | 25% | Pencarian, ajuan, kesepakatan, titipan, hitung, bayar dan konfirmasi tetap berfungsi; uji kedua peran dan seluruh keadaan |
| Teknis dan arsitektur | 20% | Pertahankan SvelteKit/Go, keamanan, token, pemuatan malas, SEO publik, PWA dan antrean; ukur performa setelah perubahan |
| UX/UI | 15% | Navigasi menurut peran, langkah singkat, angka yang terbaca, dua tema konsisten, responsive dan feedback jelas |
| Dampak dan skalabilitas | 10% | Fokus mengurangi waktu catat, selisih dan keterlambatan; pakai pengukuran pengguna, bukan klaim manfaat tanpa bukti |
| Dokumentasi/demo | 10% | Screenshot akhir, uraian teknologi/lisensi, batasan dan demo 3 sampai 5 menit dengan alur yang dapat diulang |

Guide book tidak mewajibkan warna tertentu atau tema dark-only. Logo lomba diwajibkan pada sampul proposal, bukan seluruh layar aplikasi. Times New Roman 12 pt adalah aturan proposal, bukan aturan tipografi website. Website wajib responsive pada desktop/tablet/HP dan berfungsi pada Chrome, Firefox, Safari, Edge. Karya harus orisinal; aset pihak ketiga mengikuti lisensi dan dicatat dalam proposal. Website dan repositori harus publik sampai kompetisi selesai.

Proposal akhir tetap mengikuti panduan: PDF maksimal 15 halaman A4, Times New Roman 12 pt, spasi 1,5, margin kiri 4 cm dan lainnya 3 cm, cover berlogo GTNIC. Pratinjau review bukan screenshot implementasi yang boleh diklaim sudah berjalan.

## 9. Urutan implementasi setelah review

1. Sinkronkan `DESIGN.md`, ringkasan visual PRD dan spesifikasi, serta catat keputusan tema. Jangan mengubah aturan bisnis.
2. Terapkan token semantik light/dark, pilihan tema, font, kontrol dan feedback; jaga CSP dan perilaku PWA.
3. Terapkan shell publik/aplikasi, navigasi menurut peran dan responsivitas.
4. Redesain pencatatan dan dokumen: Titipan Masuk, Hitung Sisa, stok, Setoran, Kesepakatan serta antrean.
5. Redesain dasbor, laporan, katalog, Ajuan, Kemitraan, Kebutuhan, pengaturan dan pemberitahuan.
6. Redesain beranda, direktori, profil, produk, akun, undangan, verifikasi, kebijakan, luring dan galat.
7. Verifikasi semua rute dan keadaan pada kedua tema; `make check`, `make test`, uji ujung ke ujung relevan, aksesibilitas dan anggaran performa. Ambil screenshot implementasi akhir.

Setiap perubahan dibagi menjadi lingkup yang dapat diperiksa, dengan jalur staging eksplisit dan aturan commit proyek. Perubahan pengguna yang telah ada tidak dimasukkan ke commit pekerjaan ini.

## 10. Kriteria selesai

- Seluruh rute UI yang tersedia dan keadaan komponennya menggunakan tema yang sama tanpa sisa cyan sebagai primary.
- Tidak ada informasi privat tambahan pada halaman publik dan hak akses setiap peran tetap sesuai spesifikasi.
- Tidak ada teks terpotong, isian tertutup CTA, atau overflow tak disengaja pada 360, 768, 1024 dan 1440 px; periksa juga lebar 320 px.
- Nilai uang dan jumlah tetap benar, dengan tiga angka rekap dalam urutan yang ditetapkan.
- Target PRD: pencarian/ajuan di bawah 2 menit, titipan 3 produk di bawah 30 detik, hitung 3 produk di bawah 45 detik. Ukur lewat uji pengguna setelah implementasi.
- Target teknis proyek: Lighthouse HP performa/aksesibilitas minimal 90, LCP di bawah 2,5 detik, CLS di bawah 0,1, serta anggaran bundel tetap terpenuhi.
- Tidak ada klaim seluruh ketentuan lomba telah terpenuhi hanya dari rencana desain; verifikasi fungsi, deployment dan serahan dilakukan setelah implementasi.
