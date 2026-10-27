# 0059: Revisi tampilan ruang kerja dan halaman publik

Tanggal: 27 Oktober 2026

## Konteks

Tinjauan tampilan dari tim (dokumen "Revisi Trusttip") menilai ruang kerja terlalu ramai: judul halaman tampil dua kali, sidebar memuat Pemberitahuan padahal ada lonceng, Setoran dan Laporan terpisah, halaman detail tidak punya tombol kembali yang jelas, dan beberapa istilah tidak konsisten. Halaman publik dinilai terlalu besar dan Panduan terlalu panjang.

## Keputusan

1. Header ruang kerja menjadi satu-satunya tempat judul halaman. `PageHeading` mendaftarkan judulnya ke header lewat konteks; di luar ruang kerja ia tetap tampil di isi halaman. Halaman detail dan halaman di luar menu memakai tombol kembali di header dengan jejak `Induk > Judul`.
2. Sidebar memuat merek dan nama ruang kerja, lalu dua kelompok menu. Pemberitahuan dan Setoran keluar dari menu. Setoran digabung ke tab Setoran di Laporan; rute `/app/setoran` tetap ada untuk tautan Perlu tindakan dan detail. Antrean hanya untuk toko karena penitip tidak mencatat Titipan Masuk atau Hitung Sisa.
3. Menu Akun hanya berisi Kelola profil dan Keluar. Keluar dari semua perangkat pindah ke Pengaturan.
4. Bagian Peluang dihapus dari kedua dasbor. Judul Saran di dasbor menjadi Pengingat; tombol Cari mitra di dasbor penitip menjadi Cari toko.
5. Kartu toko dan penitip di Cari tersusun sebagai baris; kartu produk tiga sampai empat kolom; panel pencarian dan filter diperkecil.
6. Navigasi publik berupa pil Beranda, Cari toko, Kebutuhan toko, Panduan dengan halaman aktif disorot. Profil, produk, dan Kebutuhan publik punya tombol Kembali.
7. Panduan diringkas: daftar isi, skenario demo, catatan, dan kotak demo dihapus; langkah tiap tahap dilipat.

## Akibat

Aturan bisnis dan API tidak berubah. Halaman Masuk/Daftar dan Pengaturan menunggu desain Figma dan belum disentuh. Bagian Peluang pada spesifikasi 15.1 dan 15.2 dihapus.
