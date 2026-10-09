# 0040. Pemberitahuan langsung lewat SSE bertiket

## Konteks

FR-NTF-02 meminta pembaruan langsung lewat Server-Sent Events, FR-NTF-03 meminta pemeriksaan tiap 30 detik bila aliran gagal, dan BR-NTF-05 meminta aliran berhenti saat tab tersembunyi lebih dari 5 menit. R-SEC-26 menetapkan aliran tidak lewat proxy dan dijaga tiket sekali pakai berumur 30 detik. R-PERF-07 meminta denyut tiap 25 detik dan klien lambat diputus. Tabel 7.9 membatasi tiket 10 per menit dan 3 koneksi per pengguna. BR-NTF-01 membuat pemberitahuan di dalam transaksi peristiwanya, sehingga peristiwa langsung tidak boleh terkirim sebelum transaksi itu jadi.

## Keputusan

1. Hub dan tiket hidup di memori proses API (`platform/sse`), sama seperti pembatas laju dan cache nonce. Tidak ada `LISTEN/NOTIFY`, karena koneksi pendengar yang selalu terbuka membuat basis data tidak pernah tidur (R-PERF-04).
2. `Pool.AfterCommit` menunda kiriman sampai transaksi berhasil di-commit. Transaksi yang batal tidak mengirim apa pun. Di luar transaksi, kiriman langsung berjalan.
3. Penyambung di akar komposisi membungkus `Inbox.Deliver`: setelah pemberitahuan tersimpan, satu peristiwa `notification` berisi `id` dan `type` saja (BR-NTF-04) didaftarkan untuk tiap penerima.
4. `POST /api/v1/stream-tickets` lewat proxy dan butuh sesi. `GET /api/v1/stream?ticket=` dibuka peramban langsung ke API, dilewatkan dari tanda tangan tepi, dan menjawab CORS hanya untuk asal app. Tiket di alamat aman karena sekali pakai, berumur 30 detik, dan log akses tidak mencatat string kueri (R-LOG-03).
5. Koneksi keempat milik satu pengguna menutup aliran tertua pengguna itu. Server baru tahu koneksi yang ditinggal peramban saat denyut berikutnya, sehingga menolak koneksi baru membuat tab yang baru dimuat ulang kehilangan aliran langsung. Tab yang alirannya ditutup jatuh ke pemeriksaan berkala dan baru membuka aliran lagi saat kembali terlihat atau aktif, sehingga tab tidak saling menendang terus-menerus.
6. Tiap aliran punya antrean 16 peristiwa. Bila penuh, aliran diputus dan klien jatuh ke pemeriksaan berkala.
7. App membuka aliran dari komponen lonceng. Bila tiket atau aliran gagal, app memeriksa jumlah belum dibaca tiap 30 detik selama tab terlihat, serta saat tab kembali terlihat atau jendela kembali aktif, sambil mencoba membuka aliran lagi.

## Akibat

- API harus berjalan sebagai satu proses. Bila kelak lebih dari satu, hub perlu perantara bersama dan keputusan ini ditinjau ulang.
- App baru meminta tiket setelah halaman terbuka 2 detik, supaya perpindahan halaman yang cepat tidak menghabiskan kuota tiket 10 per menit. Pemberitahuan tetap sampai dalam 3 detik bagi halaman yang sudah terbuka.
- Setiap muat halaman penuh yang bertahan lebih dari 2 detik menambah satu permintaan tiket dan satu aliran. Uji ujung ke ujung menjalankan semua peramban dari satu IP, sehingga jumlah worker Playwright dibatasi tiga supaya lalu lintasnya tetap di bawah batas umum 300 permintaan per menit per IP. Batas itu sendiri tidak diubah.
- Sesi yang dicabut tidak memutus aliran yang sudah terbuka. Aliran hanya membawa ID dan jenis, dan rinciannya tetap lewat API yang memeriksa sesi.
