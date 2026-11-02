# 0062: Sembunyikan halaman akun demo

Tanggal: 2 November 2026

## Konteks

Halaman `/demo` menampilkan tiga akun contoh beserta kata sandinya kepada setiap pengunjung, dan beranda, Panduan, FAQ, serta footer mengarah ke sana. Pengguna meminta halaman itu disembunyikan. Pengguna juga meminta daftar akun yang ada, supaya setiap toko dan produk di Cari Mitra bisa langsung dicoba saat demo. Ternyata semua 28 organisasi di direktori adalah organisasi demo yang punya akun dengan kata sandi yang sama, tetapi hanya tiga yang tercatat.

## Keputusan

1. Rute `/demo`, komponen daftar akun, dan tautan ke sana dari beranda, Panduan, FAQ, dan footer dihapus. Membuka `/demo` kini menampilkan halaman tidak ditemukan.
2. Akun dibagikan lewat `docs/AKUN-DEMO.md`, yang memuat ke-28 akun beserta keterangan mana yang punya riwayat transaksi. Masuk tetap lewat `/masuk` dengan captcha yang sama.
3. Direktori tidak disaring. Semua isinya sudah akun sungguhan, sehingga tidak ada yang perlu disembunyikan.

## Akibat

FR-AUTH-07 dan FR-PUB-01 diperbarui. Data demo dan setel ulang malam tidak berubah.
