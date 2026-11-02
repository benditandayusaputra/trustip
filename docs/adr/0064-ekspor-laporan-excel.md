# 0064: Ekspor laporan ke Excel

Tanggal: 2 November 2026

## Konteks

Tombol Unduh CSV di tab Penjualan menghasilkan teks polos: tanpa judul, tanpa periode, angka Rupiah tanpa format, dan kolom sempit saat dibuka di Excel. Pengguna meminta unduhan berupa Excel saja, bukan CSV, dengan tampilan yang rapi. R-CC-52 meminta pustaka terkecil yang memenuhi kebutuhan, dan pustaka Excel yang umum berukuran ratusan kilobita.

## Keputusan

1. FR-RPT-08 kini menghasilkan berkas `.xlsx` dan tombolnya menjadi Unduh Excel. Ekspor CSV dihapus.
2. Berkas dibuat di peramban tanpa pustaka baru. `lib/domain/zip.ts` membungkus bagian-bagian XML dalam arsip ZIP tanpa kompresi beserta CRC-32, dan `lib/domain/xlsx.ts` menulis satu lembar kerja dengan gaya tetap.
3. Lembar Penjualan memuat judul, periode dan pengelompokan, kepala tabel biru yang dibekukan, isi bergaris tipis, angka sebagai bilangan dengan format Rupiah, ribuan, dan persen, lalu baris Total dari ringkasan laporan. Halaman cetak mendatar dan muat satu lebar kertas.
4. Semua teks ditulis sebagai sel `inlineStr`, sehingga tidak ada teks yang menjadi rumus. Ini menggantikan awalan tanda petik pada R-SEC-55.

## Akibat

Angka di berkas bisa langsung dijumlah dan disaring di Excel. Ukuran berkas sedikit lebih besar dari CSV karena tidak dikompresi, tetapi tetap beberapa kilobita untuk laporan biasa.
