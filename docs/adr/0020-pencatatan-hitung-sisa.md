# 0020. Pencatatan Hitung Sisa

## Konteks

FR-CNT-01 sampai 07 dan BR-CNT-01 sampai 09 meminta Hitung Sisa per mitra yang membuat mutasi `SALE` dan `RETURN`, alokasi penjualan FEFO dengan nilai per alokasi, rekap bernomor, dan pemberitahuan, semuanya dalam satu transaksi. Spesifikasi bagian 12 meminta dua Hitung Sisa bersamaan diselesaikan dengan menolak yang kedua dengan `STOCK_CHANGED`. Beberapa hal belum ditetapkan: cara server tahu stok yang dilihat pengguna, nama nilai keterangan waktu hitung, sumber data formulir, dan urutan alokasi di isi yang disegel.

## Keputusan

1. Setiap baris permintaan membawa `recorded_qty`, yaitu stok tercatat yang dilihat pengguna (BR-PWA-06). Server mengunci baris kemitraan dengan `FOR UPDATE`, menghitung ulang stok dari tampilan `batch_balances`, dan membalas 409 `STOCK_CHANGED` bila ada baris yang berbeda. Permintaan kedua yang menunggu kunci melihat stok sesudah yang pertama, sehingga hanya satu yang lolos.
2. Keterangan waktu hitung disimpan sebagai `STORE_CLOSING`, `PICKUP`, `BEFORE_DROP`, dan `OTHER`.
3. `GET /partnerships/{id}/count-sheet` memberi toko produk berstok beserta batch, saldo, dan syarat salinannya, supaya pratinjau app memakai fungsi FEFO dan hitungan yang sama dengan server. Hanya toko yang boleh membukanya.
4. Waktu hitung paling awal tiga hari kalender sebelum hari ini, paling lambat 5 menit di depan jam server, dan tidak lebih awal dari Titipan Masuk yang tidak dibatalkan atau Hitung Sisa terakhir pada kemitraan itu (`BEFORE_LAST_DOCUMENT`). Titipan Masuk kini juga menolak tanggal sebelum Hitung Sisa terakhir (BR-DRP-06).
5. Isi yang disegel memuat rekap, baris, alokasi penjualan beserta nilainya, dan alokasi pengembalian. Alokasi di dalam satu baris diurutkan menurut ID supaya urutan baca tidak mengubah hash.
6. Kolom `sale_allocations.settlement_id` sudah ada tanpa rujukan. Rujukannya ke tabel Setoran, syarat periode yang sudah disetor (BR-CNT-06), dan Setoran `ON_COUNT` (BR-CNT-04) ditambahkan di Tahap 7.
7. `COUNT_RECORDED` dikirim ke semua anggota aktif penitip karena tidak memuat nilai uang gabungan.

## Akibat

- Titipan Masuk, pembatalan, dan Hitung Sisa pada satu kemitraan saling menunggu di kunci baris kemitraan.
