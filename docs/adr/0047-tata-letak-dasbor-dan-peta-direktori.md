# ADR 0047: tata letak dasbor toko dan peta direktori

Tanggal: 10 Oktober 2026. Status: diterima.

## Konteks

Saran yang panjang di kolom samping `/app/toko` membuat tinggi dasbor timpang. Saringan `/cari/toko` langsung mengambil banyak ruang pada HP dan mengurangi ruang baca kartu di desktop. Pengguna meminta perbaikan kedua halaman, aksesibilitas, serta peta dengan penanda yang membuka detail toko. Permintaan ini mengizinkan pelaksanaan FR-DIR-13 lebih awal; fitur P1 lainnya tetap mengikuti gerbang proyek.

## Keputusan

Dasbor menempatkan Hitung cepat dan Perlu tindakan di baris awal, Saran selebar area isi, lalu Produk terlaris dan Peluang. Saran menampilkan tiga kartu awal. Daftar operasional dan peringkat menampilkan lima baris awal. Semua sisanya tersedia melalui tombol buka/ringkas yang menyatakan keadaan dan konten yang dikendalikan.

Pencarian teks dan urutan selalu terlihat. Saringan tambahan dibuka melalui tombol Filter pada HP maupun desktop, dengan penjelasan tujuan dan jumlah saringan aktif. Panel menggunakan kolom sesuai ruang yang tersedia, tanpa mengambil kolom hasil secara permanen. Pilihan Daftar/Peta memakai tombol dengan `aria-pressed`.

`DirectoryStoreCard.location` berisi koordinat `organizations.location_public` atau `null`, mengikuti daftar organisasi aktif dan terdaftar yang sudah ada. Lokasi privat tidak masuk kontrak ini. Hasil Daftar dan Peta menggunakan respons, saringan, serta pagination yang sama; tidak ada permintaan profil untuk setiap marker. Peta hanya menampilkan hasil yang sudah dimuat dan menjelaskan lokasi yang belum tersedia.

Leaflet tetap diimpor hanya oleh `MapView`, dan dimuat saat peta dibuka. Penanda memakai elemen DOM dan `textContent`; nama toko tidak dirender menjadi HTML mentah. Klik, Enter, atau Space memilih toko dan menampilkan kartu detail. Daftar nama toko menyediakan jalur alternatif untuk keyboard, marker yang bertumpuk, atau kegagalan ubin. Toko pada koordinat yang persis sama memakai satu marker berisi jumlah toko, sehingga marker tidak saling menutupi. Klik marker membuka toko pertama, dan daftar nama menyediakan pilihan toko lainnya. Peta mempertahankan atribusi OpenStreetMap dan tema terang/gelap.

Pola tombol buka/ringkas mengikuti [WAI-ARIA Disclosure](https://www.w3.org/WAI/ARIA/apg/patterns/disclosure/). Ukuran kontrol, warna, jarak, dan lebar kolom memakai token desain yang ada atau token baru pada `tokens.css`.

## Konsekuensi

Koordinat publik menjadi tambahan wajib yang boleh bernilai null pada kontrak daftar toko. Kode kueri dan tipe frontend dibangkitkan ulang. Ketika toko belum memiliki koordinat, toko tetap bisa dibaca pada Daftar. Pagination tetap perlu dimuat untuk memperluas cakupan peta; tampilan ini tidak mengklaim menunjukkan semua toko sekaligus.
