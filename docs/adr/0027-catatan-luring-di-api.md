# 0027. Catatan luring di API

## Konteks

FR-PWA-10, BR-PWA-06, BR-PWA-07, R-OFF-09, dan R-OFF-11 meminta server menyimpan waktu catat di perangkat dan waktu terima, menandai dokumen yang dicatat luring, menolak bentrok stok dan versi Kesepakatan, dan menghitung jendela 48 jam dan 15 menit dari waktu terima. Kolom `client_recorded_at`, `received_at`, dan `recorded_offline` sudah ada sejak tabel `drops` dan `counts` dibuat, tetapi `client_recorded_at` selalu diisi waktu terima. Antrean luring mengirim ulang entri yang sama dengan kunci idempotensi yang sama, dan klien baru tahu sebuah entri tertunda setelah percobaan pertama gagal.

## Keputusan

1. `POST /drops` dan `POST /counts` menerima `client_recorded_at` (RFC 3339) dan `recorded_offline` (boolean). Keduanya boleh tidak dikirim. Jawaban dokumen membawa `client_recorded_at` di samping `received_at` dan `recorded_offline`.
2. Waktu perangkat yang lebih dari 5 menit di depan waktu terima diganti waktu terima. Waktu perangkat yang tidak dikirim juga sama dengan waktu terima.
3. Bila `occurred_at` atau `counted_at` tidak dikirim, waktu kejadian adalah waktu perangkat yang sudah disesuaikan, bukan waktu terima, sehingga catatan luring tercatat pada saat pengguna menyimpannya.
4. Batas mundur tiga hari (BR-DRP-06, BR-CNT-06) tetap diukur dari hari ini menurut jam server, terhadap waktu kejadian. Bila diukur dari waktu perangkat, klien bisa memundurkan batasnya dengan mengirim waktu perangkat yang lama. Catatan luring yang tertahan lebih dari tiga hari ditolak dan menjadi entri perlu ditinjau (R-OFF-09).
5. `client_recorded_at` ikut sidik permintaan. `recorded_offline` tidak, karena tanda itu baru pasti setelah percobaan pertama gagal, dan kiriman ulang dengan tanda berbeda tetap harus dibalas sebagai ulangan dokumen semula.
6. Jendela 15 menit pembatalan dan 48 jam "dianggap sesuai" tetap dihitung dari `received_at`, sama seperti sebelumnya.
7. Isi yang disegel tidak berubah. Menambah kunci ke isi kanonik akan membuat segel dokumen lama tampil tidak sah saat dihitung ulang.

## Akibat

- App mengirim waktu simpan di perangkat untuk setiap entri antrean dan menandai entri yang tidak terkirim pada percobaan pertama.
- Kedua pihak melihat waktu dicatat dan waktu terkirim dari jawaban dokumen yang sama.
