# 0037. Papan Kebutuhan dan tanggapan penitip

## Konteks

FR-NED-03 sampai FR-NED-05 dan FR-NED-07 meminta Papan Kebutuhan yang terurut dari yang terdekat, tanda Kebutuhan terbuka di kartu toko, tanggapan penitip lewat Ajuan yang tertaut ke Kebutuhan, dan pemberitahuan ke penitip dengan produk di kategori yang sama dalam jarak 10 km. BR-NED-03 meminta toko diingatkan 3 hari sebelum masa tayang berakhir, dan spesifikasi 21.5 menjadwalkannya pukul 06.00. Spesifikasi tidak menetapkan titik ukur jarak 10 km, tujuan tautan pemberitahuan, siapa yang boleh menanggapi, dan cara mencegah pengingat ganda.

## Keputusan

1. Papan Kebutuhan adalah rute `GET /api/v1/directory/needs` di modul `directory`, memakai alur pencarian yang sama (ADR 0032) dengan urutan terdekat atau terbaru dan saringan kategori, jarak, dan teks. Hanya Kebutuhan `OPEN` yang belum lewat masa tayang dari toko yang tampil di Cari Mitra.
2. Kartu toko membawa kategori Kebutuhan terbukanya. Saringan `has_open_need` (FR-DIR-06) dan bagian skor "Kebutuhan terbuka pada kategori produk penitip" (30 poin) memakai data yang sama, di direktori toko maupun di direktori penitip dan produk bila pemirsanya toko. Ini menutup butir 8 ADR 0032.
3. Tanggapan adalah Ajuan langsung dengan `need_id`. Hanya penitip yang boleh menanggapi, sasarannya harus toko pemilik Kebutuhan, dan Kebutuhan harus terbuka. Selain itu dibalas galat isian `need_id` atau `NEED_NOT_OPEN` dengan pesan bagian 5.
4. Penitip yang diberi tahu adalah organisasi penitip `ACTIVE` dengan minimal satu produk aktif di kategori itu dan lokasi publik paling jauh 10 km dari lokasi persis toko. Pemberitahuan `NEED_POSTED` hanya untuk OWNER dan menaut ke formulir Ajuan dengan sasaran dan Kebutuhan terisi, karena tautan pemberitahuan harus di area `/app`.
5. Pengingat berjalan pukul 06.00 dan menandai `reminded_at` dengan `UPDATE ... RETURNING` di transaksi yang sama dengan pemberitahuannya, sehingga setiap Kebutuhan diingatkan sekali. Perpanjangan mengosongkan `reminded_at`.

## Akibat

- Halaman publik Kebutuhan dan Papan Kebutuhan tanpa masuk tetap di 11.5.
