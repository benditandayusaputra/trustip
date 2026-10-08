# 0031. Ajakan pasang dan penyimpanan persisten

## Konteks

FR-PWA-03 meminta ajakan pasang yang muncul setelah pengguna menyelesaikan satu aksi, dan petunjuk "Tambahkan ke Layar Utama" di iOS. R-OFF-15 meminta permintaan penyimpanan persisten setelah catatan luring pertama dan peringatan mencolok untuk catatan yang menunggu lebih dari 24 jam. Spesifikasi tidak menyebut aksi apa yang dihitung dan di mana tanda "sudah ditolak" disimpan.

## Keputusan

1. Aksi yang dihitung: Titipan Masuk atau Hitung Sisa tersimpan (terkirim maupun masuk antrean), dan penitip menyatakan Titipan Masuk atau Hitung Sisa sesuai. Itulah aksi utama kedua tipe organisasi.
2. Tanda "aksi pertama sudah ada" dan "ajakan sudah ditolak" disimpan di `localStorage` perangkat. Isinya hanya tanda, tanpa data pengguna. Bila `localStorage` tidak bisa dipakai, tanda berlaku sampai halaman ditutup.
3. Di peramban yang mengirim `beforeinstallprompt`, ajakan memakai dialog pasang bawaan peramban. Di iPhone dan iPad, ajakan berupa petunjuk tombol Bagikan lalu "Tambahkan ke Layar Utama". Ajakan tidak muncul bila app sudah berjalan sebagai aplikasi terpasang. Menekan "Pasang" atau "Nanti" menyembunyikannya untuk seterusnya di perangkat itu.
4. Penyimpanan persisten diminta lewat `navigator.storage.persist()` setiap kali sebuah catatan masuk antrean tanpa langsung terkirim. Peramban sendiri yang memutuskan dan mengingat jawabannya.
5. Catatan yang menunggu lebih dari 24 jam memunculkan peringatan di halaman Antrean dan mengubah tanda Antrean di kepala app menjadi warna bahaya.

## Akibat

- Pengguna yang belum pernah mencatat atau mengonfirmasi tidak pernah melihat ajakan pasang.
