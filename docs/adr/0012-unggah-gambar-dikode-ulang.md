# 0012. Unggah gambar: kode ulang di peramban dan di server

## Konteks

BR-ORG-06 dan NFR-09 mengizinkan JPG, PNG, dan WebP, dikecilkan di peramban ke lebar 1280 piksel dan paling besar 1 MB. R-SEC-70 meminta server menentukan jenis dari isi, menolak sisi di atas 4096 piksel, dan mengodekan ulang sebelum menyimpan. Pustaka standar Go bisa mendekode dan mengodekan JPEG dan PNG, tetapi tidak WebP. R-SEC-71 dan 72 memisahkan sajian berkas publik dan privat, dan R-SEC-26 meminta berkas publik dilayani tanpa proxy hanya untuk GET atas kunci acak.

## Keputusan

1. Peramban menerima JPG, PNG, dan WebP, lalu selalu mengodekannya ulang lewat kanvas: PNG bila ada piksel transparan, selain itu JPEG dengan kualitas yang diturunkan sampai muat 1 MB (`app/src/lib/api/shrink-image.ts`). Nama berkas asli tidak dikirim.
2. Server hanya menerima JPEG dan PNG. Ia membaca dimensi lebih dulu, menolak sisi di atas 4096 piksel sebelum dekode penuh, lalu mengodekan ulang ke format yang sama (`platform/imaging`). Metadata EXIF, termasuk lokasi, ikut terbuang. Tidak ada dependensi dekoder WebP.
3. Berkas disimpan di luar basis data di `FILE_STORAGE_DIR`, dipisah folder `public/` dan `private/`. Nama berkas adalah kunci acak 128 bit dari server ditambah ekstensi (R-SEC-54). Tabel `files` mencatat pemilik, tujuan, ukuran, dimensi, dan hash SHA-256.
4. Tujuan unggahan menentukan sifatnya: `LOGO` dan `STORE_PHOTO` publik, `PAYMENT_QR` privat. `STORE_PHOTO` hanya untuk toko dan `PAYMENT_QR` hanya untuk penitip.
5. Berkas publik dilayani `GET /files/{kunci}` dari asal API tanpa tanda tangan tepi dan tanpa menyentuh basis data, hanya dari folder `public/`, dengan `immutable` setahun dan `Cross-Origin-Resource-Policy: cross-origin`.
6. Berkas privat dilayani `GET /api/v1/files/{id}` lewat proxy, hanya untuk anggota organisasi pemilik, dengan `no-store`. Akses pihak lawan kemitraan ditambahkan bersama fitur yang membutuhkannya.

## Akibat

- Klien selain app yang mengirim WebP langsung ke API ditolak `INVALID_IMAGE`. Itu tidak terjadi lewat app karena peramban sudah mengubahnya.
- Berkas yang tersimpan di disk tetapi gagal dicatat di basis data tertinggal sebagai sampah tanpa rujukan dan tanpa jalan publik ke berkas privat.
