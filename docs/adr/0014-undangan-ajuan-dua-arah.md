# 0014. Undangan Ajuan dari kedua tipe organisasi

## Konteks

FR-PRP-07 meminta Ajuan tanpa sasaran yang dibagikan sebagai tautan `/undangan/<kode>`. Model data di spesifikasi bagian 22.3 menulis bahwa `store_org_id` boleh kosong sebelum undangan diklaim, tetapi kriteria penerimaan bagian 6 juga menyebut "toko membuka tautan undangan milik toko lain", yang berarti toko juga bisa mengundang. BR-PRP-01 meminta produk di Ajuan selalu milik penitip pihak Ajuan itu. Spesifikasi tidak menetapkan bentuk nomor Ajuan, tidak menyebut kode galat untuk undangan yang kedaluwarsa, untuk Ajuan terbuka ganda (BR-PRP-04), dan untuk pasangan yang sudah bermitra (BR-PRP-05).

## Keputusan

1. Kedua tipe organisasi bisa membuat undangan. Kolom pihak yang belum diketahui, `consignor_org_id` atau `store_org_id`, kosong sampai diklaim. Batasan basis data meminta pengirim selalu salah satu pihak.
2. Undangan dari toko tidak boleh memuat produk, karena penitipnya belum diketahui. Penitip yang mengklaimnya bisa menerima syarat umum lalu menambah produk lewat jalur cepat tambah produk (FR-AGR-06).
3. Mengklaim undangan menjadikan organisasi pengklaim pihak lawan dan giliran berpindah kepadanya. Klaim ulang oleh pihak yang sama dianggap berhasil tanpa mengubah apa pun. Organisasi lain mendapat 404. Klaim dijaga satu `UPDATE` bersyarat, sehingga dua klaim bersamaan hanya satu yang lolos.
4. Ringkasan undangan sebelum diklaim hanya memuat nama dan tipe pengirim, jumlah produk, dan batas waktu. Syarat dan harga baru terlihat setelah diklaim, sesuai tabel hak data (Ajuan hanya untuk kedua pihak).
5. Nomor Ajuan `AJ-<YYYYMM>-<4 angka urut per pengirim>`, mengikuti pola nomor Setoran (FR-STL-03). Penghitungnya tabel `document_sequences` yang dinaikkan di transaksi pembuatan.
6. Kode galat tambahan: `INVITE_WRONG_ORG_TYPE` (422, pesan dari bagian 24), `PROPOSAL_EXPIRED` (409), `PROPOSAL_ALREADY_OPEN` (409), dan `ALREADY_PARTNERS` (409, mengarahkan ke Ajuan perubahan sesuai BR-PRP-05).
7. Unggah foto produk memakai aksi `MANAGE_PRODUCTS`, bukan `EDIT_PROFILE`, karena STAFF penitip boleh mengelola produk. Aksi yang dibutuhkan ditentukan dari tujuan unggahan.

## Akibat

- Undangan dari toko menghasilkan Kesepakatan tanpa produk bila langsung diterima. Produk ditambahkan sesudahnya lewat Ajuan perubahan.
- Batas 10 Ajuan baru per hari (BR-PRP-06) belum ditegakkan untuk undangan dan ditambahkan bersama Ajuan langsung dari Cari Mitra.
