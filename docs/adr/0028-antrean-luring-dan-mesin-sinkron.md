# 0028. Antrean luring dan mesin sinkron

## Konteks

FR-PWA-06, BR-PWA-03 sampai 05, dan R-OFF-02 sampai 08, 13, 14 meminta semua penyimpanan Titipan Masuk dan Hitung Sisa lewat antrean di perangkat, dikirim berurutan per kemitraan oleh satu pengirim, dicoba ulang berjeda, dan hanya atas nama pengguna yang membuatnya. Beberapa hal tidak ditetapkan: pemetaan jawaban server ke status entri, nasib entri yang langsung ditolak saat pengguna masih di formulir, bentuk kunci untuk entri yang diperbaiki, dan sinkron latar belakang saat app tertutup.

## Keputusan

1. Entri tinggal di IndexedDB (`outboxStore`, indeks per pengguna). Isinya ID, nomor revisi, kunci idempotensi, jenis, kemitraan, nama mitra, permintaan, baris titipan beserta syarat dari paket kerja, waktu simpan, pengguna, organisasi, status, jumlah percobaan, galat terakhir, dan dokumen hasil. Tidak ada token atau data pembayaran (R-OFF-14).
2. Formulir selalu memasukkan entri ke antrean lalu langsung meminta kiriman. Bila terkirim, formulir membuka dokumennya. Bila tertunda, formulir menampilkan "Tersimpan di perangkat, menunggu sinyal." Bila server langsung menolak dengan 409 atau 422 saat pengguna masih di formulir, entri dihapus dan galatnya tampil di formulir seperti sebelumnya, karena formulir itulah layar tinjaunya.
3. Pemetaan jawaban: jaringan putus, 408, 429, dan 5xx membuat entri kembali `PENDING` dan kiriman berhenti untuk dicoba lagi. 401 setelah gagal menyegarkan sesi menghentikan kiriman tanpa mengubah entri (BR-PWA-05). Jawaban 4xx lain membuat entri `NEEDS_REVIEW`, dan entri berikutnya pada kemitraan yang sama menjadi `HELD`. Kemitraan lain tetap dikirim (BR-PWA-04).
4. Satu pengirim dijaga kunci web `trustip-outbox`. Tab lain yang meminta kiriman menunggu kunci itu, bukan mengirim bersamaan (R-OFF-07). Perubahan antrean diumumkan lewat `BroadcastChannel` supaya tab lain ikut memperbarui tampilannya.
5. Kiriman dipicu saat app dibuka dalam keadaan daring, saat perangkat kembali daring, saat app kembali ke depan, saat formulir disimpan, dan oleh percobaan ulang berjeda: 2 detik dikali dua tiap percobaan, paling lama 5 menit, dengan acakan setengah ke bawah dan ke atas (R-OFF-05).
6. Sinkron latar belakang peramban tidak dipakai. Antrean hidup di tingkat app (R-OFF-13), jadi tanpa tab yang terbuka tidak ada yang bisa mengirim. Catatan terkirim saat app dibuka lagi, sama seperti batas di iPhone (BR-PWA-09).
7. Setiap kiriman membawa `client_recorded_at` berupa waktu simpan entri, dan `recorded_offline` bernilai benar bila entri disimpan saat perangkat luring atau tidak terkirim pada percobaan pertama (ADR 0027).
8. Entri yang diperbaiki naik satu revisi dan mendapat UUID baru sebagai kunci idempotensi (R-OFF-10), karena API hanya menerima UUID.
9. Setelah ada entri terkirim, paket kerja diambil ulang. Formulir Hitung Sisa menambahkan titipan yang masih di antrean ke stok tercatat paket kerja sebagai batch sementara (R-OFF-04).

## Akibat

- Tidak ada jalur tulis kedua: daring maupun luring, formulir memakai antrean yang sama.
- Hitung Sisa yang masih di antrean tidak mengurangi stok di formulir berikutnya. Bila stok berbeda saat terkirim, server menolak dengan `STOCK_CHANGED` dan entri masuk tinjauan.
