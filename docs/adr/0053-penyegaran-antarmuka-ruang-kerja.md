# 0053: Penyegaran antarmuka ruang kerja

Tanggal: 16 Oktober 2026

## Konteks

Pengguna meminta redesain seluruh frontend `app` dengan referensi dashboard berkartu, sidebar biru primary, ruang yang lebih lapang, dan ikon yang relevan. Primary biru yang sudah dipakai tetap dipertahankan. Kode harus bersih, tanpa komentar dan tanpa emoji pada antarmuka.

## Keputusan

Gunakan sidebar biru primary dengan navigasi aktif putih dengan teks biru, kelompok menu dan konteks peran. Header menampilkan ikon modul, jejak halaman dan tautan pencarian mitra yang berfungsi. Pertahankan navigasi bawah pada layar kecil, tema terang/gelap dan pilihan sistem yang tersimpan.

Pusatkan tampilan pada token dan komponen bersama. Gunakan radius kartu 20 px, garis pemisah halus, bayangan tipis, judul berikon, filter dalam panel, tab bersegmen, serta formulir dalam kartu. Katalog produk memakai grid foto asli dengan placeholder ikon jika foto belum tersedia. Kartu stok utama memakai permukaan gelap tanpa menambah angka atau tren rekaan.

Beranda, kerangka akun, direktori, profil, halaman operasional, galat, keadaan kosong, dan luring mengikuti sistem yang sama. Gunakan ikon Lucide yang sudah terpasang; tidak menambah pustaka, aset produk rekaan, maupun perubahan API.

## Akibat

Perubahan komponen bersama menjangkau semua halaman tanpa menggandakan aturan visual. Perilaku bisnis, otorisasi, data publik, validasi dan alur luring mengikuti implementasi yang sudah ada. Pemeriksaan mencakup format, lint, tipe, token, unit/komponen, build, responsivitas, dan aksesibilitas kedua tema.
