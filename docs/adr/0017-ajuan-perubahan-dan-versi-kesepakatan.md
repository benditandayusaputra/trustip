# 0017. Ajuan perubahan dan versi Kesepakatan

## Konteks

FR-PRP-08 meminta Ajuan jenis perubahan untuk kemitraan aktif: ubah syarat, tambah produk, dan hentikan produk. BR-PRP-03 meminta penerimaannya membuat versi Kesepakatan berikutnya, BR-AGR-05 meminta versi lama menjadi `SUPERSEDED` tanpa diubah, dan FR-AGR-04 meminta riwayat versi dengan beda antar versi. FR-AGR-06 meminta jalur cepat tambah produk. Spesifikasi tidak menetapkan isi Ajuan perubahan yang hanya menyebut sebagian produk, teks pemberitahuan saat Ajuan perubahan masuk, dan bentuk beda versi.

## Keputusan

1. Revisi Ajuan perubahan memuat syarat umum lengkap dan hanya baris produk yang berubah, masing-masing bertindakan `ADD`, `UPDATE`, atau `STOP`. Produk yang tidak disebut tetap seperti versi sebelumnya. Baris `STOP` menyimpan syarat terakhirnya supaya batasan tabel tetap berlaku.
2. `ADD` hanya untuk produk yang belum aktif di Kesepakatan, termasuk produk yang pernah dihentikan. `UPDATE` dan `STOP` hanya untuk produk yang masih aktif. Ajuan yang tidak mengubah apa pun ditolak.
3. Ajuan perubahan langsung punya kedua pihak, giliran ada pada pihak lawan, dan menaut ke kemitraannya. Batas satu Ajuan terbuka per jenis per pasangan (BR-PRP-04) memisahkan Ajuan baru dan Ajuan perubahan.
4. Menerima Ajuan perubahan mengunci versi aktif dengan `FOR UPDATE`, mengubahnya menjadi `SUPERSEDED`, membuat versi berikutnya, lalu menyegelnya dengan peristiwa `AGREEMENT_ACCEPTED`. Nomornya nomor Ajuan perubahan ditambah versi.
5. Pemberitahuan: `PROPOSAL_RECEIVED` ke OWNER pihak lawan dengan teks "{Org} mengajukan perubahan Kesepakatan ({no})", lalu saat diterima `PROPOSAL_ACCEPTED` ke penulis revisi dan `AGREEMENT_UPDATED` ke OWNER kedua pihak sesuai katalog peristiwa.
6. Beda versi dihitung dari dua versi berurutan di `agreement/domain`: perubahan syarat umum per isian, dan produk yang ditambah, diubah, atau dihentikan beserta syarat sebelum dan sesudahnya. Versi 1 menampilkan semua produknya sebagai tambahan.
7. Jalur cepat tambah produk memakai rute Ajuan perubahan yang sama dengan satu baris `ADD` dan syarat umum yang tidak berubah. Antarmukanya yang menyederhanakan langkah.

## Akibat

- Tawar-menawar atas Ajuan perubahan menunggu fitur tawar-menawar, sama seperti Ajuan baru.
- Titipan Masuk nanti harus menolak produk yang tidak aktif di versi Kesepakatan yang berlaku (BR-AGR-07).
