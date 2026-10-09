# 0041. Halaman publik dan akses pengunjung

## Konteks

Bagian 17 spesifikasi meminta beranda, profil toko dan penitip, halaman produk dengan "Tersedia di", halaman Kebutuhan yang bisa dibagikan, serta direktori dan Papan Kebutuhan yang bisa dibuka tanpa masuk. Daftar putih data publik membatasi kolom yang boleh keluar. R-APP-02 meminta halaman publik dirender server, R-PERF-08 meminta cache tepi singkat, dan ADR 0032 menunda akses pengunjung direktori sampai tahap ini. Spesifikasi tidak menetapkan bentuk rute API publik, sumber gambar pratinjau, dan cara direktori melayani pengunjung yang tidak punya organisasi.

## Keputusan

1. Rute API publik berawalan `/api/v1/public/` (`stores/{slug}`, `consignors/{slug}`, `products/{slug}/{sku}`, `needs/{id}`) hidup di modul `directory`, supaya aturan reputasi "Baru" dan pemetaan syarat toko tidak ditulis dua kali. Setiap jawaban punya struktur jawaban publik sendiri dan skemanya di OpenAPI memakai `additionalProperties: false`, sehingga uji kontrak menolak kolom di luar daftar putih.
2. ID organisasi dan ID produk ikut di jawaban publik karena tombol Laporkan dan Ajukan kerja sama membutuhkannya. ID tidak membuka data lain: semua rute ber-ID tetap memeriksa lingkup.
3. Tautan ke profil toko atau penitip hanya diberikan bila organisasi itu tampil di Cari Mitra. Halaman produk tetap terbuka lewat QR selama produk tidak diarsipkan (BR-PUB-02).
4. "Tersedia di" memuat paling banyak 20 toko yang kemitraannya aktif, sakelar tampilnya menyala, dan sisa stok produk itu di atas 0. Urutan terdekat dihitung server dari `lat` dan `lng` yang dikirim app setelah pengunjung mengizinkan lokasi. Tanpa lokasi, toko diurutkan menurut nama.
5. Profil penitip menampilkan paling banyak 20 produk aktif dan lokasi publik yang sudah dibulatkan.
6. Halaman `/`, `/t/`, `/u/`, `/p/`, dan `/kebutuhan/{id}` dirender server lewat `event.fetch` ke proxy dan memasang `Cache-Control: public, s-maxage=60, stale-while-revalidate=300`. Jawaban API sendiri tetap `no-store`.
7. Rute daftar direktori dan Papan Kebutuhan memakai autentikasi opsional: tanpa cookie, permintaan berjalan sebagai pengunjung; dengan cookie yang tidak sah, tetap 401 supaya app menyegarkan sesi. Pengguna yang sudah masuk tetapi belum punya organisasi tetap ditolak. Pengunjung mendapat penonton tanpa organisasi (ID nol), titik asal kota atau lokasi sekarang, dan tanpa skor kecocokan.
8. Halaman direktori dan Papan Kebutuhan tetap cangkang yang dirender klien. Pengunjung melihat tombol Masuk di kepala halaman, dan tombol yang butuh sesi mengarah ke halaman masuk lalu kembali.
9. Gambar pratinjau tautan memakai foto produk atau logo organisasi bila ada, selain itu ikon aplikasi. Tidak ada pembangkit gambar di server.

## Akibat

- Tombol "Coba akun demo" di beranda menyusul bersama akun demo (12.1).
- Gambar pratinjau yang dibuat khusus per halaman dicatat sebagai usulan, bukan dikerjakan.
