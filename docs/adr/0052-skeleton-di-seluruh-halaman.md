# ADR 0052: skeleton di seluruh halaman

Tanggal: 12 Oktober 2026. Status: diterima.

## Konteks

Pengguna meminta skeleton loading di seluruh halaman. Sebagian besar pemuatan data masih berupa satu garis, sedangkan perpindahan ke halaman publik menunggu data server tanpa bentuk isi tujuan.

## Keputusan

Komponen bersama menyediakan pola daftar, kartu, detail, formulir, laporan, profil, produk publik, artikel, beranda, serta akun. Pola dasbor yang sudah ada dipindahkan ke komponen bersama. Penjaga akun memilih skeleton berdasarkan rute; setiap fitur memakai pola sesuai isi yang sedang dimuat.

Cangkang publik, akun, dan aplikasi menampilkan skeleton tujuan saat navigasi antarhalaman masih berlangsung. Isi sebelumnya disembunyikan dan tidak dapat difokuskan, tetapi tetap terpasang agar navigasi yang dibatalkan mempertahankan isian. Query filter tetap ditangani oleh pemuatan data fitur. Halaman siap tidak ditunda untuk menampilkan animasi.

Peta dan QR memakai placeholder dengan ukuran konten. Placeholder tidak berisi kontrol aktif atau nilai palsu, memiliki status aksesibel, memakai token tema, dan menghormati preferensi pengurangan gerak.

## Konsekuensi

Tidak ada perubahan kontrak API, izin, basis data, atau dependensi. Render server dan metadata halaman publik dipertahankan. Verifikasi mencakup cakupan rute, pemuatan akun dan data, navigasi server yang ditunda, pembatalan navigasi, kedua tema, layar kecil, serta aksesibilitas.
