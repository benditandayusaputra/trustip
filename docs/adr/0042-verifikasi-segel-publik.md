# 0042. Verifikasi segel publik dan pemeriksaan rantai harian

## Konteks

FR-SIG-03 meminta halaman `/verifikasi/<kode>` tanpa masuk, FR-SIG-04 meminta kunci publik dan cara memeriksa mandiri diterbitkan, dan FR-SIG-05 meminta tugas harian memeriksa keutuhan rantai segel tiap kemitraan. BR-SIG-06 dan R-SEC-34 membatasi isi halaman verifikasi. Isi dokumen tidak boleh dibuka untuk umum, sehingga halaman publik tidak bisa menghitung ulang `content_hash`.

## Keputusan

1. Halaman publik menyatakan "Segel sah" bila hash entri yang dihitung ulang cocok, tanda tangan Ed25519 cocok dengan kunci yang tercatat di entri, dan `prev_hash` sama dengan hash entri sebelumnya pada kemitraan yang sama. Pemeriksaan isi dokumen tetap di halaman dokumen bagi kedua pihak (R-SEC-33).
2. Kode diterima dalam huruf kecil, dengan atau tanpa tanda hubung dan spasi, lalu dinormalkan. Kode yang bentuknya salah dijawab 404 tanpa menyentuh basis data. Rute dibatasi 30 permintaan per menit per IP.
3. Segel yang dinyatakan tidak sah di halaman publik dicatat sebagai peristiwa keamanan `seal_invalid`.
4. `GET /api/v1/public/seal-keys` menerbitkan semua kunci publik di cincin kunci segel, termasuk kunci lama setelah rotasi, supaya segel lama tetap bisa diperiksa.
5. Tugas `seal-chain-check` berjalan setiap hari pukul 00.20 Asia/Jakarta, memeriksa rantai setiap kemitraan yang punya segel, dan mencatat `seal_chain_broken` per kemitraan yang rantainya putus.
6. Kode verifikasi di jejak segel setiap dokumen menjadi tautan ke halaman verifikasi.

## Akibat

- Panduan format kanonik dan langkah pemeriksaan mandiri ada di `docs/VERIFIKASI-SEGEL.md`.
- Pemeriksaan harian membaca seluruh rantai setiap hari. Bila jumlah segel kelak besar, pemeriksaan bisa dibatasi pada entri sejak pemeriksaan terakhir.
