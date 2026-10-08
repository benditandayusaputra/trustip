# 0030. Tinjauan catatan yang bentrok

## Konteks

R-OFF-09 meminta layar yang menampilkan isian pengguna di samping keadaan terbaru untuk catatan yang ditolak server. R-OFF-10 meminta perbaikan memakai kunci idempotensi baru. Spesifikasi tidak menyebut di mana perbaikan dilakukan.

## Keputusan

1. Layar `/app/antrean/<id>` memuat entri dari perangkat dan keadaan terbaru dari server atau dari simpanan baca: lembar Hitung Sisa untuk Hitung Sisa, Kesepakatan aktif untuk Titipan Masuk. Hitung Sisa menampilkan per produk stok yang dilihat pengguna, stok tercatat sekarang, sisa yang diisi, dan yang dikembalikan. Titipan Masuk menampilkan versi Kesepakatan yang dipakai dan yang berlaku, serta apakah tiap produk masih ada di Kesepakatan.
2. Tombol "Perbaiki dan kirim ulang" membuka formulir yang sama dengan parameter `antrean`. Formulir memuat stok dan Kesepakatan terbaru, lalu mengisi ulang jumlah, tanggal, keterangan, dan catatan dari entri. Validasi formulir langsung menandai isian yang tidak lagi cocok.
3. Menyimpan dari formulir itu memperbarui entri yang sama: revisi naik satu, kunci idempotensi baru, status kembali `PENDING`, dan waktu simpan semula tetap, sehingga urutannya di antrean kemitraan tidak berubah. Bila server langsung menolak lagi, galatnya tampil di formulir dan entri dihapus, sama seperti simpanan baru.

## Akibat

- Tidak ada formulir kedua untuk tinjauan. Aturan dan pratinjau yang sama berlaku untuk catatan baru maupun yang diperbaiki.
