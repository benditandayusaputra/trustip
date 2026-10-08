# 0024. Mengakhiri kemitraan

## Konteks

FR-PRT-04 dan BR-PRT-02 sampai 04 meminta kemitraan bisa diakhiri hanya bila tidak ada stok di toko, dokumen yang menunggu konfirmasi, atau Setoran belum lunas, dengan daftar penghalang yang bertaut. BR-PRT-03 meminta Setoran untuk penjualan yang belum disetorkan diterbitkan lebih dulu. FR-PRT-01 yang ditunda dari Tahap 5 meminta daftar mitra menampilkan sisa stok dan saldo Setoran belum lunas.

## Keputusan

1. Permintaan mengakhiri berjalan dalam dua transaksi. Transaksi pertama mengunci kemitraan dan menerbitkan Setoran untuk semua penjualan yang belum tertaut sampai saat itu, apa pun jadwalnya. Transaksi kedua mengunci kemitraan lagi, menghitung penghalang, lalu mengakhiri atau menolak. Dengan begitu Setoran yang baru terbit tetap tersimpan walau pengakhiran ditolak, dan ia muncul sebagai penghalang "belum lunas".
2. Penghalang dihitung di satu kueri: saldo batch lebih dari 0, Titipan Masuk dan Hitung Sisa berstatus `RECORDED` yang belum lewat 48 jam, pembayaran yang menunggu, dan Setoran `ISSUED` tanpa pembayaran yang menunggu. Penolakan memakai 409 `PARTNERSHIP_HAS_BLOCKERS` dengan pesan yang menyebut setiap penghalang. `GET /partnerships/{id}/blockers` memberi daftar yang sama beserta ID dokumen untuk tautan di app.
3. Mengakhiri dan melihat penghalang hanya untuk OWNER (`END_PARTNERSHIPS`). Pengakhiran disegel sebagai `PARTNERSHIP_ENDED` dengan alasan, dan pihak lawan diberi tahu.
4. Sakelar "tampil sebagai tempat membeli" (FR-PRT-03) hanya bisa diubah OWNER toko pada kemitraan itu, lewat `PATCH /partnerships/{id}/public-listing`.
5. Daftar mitra menyertakan `stock_qty` dari tampilan `batch_balances` dan `unpaid_amount` dari Setoran `ISSUED`.

## Akibat

- Kemitraan yang berakhir tetap bisa dibaca (FR-PRT-05), dan kedua pihak bisa bermitra lagi lewat Ajuan baru karena indeks unik pasangan hanya berlaku untuk status `ACTIVE` (BR-PRT-04).
