# 0035. Tawar-menawar Ajuan

## Konteks

FR-PRP-03 dan FR-PRP-04 meminta penerima bisa Tawar dengan mengirim revisi, dan setiap revisi menampilkan perbedaannya dengan revisi sebelumnya. BR-PRP-08 meminta masa berlaku dihitung ulang 14 hari setiap ada revisi. Kriteria penerimaan bagian 6 meminta giliran pindah ke pengirim awal dan baris yang berubah ditandai. Spesifikasi tidak menetapkan bentuk revisi tawaran untuk Ajuan perubahan, cara menjaga dua tawaran bersamaan, dan bentuk beda revisi.

## Keputusan

1. Tawaran adalah revisi lengkap dengan bentuk yang sama seperti revisi pertama: syarat umum, daftar produk, dan pesan. Hanya pihak yang mendapat giliran yang boleh menawar, sama seperti menerima (BR-PRP-02). Undangan yang belum diklaim belum punya pihak bergiliran sehingga belum bisa ditawar.
2. Untuk Ajuan baru, tawaran hanya memuat baris tambah. Untuk Ajuan perubahan, tawaran diperiksa terhadap versi Kesepakatan yang berlaku dengan aturan ADR 0017.
3. Tawaran mengunci baris Ajuan dengan `FOR UPDATE`, lalu menaikkan `current_revision_no` dengan `UPDATE` bersyarat status `OPEN` dan nomor revisi lama. Tawaran kedua yang datang bersamaan mendapat `PROPOSAL_ALREADY_DECIDED`.
4. Giliran pindah ke pihak lawan penulis tawaran, `expires_at` menjadi 14 hari dari waktu tawaran, dan pemberitahuan `PROPOSAL_COUNTERED` dikirim ke OWNER pihak lawan dengan teks katalog peristiwa.
5. Beda revisi dihitung di `proposal/domain` dari dua revisi berurutan: isian syarat umum yang berubah, dan setiap baris produk bertanda `ADDED`, `CHANGED`, `KEPT`, atau `REMOVED`. Revisi pertama menandai semua barisnya `ADDED`. Rute `GET /api/v1/proposals/{id}/revisions` mengembalikan seluruh riwayat dalam dua kueri.

## Akibat

- Menerima Ajuan selalu memakai revisi terakhir, sehingga Kesepakatan memuat harga hasil tawar.
