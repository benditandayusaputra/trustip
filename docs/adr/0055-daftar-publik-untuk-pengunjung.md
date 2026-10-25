# 0055: Daftar publik untuk pengunjung

Tanggal: 25 Oktober 2026

## Konteks

Pengunjung yang belum masuk membuka Cari toko, Cari penitip, dan Papan Kebutuhan, tetapi pesan "Lokasi usaha belum diatur" tidak berlaku bagi mereka karena mereka tidak punya usaha. Kartu hasil berbeda tinggi, isi panjang membuat kartu melar, dan hanya judul yang bisa diklik. Filter ikut tergulir sehingga harus kembali ke atas untuk mengubahnya. Data contoh belum berfoto sehingga daftar terlihat kosong.

## Keputusan

1. Saat pengunjung membuka daftar tanpa titik asal, aplikasi meminta lokasi perangkat sekali dan mengurutkan dari lokasi itu. Bila izin ditolak, isi tiap halaman hasil diacak di peramban. Urutan Reputasi dan Terbaru tidak diacak.
2. Jarak untuk pengunjung selalu "Semua jarak" dan pilihannya dinonaktifkan. Anggota tetap memilih jarak seperti sebelumnya.
3. Pada pemilih titik asal, pilihan "Lokasi usaha saya" ditampilkan sebagai "Tanpa titik asal (acak)" bagi pengunjung.
4. Pencarian dan tombol Filter berada di atas daftar pada ketiga halaman dan menempel di atas layar saat digulir. Urutan pada Cari toko pindah ke panel Filter supaya bilah yang menempel tetap pendek di HP.
5. Kartu dalam satu daftar sama tinggi. Nama dan baris keterangan dipotong dengan elipsis, deskripsi dibatasi dua baris, dan kategori ditampilkan dua lalu "+N lagi". Seluruh kartu membuka halaman detail. Tombol aksi tetap di dasar kartu, dan tautan Laporkan hanya ada di halaman detail.
6. Data contoh memakai foto berlisensi terbuka dari Wikimedia Commons untuk logo, foto toko, dan foto produk. Sumber dan lisensi tiap foto dicatat di `docs/KREDIT-FOTO.md`. Setel ulang demo menghapus berkas foto demo lama dari disk sebelum mengunggah yang baru.

## Akibat

Urutan acak berlaku per halaman hasil, bukan seluruh daftar, karena halaman berikutnya tetap diambil dengan kursor urutan kota. Tidak ada perubahan kontrak API. Pengunjung tetap harus masuk untuk Ajukan kerja sama, Tanggapi, dan melihat nomor WhatsApp.
