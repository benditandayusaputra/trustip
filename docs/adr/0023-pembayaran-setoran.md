# 0023. Pembayaran Setoran

## Konteks

FR-STL-05 sampai 08 dan BR-STL-05, 06, 08 meminta toko mencatat bayar dengan bukti, penitip menyatakan uang diterima atau menolak, penitip bisa menandai lunas sendiri, dan konfirmasi pembayaran berarti menyetujui Hitung Sisa sumbernya. R-SEC-35 meminta hash bukti bayar ikut disegel, dan R-SEC-72 meminta bukti bayar hanya bisa dibuka kedua pihak. Beberapa hal tidak ditetapkan spesifikasi: kolom tanggal bayar, bentuk kode galat untuk pembayaran yang masih menunggu, siapa yang boleh melihat QRIS penitip lewat API berkas, dan cara menyetujui Hitung Sisa sumber.

## Keputusan

1. Pembayaran tinggal di modul `settlement` karena tidak berdiri tanpa Setoran. Tabel `payments` menyimpan tanggal bayar sebagai `paid_on` bertipe `date`, karena isiannya tanggal bisnis Asia/Jakarta (R-RES-03), bukan waktu.
2. Satu Setoran hanya boleh punya satu pembayaran yang menunggu atau diterima, dijaga indeks unik sebagian di basis data. Pencatatan kedua saat masih menunggu dibalas 409 `PAYMENT_PENDING`, dan pencatatan atas Setoran yang sudah lunas dibalas 409 `SETTLEMENT_ALREADY_PAID`. Pembayaran yang ditolak tidak menghalangi pencatatan ulang.
3. `PAYMENT_AMOUNT_MISMATCH` dan `PROOF_REQUIRED` adalah galat isian pada `amount` dan `proof_file_id`, dengan pesan dari spesifikasi bagian 24. Bukti adalah berkas berjenis `PAYMENT_PROOF` milik toko itu sendiri. Bukti untuk transfer dan QRIS juga diwajibkan batasan basis data.
4. Pencatatan memakai kunci idempotensi per toko dengan pola yang sama seperti ADR 0018.
5. Entri segel atas dokumen `PAYMENT` dengan nomor Setoran: `PAYMENT_RECORDED` saat dicatat, lalu `PAYMENT_CONFIRMED` atau `PAYMENT_REJECTED`. Isinya memuat jumlah, metode, tanggal, pencatat, dan SHA-256 bukti, sehingga bukti yang ditukar membuat segel tidak sah.
6. "Tandai lunas" oleh penitip (FR-STL-07) membuat pembayaran tunai yang langsung `CONFIRMED` atas nama penitip, dengan tanggal hari ini, lalu menjadikan Setoran `PAID`.
7. Saat pembayaran diterima, Hitung Sisa sumber yang masih `RECORDED` ditetapkan `CONFIRMED` atas nama penitip yang menerima (BR-STL-08). Modul `settlement` memintanya lewat port `CountApprover` yang dipenuhi repositori modul `count`. Tidak ada entri segel `COUNT_CONFIRMED` tambahan karena penerimaan pembayaran sudah disegel.
8. API berkas mengizinkan pembacaan berkas milik organisasi lain hanya dalam dua keadaan: bukti bayar pada Setoran tempat pelaku menjadi pihak, dan QRIS penitip untuk toko yang punya kemitraan aktif dengannya. Halaman Setoran menampilkan info pembayaran penitip hanya kepada penitip itu dan toko pada kemitraan aktif.
9. Pemberitahuan `PAYMENT_RECORDED`, `PAYMENT_CONFIRMED`, dan `PAYMENT_REJECTED` hanya untuk OWNER karena memuat nilai uang. Di P0 hanya OWNER toko yang boleh mencatat bayar, sehingga pencatat selalu termasuk OWNER yang diberi tahu.

## Akibat

- Pembayaran sebagian (FR-STL-11) di P1 perlu melonggarkan indeks unik dan aturan jumlah.
