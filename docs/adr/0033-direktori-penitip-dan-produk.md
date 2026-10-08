# 0033. Direktori penitip dan produk

## Konteks

FR-DIR-02 meminta direktori penitip dengan tab Produk dan Penitip, dan FR-DIR-07 menyebut saringannya. Spesifikasi tidak menetapkan titik mana yang dipakai untuk saringan "jangkauan antar mencakup toko saya", bagaimana perluasan jarak (BR-DIR-03) dihitung di tab Produk, dan dari mana angka reputasi penitip diambil sebelum ringkasan harian (10.5) dan pelaporan (10.6) ada.

## Keputusan

1. Kedua tab memakai alur yang sama dengan direktori toko (ADR 0032): titik asal, langkah jarak, urutan daftar putih, kursor kunci set, dan skor yang dihitung di Go. Bagian pencarian bersama dibentuk sekali di `directory/domain`, saringan per daftar ditambahkan di atasnya.
2. Jarak ke penitip selalu dihitung ke lokasi publik yang dibulatkan (BR-ORG-03) dan app menampilkannya dengan awalan "±" (BR-DIR-01).
3. Saringan jangkauan antar memakai lokasi persis toko pemirsa, bukan titik asal pencarian, dan meminta lokasi itu sudah diatur. Bila belum, permintaan ditolak dengan galat isian `location`.
4. Saringan harga, masa simpan, dan kategori pada tab Penitip berarti penitip punya minimal satu produk aktif yang memenuhinya. Pada tab Produk saringan berlaku per produk dan hanya produk yang sedang mencari toko yang tampil.
5. Perluasan jarak di tab Produk menghitung penitip yang punya produk cocok, bukan jumlah produknya. Akibatnya perluasan bisa sedikit lebih jarang terjadi bila satu penitip punya banyak produk.
6. Skor dihitung bila pemirsanya toko: kategori produk penitip atau produk itu terhadap kategori yang diterima toko, jenis toko terhadap incaran penitip, jangkauan antar penitip sebagai `R`, dan reputasi penitip dari `org_stats.active_partner_count` tanpa laporan terbukti sampai 10.6 ada.
7. Kartu produk tidak pernah memuat harga setor atau komisi (BR-AUTHZ-04), hanya harga jual saran.

## Akibat

- Selama `org_stats` belum diisi, semua penitip mendapat 5 poin reputasi dan urutan reputasi belum bermakna.
