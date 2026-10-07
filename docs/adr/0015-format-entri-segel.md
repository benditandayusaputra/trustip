# 0015. Format entri segel dan cara membangunnya

## Konteks

R-SEC-28 sampai 31 dan BR-SIG-01 sampai 05 menetapkan isi entri segel, hash, tanda tangan Ed25519, rantai per kemitraan, dan tabel yang hanya bisa ditambah. Rincian bentuk tiap bidang, misalnya cara menulis waktu dan hash di bentuk kanonik, belum ditetapkan, padahal pihak luar harus bisa memeriksa segel secara mandiri (FR-SIG-04). Segel dibuat oleh banyak modul, sedangkan modul tidak boleh saling mengimpor (R-API-02).

## Keputusan

1. Bentuk kanonik ditulis sendiri di `kernel/canonical` sebagai bagian RFC 8785 yang dipakai Trustip: objek dengan kunci terurut menurut unit kode UTF-16, larik, string, dan bilangan bulat. Nilai lain ditolak. String di-escape seperti `JSON.stringify`, sehingga huruf non-ASCII dan U+2028 tetap apa adanya.
2. Entri yang di-hash berisi kunci `v`, `partnership_id`, `seq`, `event_type`, `entity_type`, `entity_id`, `entity_number`, `content_hash`, `actor_user_id`, `actor_org_id`, `occurred_at`, dan `prev_hash`. Hash ditulis heksadesimal huruf kecil, `prev_hash` kosong untuk entri pertama, dan `occurred_at` ditulis UTC dengan enam angka mikrodetik (`2026-10-19T08:00:00.123456Z`) karena itu ketelitian `timestamptz`.
3. `entry_hash` adalah SHA-256 bentuk kanonik entri. Tanda tangan Ed25519 dibuat atas 32 byte `entry_hash`. Pengenal kunci disimpan di samping entri, tidak ikut di-hash, supaya rotasi kunci tidak mengubah entri.
4. Kode verifikasi adalah 16 karakter pertama base32 standar `entry_hash`, tanpa padding, dikelompokkan empat dengan tanda hubung.
5. Modul pemilik peristiwa mendefinisikan port penyegel di layanannya sendiri. Akar komposisi menyambungkannya ke `seal/service.Recorder`, yang mengunci baris kemitraan, menambah entri, dan menggeser kepala rantai di transaksi milik pemanggil.
6. Pemicu `seals_append_only` menolak ubah dan hapus. Hapus hanya lolos bila kedua organisasi pada kemitraan itu bertanda demo. `TRUNCATE` selalu ditolak.
7. Kueri yang sengaja merusak data hanya untuk uji ditaruh di `db/queries/tamper.sql` dengan awalan nama `Tamper`.

## Akibat

- Vektor `api/testdata/vectors/seal.json` mengunci bentuk kanonik isi, entri, hash, kode, dan tanda tangan. Mengubah format berarti menaikkan `v` dan menambah vektor baru, bukan mengubah yang lama.
- Pemeriksa mandiri cukup menyusun ulang entri dari data halaman verifikasi dan kunci publik, tanpa kode Trustip.
