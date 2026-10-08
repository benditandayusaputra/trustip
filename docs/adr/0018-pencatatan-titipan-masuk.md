# 0018. Pencatatan Titipan Masuk

## Konteks

FR-DRP-01 sampai 05 dan BR-DRP-01 sampai 07 meminta toko mencatat Titipan Masuk yang menambah stok dalam transaksi yang sama, menyalin syarat Kesepakatan ke tiap batch, memberi nomor, memberi tahu penitip, dan disegel. R-RES-06 meminta kunci idempotensi untuk semua pembuatan dokumen, tetapi belum ada rute yang menerapkannya. Beberapa hal tidak ditetapkan spesifikasi: tempat fungsi FEFO yang dipakai beberapa modul, bentuk kode galat untuk produk di luar Kesepakatan, isi yang disegel untuk dokumen yang statusnya berubah, isi kolom `batch_code`, dan bulan yang dipakai di nomor dokumen.

## Keputusan

1. Fungsi murni stok (saldo, FEFO, tanda batch) tinggal di `kernel/stock`, sama seperti `kernel/pricing`, karena dipakai modul `drop`, `count`, dan `stock` yang tidak boleh saling mengimpor domain.
2. Modul `drop` menulis `drops`, `drop_items`, dan mutasi `RECEIVE` di `stock_movements`. Modul `count` nanti menulis mutasi `SALE` dan `RETURN`. Modul `stock` hanya membaca.
3. `drops` menyimpan `store_org_id` supaya nomor dan kunci idempotensi bisa unik per toko di basis data.
4. Kepala `Idempotency-Key` wajib berupa UUID. Tanpa kepala atau bukan UUID dibalas 400 `MALFORMED_REQUEST`. Server menyimpan kunci beserta SHA-256 bentuk kanonik isi permintaan. Kunci sama dengan isi sama mengembalikan dokumen semula dengan status 201 dan kepala `Idempotent-Replay: true`. Kunci sama dengan isi berbeda dibalas 422 `IDEMPOTENCY_KEY_REUSED`. Dua permintaan berkunci sama yang datang bersamaan diselesaikan oleh batasan unik: yang kalah membatalkan transaksinya lalu membaca dokumen pemenang.
5. Permintaan membawa `agreement_version` yang dilihat pengguna. Bila versi aktif sudah berbeda, dibalas 409 `AGREEMENT_CHANGED` dengan versi terbaru. Ini juga dipakai catatan luring (BR-PWA-06).
6. Produk yang tidak aktif di versi berjalan (BR-DRP-01, BR-AGR-07) ditolak sebagai galat isian `PRODUCT_NOT_IN_AGREEMENT` pada `items[i].product_id`, dengan pesan dari spesifikasi bagian 24. `COUNT_EXCEEDS_STOCK` memakai bentuk yang sama.
7. Tanggal kedaluwarsa yang kosong diisi server dari tanggal titip ditambah masa simpan produk (BR-PRD-02). Tanggal sebelum tanggal titip ditolak sebagai `EXPIRY_BEFORE_DROP`. Peringatan "kurang dari 2 hari" (BR-DRP-05) hanya tampil di app karena tidak menghalangi.
8. Tanggal titip dihitung dalam kalender Asia/Jakarta: paling awal tiga hari sebelum hari ini dan paling lambat 5 menit di depan jam server (BR-PWA-07). Syarat "tidak lebih awal dari Hitung Sisa terakhir" ditambahkan saat tabel Hitung Sisa ada.
9. `batch_code` adalah nomor dokumen ditambah nomor baris, misalnya `TM-202611-0007/2`. Nomor dokumen memakai bulan waktu terima di Asia/Jakarta.
10. Isi yang disegel adalah bagian dokumen yang tidak boleh berubah: nomor, kemitraan, versi Kesepakatan, tanggal titip, pencatat, dan baris beserta syarat salinannya. Status tidak ikut. Konfirmasi dan pembatalan menjadi entri segel baru atas isi yang sama, sehingga penetapan "dianggap sesuai" oleh tugas harian tidak membuat segel tidak sah.
11. `DROP_RECORDED` dikirim ke semua anggota aktif penitip karena tidak memuat nilai uang. Tautannya `/app/titipan/<id>`.

## Akibat

- Rute pembuatan dokumen berikutnya (Hitung Sisa, pembayaran) memakai pola idempotensi yang sama.
- Kemitraan dikunci oleh penyegel di transaksi pencatatan, sehingga Titipan Masuk dan Hitung Sisa pada satu kemitraan tidak berjalan bersamaan.
