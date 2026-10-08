# 0034. Ajuan langsung dari Cari Mitra

## Konteks

FR-PRP-01 dan FR-DIR-11 meminta Ajuan dikirim langsung dari kartu dan profil di Cari Mitra, ditambah tombol Tanya lewat WhatsApp. BR-PRP-06 membatasi 10 Ajuan baru per hari dan meminta jeda 7 hari setelah ditolak, dan R-SEC-44 meminta batas bisnis ditegakkan di basis data. Spesifikasi tidak menyebut kode galat untuk jeda setelah ditolak, cara menghitung "per hari", siapa yang boleh menjadi sasaran, dan bagaimana nomor WhatsApp yang tidak termasuk daftar putih publik sampai ke tombol.

## Keputusan

1. Rute `POST /api/v1/proposals` membawa `target_organization_id`. Sasaran harus organisasi `ACTIVE` yang tampil di Cari Mitra dan bertipe berlawanan. Sasaran lain dibalas 404 atau galat isian. Produk tetap milik penitip pihak Ajuan (BR-PRP-01), sehingga toko hanya bisa menyebut produk penitip sasaran.
2. Kuota harian disimpan di tabel `proposal_quotas` per organisasi per tanggal Asia/Jakarta, dengan batasan `CHECK (sent BETWEEN 1 AND 10)`. Penambahan dilakukan di transaksi pembuatan Ajuan, sehingga Ajuan kesebelas ditolak basis data walau dua permintaan datang bersamaan. Kuota yang sama dipakai undangan (FR-PRP-07), menutup catatan di ADR 0014. Ajuan perubahan tidak dihitung karena bukan Ajuan baru.
3. Jeda dihitung dari `decided_at` Ajuan terakhir milik pengirim yang ditolak oleh sasaran yang sama. Kode galat baru `PROPOSAL_COOLDOWN` (409) menyebut tanggal kapan Ajuan bisa dikirim lagi. `PROPOSAL_LIMIT_REACHED` memakai pesan bagian 24 dengan status 409.
4. Pasangan yang sudah bermitra mendapat `ALREADY_PARTNERS` seperti klaim undangan (BR-PRP-05).
5. Nomor WhatsApp dibuka lewat `GET /api/v1/directory/organizations/{id}` hanya untuk pengguna yang masuk dan hanya bila organisasi itu tampil di Cari Mitra. App menyusun tautan `wa.me` sendiri dari nomor yang sudah dinormalkan (R-SEC-64), sesuai BR-AUTHZ-04 yang melarang nomor ikut di data publik.
6. Profil publik `/t/<slug>` dan `/u/<slug>` baru ada di 11.3, sehingga tombol di profil ditambahkan di sana dengan tautan yang sama.

## Akibat

- Undangan yang dibuat di atas kuota kini ditolak dengan `PROPOSAL_LIMIT_REACHED`.
