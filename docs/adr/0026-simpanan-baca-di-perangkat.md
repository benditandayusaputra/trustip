# 0026. Simpanan baca di perangkat

## Konteks

FR-PWA-04 meminta daftar dan dokumen yang terakhir dibuka tetap terbaca saat luring dengan tanda "data terakhir". BR-PWA-02 meminta simpanan itu dihapus saat keluar. ADR 0005 menduga simpanan baca akan ditambahkan di service worker. Paket kerja (R-OFF-01) dan antrean (R-OFF-02) harus tinggal di IndexedDB lewat `offline/store.ts` (R-CC-50), dan formulir luring perlu tahu kapan datanya diambil supaya stok bisa digabung dengan antrean (R-OFF-04). Service worker tidak tahu siapa yang sedang masuk.

## Keputusan

1. Simpanan baca tinggal di IndexedDB, bukan di service worker. `apiClient` menyimpan setiap jawaban GET yang berhasil beserta waktunya. Bila GET gagal karena jaringan, `apiClient` menjawab dari salinan itu. Bila tidak ada salinan, galat jaringan diteruskan. Permintaan tulis tidak pernah dijawab dari simpanan (BR-PWA-01).
2. Jawaban info pembayaran dan berkas privat tidak disimpan di perangkat (R-OFF-14, R-SEC-72).
3. Saat jawaban datang dari salinan, keadaan `Freshness` mencatat waktu salinan tertua yang sedang tampil. Pita luring di area `/app` muncul bila perangkat luring atau bila ada data dari salinan, dan menyebut waktunya. Jawaban berikutnya yang sampai ke server menghapus tanda itu.
4. Simpanan baca dihapus saat keluar dari perangkat ini, saat keluar dari semua perangkat, dan saat masuk berhasil, supaya data pengguna sebelumnya tidak terbaca pengguna berikutnya.
5. Service worker tetap tidak menyimpan jawaban API. Ia hanya menyimpan halaman `/app` yang pernah dibuka, per versi app, supaya app bisa dibuka ulang tanpa sinyal. Halaman itu tidak memuat data pengguna karena area `/app` dirender di klien.
6. Paket kerja (R-OFF-01) adalah jawaban Kesepakatan aktif, katalog mitra, dan lembar Hitung Sisa tiap kemitraan aktif yang diambil lebih awal lewat `apiClient`, sehingga tersimpan di simpanan baca bersama waktunya. Paket diambil untuk anggota toko saat area `/app` terbuka dalam keadaan daring dan setiap kali perangkat kembali daring. Formulir Titipan Masuk dan Hitung Sisa membaca salinan yang sama tanpa jalur kode terpisah.

## Akibat

- Satu tempat yang tahu data mana berasal dari salinan, dan bisa diuji dengan IndexedDB asli di uji peramban.
- Paket kerja nanti cukup mengambil jawaban yang sama lebih awal, lalu formulir membaca salinannya lewat `apiClient`.
- Halaman `/app` yang belum pernah dibuka tetap jatuh ke `/luring` saat tanpa sinyal.
