# 0032. Direktori toko di Cari Mitra

## Konteks

FR-DIR-01 sampai FR-DIR-12 meminta direktori toko yang terurut dari yang terdekat, dengan titik asal yang bisa diganti, saring jarak yang meluas sendiri (BR-DIR-03), saringan toko, urutan lain, pencarian teks, dan skor kecocokan dengan alasan singkat (FR-DIR-10). Spesifikasi tidak menetapkan bentuk alasan skor, cara mengurutkan "paling cocok" tanpa menulis ulang rumus skor di SQL, sumber angka reputasi sebelum ringkasan harian ada (FR-TRS-01 di 10.5), dan siapa yang mengisi `organizations.last_active_at` untuk bagian keaktifan. Kebutuhan Toko (bagian 5) baru dibuat di 9.6.

## Keputusan

1. Modul baru `directory` hanya membaca. Satu kueri daftar memakai daftar putih urutan (`NEAREST`, `REPUTATION`, `NEWEST`, dan `CITY` bila titik asal tidak ada, BR-DIR-04) yang dipetakan ke kunci urut teks dan angka, sehingga kursor tetap kunci set tanpa `OFFSET` (R-RES-04, R-SEC-58). Saring jarak memakai `ST_DWithin` atas `location_public` supaya indeks spasial terpakai.
2. Perluasan jarak (BR-DIR-03) dihitung di halaman pertama dengan satu kueri yang menghitung toko per langkah jarak, lalu fungsi murni `EffectiveRadius` memilih langkah pertama yang berisi minimal 5 toko, atau semua jarak. Jarak yang dipakai disimpan di kursor supaya halaman berikutnya konsisten. Satu permintaan memakai paling banyak tiga kueri.
3. Skor dihitung di Go dengan fungsi murni yang sama untuk semua urutan. "Paling cocok" mengambil 200 toko terdekat di dalam jarak terpilih, memberi skor, lalu mengurutkan di memori. Batas ini menyederhanakan dan tercatat sebagai batas: toko di luar 200 terdekat tidak ikut diurutkan menurut skor. Pengganti bila dibutuhkan: ringkasan skor tersimpan per pasangan.
4. Alasan skor dikirim sebagai kode bertipe (`DISTANCE`, `NEEDS_CATEGORY`, `ACCEPTS_CATEGORY`, `ON_TIME_PAYMENT`, `ESTABLISHED`, `PREFERRED_STORE_TYPE`, `ACTIVE_RECENTLY`) dengan nilainya, paling banyak tiga, dalam urutan bagian skor. Teks dan format jarak dibuat app dari katalog teks (R-APP-09, R-APP-12). Skor hanya dihitung bila pemirsanya penitip.
5. Tabel `org_stats` dibuat sekarang dan dibaca dengan `LEFT JOIN`. Selama belum diisi tugas harian di 10.5, semua toko berstatus "Baru" dan mendapat 5 poin reputasi.
6. `last_active_at` diisi pemicu basis data setiap kali sesi baru dibuat, yaitu saat masuk dan saat token disegarkan, sehingga modul `auth` tidak perlu tahu tentang direktori.
7. Kategori bawaan saringan (BR-DIR-06) adalah kategori produk aktif milik penitip. Parameter `category_id=any` membuang saringan kategori, dan kategori yang dipakai dikembalikan di `meta.search`.
8. Tanda dan saringan Kebutuhan terbuka serta bagian skor Kebutuhan ditambahkan bersama Papan Kebutuhan di 9.7, setelah tabel `needs` ada.
9. Selama Tahap 9 rute direktori meminta sesi. Akses pengunjung dan cache tepi dibuka di 11.5 (FR-PUB-07).

## Akibat

- Urutan reputasi baru bermakna setelah 10.5 mengisi `org_stats`.
- Pembatas laju Cari Mitra 60 per menit per IP (tabel 7.9) dipasang pada rute direktori.
