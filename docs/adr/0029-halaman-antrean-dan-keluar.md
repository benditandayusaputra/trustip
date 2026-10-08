# 0029. Halaman Antrean dan keluar dengan catatan menunggu

## Konteks

FR-PWA-09, FR-PWA-10, BR-PWA-08, R-OFF-12, dan R-OFF-16 meminta halaman Antrean dengan status tiap catatan, tombol kirim sekarang dan buang, jumlah catatan di kepala app, tanda "dicatat luring" di dokumen, dan pilihan saat keluar. Spesifikasi tidak menyebut berapa lama catatan yang sudah terkirim tetap terlihat, padahal kriteria penerimaan bagian 19 meminta tanda catatan diganti nomor dokumen setelah terkirim.

## Keputusan

1. Catatan yang terkirim tetap di antrean dengan nomor dan tautan dokumennya, lalu dihapus oleh pengirim setelah berumur lebih dari 24 jam.
2. Kepala app menampilkan tautan ke Antrean dengan jumlah catatan yang belum terkirim. Tautan itu hanya muncul bila jumlahnya lebih dari nol.
3. Catatan yang perlu ditinjau atau tertahan tidak bisa dikirim ulang apa adanya dari halaman Antrean. Pengguna meninjaunya atau membuangnya. Membuang catatan langsung memicu kiriman supaya catatan yang tertahan di belakangnya ikut berjalan.
4. Keluar, baik dari perangkat ini maupun dari semua perangkat, lebih dulu memeriksa catatan milik pengguna itu yang belum terkirim. Bila ada, dialog menawarkan kirim dulu, simpan di perangkat, atau buang. Kirim dulu baru keluar bila semua catatan terkirim. Simpanan baca tetap dihapus apa pun pilihannya (R-OFF-16).
5. Dokumen yang dicatat luring menampilkan lencana "Dicatat luring" beserta waktu dicatat dan waktu terkirim, kepada kedua pihak, dari jawaban dokumen yang sama.
6. Halaman Antrean menjelaskan bahwa di iPhone catatan terkirim saat app dibuka kembali (BR-PWA-09).

## Akibat

- Pengguna lain yang masuk di perangkat yang sama tidak melihat dan tidak mengirim catatan yang disimpan pengguna sebelumnya, karena daftar dan pengirim memakai indeks pengguna.
