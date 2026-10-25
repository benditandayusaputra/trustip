# 0058: Segel bekerja di belakang layar

Tanggal: 25 Oktober 2026

## Konteks

Detail Kesepakatan, Titipan Masuk, Hitung Sisa, dan Setoran menampilkan panel "Segel sah" lengkap dengan jejak segel dan kode verifikasi, dan app punya halaman publik `/verifikasi/<kode>`. Pengguna menilai panel ini istilah teknis yang tidak perlu dilihat penitip dan toko, dan meminta segel tetap bekerja tetapi hanya di belakang layar.

## Keputusan

1. App tidak lagi menampilkan status segel, jejak segel, atau kode verifikasi di detail dokumen mana pun.
2. Halaman `/verifikasi/<kode>` dihapus dari app. Rute API `GET /api/v1/public/seals/<kode>` dan `GET /api/v1/public/seal-keys` tetap ada untuk pemeriksaan oleh tim pengelola dan pemeriksaan mandiri (FR-SIG-04).
3. API tidak berubah. Setiap peristiwa dagang tetap disegel dalam transaksi yang sama (BR-SIG-03), server tetap menghitung ulang `content_hash` saat dokumen dibuka, dan ketidakcocokan dicatat sebagai peristiwa keamanan `seal_invalid` (R-SEC-33). Tugas harian tetap memeriksa keutuhan rantai.
4. Panduan dan FAQ tidak lagi membahas segel. Kebijakan privasi dan ketentuan tetap menyebut segel karena datanya masih diproses.

## Akibat

Penitip dan toko tidak lagi melihat peringatan "Segel tidak sah" bila isi dokumen berubah di basis data; pengawasannya berpindah ke log keamanan. FR-SIG-02, FR-SIG-03, BR-SIG-06, R-SEC-33, dan R-SEC-34 disesuaikan. Panduan menjadi delapan tahap.
