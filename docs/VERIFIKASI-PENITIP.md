# Verifikasi tata letak dasbor penitip

Tanggal: 10 Oktober 2026.

## Cakupan

`/app/penitip` memakai grafik dan Perlu tindakan pada baris pertama desktop, Saran selebar isi dengan tiga kartu awal, dua peringkat, dan Peluang selebar isi. Peringkat dan Peluang menampilkan lima baris awal dengan tombol buka/ringkas untuk sisanya. Pada HP, Perlu tindakan mendahului grafik. Filter toko/produk berada di dalam kartu grafik. Dasbor staf memakai ruang penuh untuk bagian yang boleh diakses.

## Pemeriksaan

- `make check`: lulus format, lint, tipe, batas kode, arsitektur, token desain, pemeriksaan rahasia dan keamanan dependensi.
- `make test`: seluruh paket Go dengan deteksi race lulus; 83 berkas dan 501 uji frontend lulus.
- `make e2e`: 67 uji lulus, termasuk alur transaksi, keamanan, Lighthouse, dan audit aksesibilitas seluruh halaman pada kedua tema.
- Uji regresi memastikan tiga Saran awal, buka seluruh lima Saran, tautan aksi tetap benar, ringkas kembali, dan audit aksesibilitas komponen.
- Uji filter memastikan grafik menerima toko yang dipilih dan fokus tetap berada pada kontrol pilihan. Preview produksi juga memeriksa muat ulang laporan sesudah pilihan berubah serta perpindahan grafik/tabel.

## Tampilan produksi

30 tangkapan layar mencakup tampilan awal, Saran terbuka, dan tabel penjualan pada lebar 320, 390, 768, 1024, dan 1440 px, masing-masing dalam tema terang dan gelap. Tidak ditemukan overflow horizontal halaman atau galat JavaScript. Tangkapan layar dan catatan pemeriksaan berada di [galeri penitip](../../output/consignor-layout/index.html) dan [hasil pemeriksaan](../../output/consignor-layout/verification.json).

Tidak ada perubahan kontrak API, perhitungan keuangan, hak akses, atau dependensi. Uji otomatis aksesibilitas tidak menggantikan evaluasi langsung dengan pengguna pembaca layar.
