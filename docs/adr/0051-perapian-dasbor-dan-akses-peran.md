# ADR 0051: perapian dasbor dan akses peran

Tanggal: 11 Oktober 2026. Status: diterima.

## Konteks

Pengguna meminta skeleton yang menyerupai isi, ikon angka kunci, tinggi kartu yang seragam, Peluang di atas Produk terlaris selebar isi, foto dan tautan detail produk, serta perbaikan pergantian akun toko ke penitip.

## Keputusan

Kedua dasbor memakai skeleton untuk ringkasan dan bagian-bagian halaman sejak pemuatan akun. Angka laporan mempertahankan tempatnya selama pemuatan dan menampilkan kegagalan bila laporan tidak tersedia. Ikon memakai pustaka yang sudah tersedia dan warna token tema. Grid merentangkan Hitung cepat dan Perlu tindakan serta menyamakan tinggi kartu Saran. Peluang mendahului peringkat; tiap bagian memakai lebar penuh.

Laporan produk menyertakan metadata foto dan alamat produk lewat gabungan pada kueri laporan yang tetap dibatasi ke organisasi pelaku. Tidak ada permintaan per produk dan tidak ada perluasan izin laporan. Foto yang belum tersedia memakai ikon; detail menuju halaman produk publik. Metadata tambahan tidak memuat harga setor atau alamat privat.

Tujuan setelah login dihitung dari tipe organisasi pada jawaban login. Parameter lanjut yang mengarah ke dasbor tipe lain diganti dengan dasbor akun baru. Penjaga area anggota menyembunyikan isi dan mengalihkan akses langsung ke dasbor yang sesuai. Keluar menghapus akun dari memori selain simpanan baca perangkat. Otorisasi API tetap diperiksa server.

## Konsekuensi

Kontrak laporan mendapat metadata produk opsional tanpa perubahan hitungan. Tidak ada dependensi atau migrasi baru. Pemeriksaan mencakup perpindahan akun, akses kedua arah, tautan dan foto produk, pemuatan serta kegagalan laporan, tata letak kedua tema dan ukuran layar, serta aksesibilitas.
