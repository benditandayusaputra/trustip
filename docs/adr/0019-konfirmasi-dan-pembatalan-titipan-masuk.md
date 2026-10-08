# 0019. Konfirmasi dan pembatalan Titipan Masuk

## Konteks

FR-DRP-06, FR-DRP-07, dan FR-DRP-10 meminta penitip menekan Sesuai, titipan yang tidak ditanggapi 48 jam dianggap sesuai, dan toko bisa membatalkan dalam 15 menit selama stoknya belum berubah. BR-DRP-08 meminta status "dianggap sesuai" dihitung saat dokumen dibaca dan ditetapkan tugas harian. Stok adalah jumlah mutasi (bagian 11), sehingga pembatalan harus mengembalikan stok tanpa menghapus riwayat. Spesifikasi tidak menyebut kode galat untuk dokumen yang sudah ditanggapi, batas 15 menit yang lewat, atau stok yang sudah dihitung, dan katalog pemberitahuan tidak punya peristiwa pembatalan.

## Keputusan

1. Status turunan dihitung di `drop/domain` (`View.Effective`): titipan `RECORDED` yang waktu terimanya sudah lewat 48 jam dibaca sebagai `CONFIRMED` otomatis pada batas itu. Tugas harian di Tahap 7 menetapkannya ke basis data.
2. Konfirmasi dan pembatalan mengunci baris titipan dan baris kemitraannya dengan `FOR UPDATE`, lalu menyimpan status dengan `UPDATE` bersyarat `status = 'RECORDED'`. Keputusan kedua mendapat 409 `DOCUMENT_NOT_PENDING`.
3. Pembatalan hanya oleh toko pencatat, paling lambat 15 menit setelah waktu terima, dan hanya bila batch titipan itu belum punya mutasi selain `RECEIVE`. Galatnya 409 `CANCEL_WINDOW_PASSED` dan 409 `DROP_ALREADY_COUNTED`.
4. Stok dikembalikan dengan mutasi `ADJUSTMENT` bernilai negatif per batch yang bersumber dari titipan itu, bukan dengan menghapus mutasi `RECEIVE`. Kartu stok menampilkan keduanya.
5. `DROP_CONFIRMED` dikirim hanya ke pengguna pencatat di toko, sesuai katalog. Pembatalan tidak memberi tahu siapa pun karena tidak ada di katalog, tetapi tetap disegel dan masuk log audit.
6. Jawaban titipan memuat `cancel_deadline` supaya app bisa menyembunyikan tombol Batalkan setelah batasnya. Server tetap yang memutuskan.

## Akibat

- Hitung Sisa nanti mengunci baris kemitraan yang sama, sehingga pembatalan dan Hitung Sisa atas stok yang sama tidak bisa lolos bersamaan.
