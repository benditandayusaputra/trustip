# 0063: Cara bayar penitip dan bukti bayar wajib

Tanggal: 2 November 2026

## Konteks

Pengguna meminta penitip bisa mengatur cara dibayar, yaitu tunai, QRIS, atau transfer bank, dan setiap pembayaran toko wajib disertai bukti berupa gambar atau PDF. Sebelumnya info pembayaran hanya teks bebas dan gambar QRIS, toko bisa memilih metode apa pun, dan bukti hanya wajib untuk transfer dan QRIS. Bukti juga hanya boleh gambar, karena R-SEC-70 mewajibkan setiap unggahan dikodekan ulang.

## Keputusan

1. Migrasi 00028 menambah `organizations.payment_methods` (satu sampai tiga dari `CASH`, `TRANSFER`, `QRIS`), `bank_name`, `bank_account_number`, dan `bank_account_holder`. Penitip lama mendapat metode dari datanya: tunai selalu, transfer bila teks info terisi, QRIS bila gambar QRIS ada. Organisasi baru mulai dengan tunai saja.
2. `PATCH /organizations/{id}/payment-info` menerima keempat isian baru. Minimal satu metode wajib dipilih. Transfer mewajibkan nama bank, nomor rekening (angka, spasi, atau tanda hubung, 5 sampai 30 karakter), dan atas nama. QRIS mewajibkan gambar QRIS.
3. Halaman Setoran menampilkan metode, rekening, dan catatan penitip kepada toko. Toko hanya bisa membayar dengan metode yang diterima penitip, dan bukti wajib untuk semua metode. Penitip yang menandai lunas sendiri (FR-STL-07) tetap tidak perlu bukti.
4. Unggahan bertujuan `PAYMENT_PROOF` boleh berupa PDF. PDF tidak bisa dikodekan ulang seperti gambar, jadi server menyimpannya apa adanya setelah memeriksa kepala `%PDF-` dan penanda `%%EOF`. PDF ditolak bila terkunci (`/Encrypt`) atau memuat skrip, aksi peluncur, lampiran, media, XFA, atau formulir kirim. Pemeriksaan itu juga berlaku pada nama yang disamarkan dengan `#xx` dan pada isi stream FlateDecode yang dibuka sampai 8 MB. PDF tetap privat dan hanya bisa dibaca kedua pihak Setoran lewat pemeriksaan yang sama dengan bukti gambar. Kode galat baru `INVALID_PROOF`.
5. Seeder demo memberi Roti Mas Bendi dan Keripik Bu Rina rekening contoh, memberi Dapur Sambal Mak Ijah tunai saja, dan melampirkan gambar nota contoh di setiap pembayaran toko.

## Akibat

Filter lain selain FlateDecode (misalnya LZW di dalam object stream) tidak dibuka. Celah sisa ini diterima karena PDF hanya dibuka oleh mitra yang sudah bertransaksi, dan bisa ditutup dengan pustaka PDF bila perlu. Migrasi hanya menambah kolom dan melonggarkan batasan, sehingga API versi sebelumnya tetap berjalan di atas skema baru.
