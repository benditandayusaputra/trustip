# 0057: Akun demo Bendi, Zara, dan Jessica

Tanggal: 25 Oktober 2026

## Konteks

Video lomba memperlihatkan alur dua pihak memakai akun demo. Pengguna meminta nama akun contoh memakai Bendi, Zara, atau Jessica. Pada data lama, toko demo selalu mengembalikan seluruh sisa roti tiap malam sehingga stok roti nol saat demo dibuka, dan penitip demo belum mengisi info rekening, sehingga layar Setoran menampilkan pesan kosong.

## Keputusan

1. Penitip demo menjadi Roti Mas Bendi milik Bendi, toko demo menjadi Warung Mbak Zara milik Zara, dan pemilik Minimarket Berkah menjadi Jessica. Email `penitip@demo.trustip.id` dan `toko@demo.trustip.id` tetap. Jessica masuk dengan `minimarket@demo.trustip.id`. Kata sandi semua akun tetap `Coba-Demo-2026`.
2. Halaman `/demo` menampilkan tiga akun itu. Ajuan terbuka dari Minimarket Berkah ke Roti Mas Bendi dipakai untuk memperagakan penerimaan Ajuan.
3. Kemitraan Roti Mas Bendi dengan Warung Mbak Zara tidak dihitung pada hari terakhir riwayat, sehingga titipan kemarin masih ada di rak dan siap dipakai untuk memperagakan Hitung Sisa dan Setoran.
4. Penitip demo mengisi info pembayaran berlabel data demo lewat kasus pakai yang sama dengan pengaturan.

## Akibat

Jumlah organisasi, kemitraan, Kebutuhan, dan Ajuan tidak berubah (ADR 0044). Slug profil penitip demo menjadi `roti-mas-bendi`. Perubahan berlaku pada setel ulang demo berikutnya.
