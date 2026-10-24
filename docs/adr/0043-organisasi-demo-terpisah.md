# 0043. Organisasi demo terpisah dari organisasi nyata

## Konteks

Spesifikasi bagian 25 meminta data contoh dengan akun demo dan 20 organisasi direktori, dan data organisasi demo disetel ulang tiap malam. BR-SIG-04 dan R-SEC-31 hanya mengizinkan hapus segel pada kemitraan yang **kedua** pihaknya organisasi demo. Basis data aplikasi di Neon juga dipakai untuk uji ujung ke ujung (ADR 0009), sehingga terus berisi organisasi uji.

Bila organisasi nyata bisa bermitra dengan organisasi demo, segel kemitraan campuran tidak bisa dihapus dan setel ulang malam macet. Bila akun demo melihat organisasi nyata, Cari Mitra di demo penuh organisasi uji.

## Keputusan

1. Organisasi demo hanya melihat dan menjangkau organisasi demo, organisasi nyata hanya melihat dan menjangkau organisasi nyata. Pengunjung tanpa masuk melihat keduanya, karena melihat tidak membuat hubungan dagang dan data live sebagian besar berisi organisasi demo. Mengajukan kerja sama tetap harus masuk lebih dulu.
2. Aturan itu ditegakkan di SQL dengan satu predikat yang sama: `o.is_demo` harus sama dengan `is_demo` milik pemirsa, kecuali pemirsa adalah pengunjung (ID nol). Tempatnya: daftar dan hitungan Cari Mitra (toko, penitip, produk), Papan Kebutuhan, pihak Ajuan langsung, klaim undangan, dan penitip yang diberi tahu tentang Kebutuhan baru.
3. Target di luar lingkup dibalas 404, sama dengan organisasi yang tidak terdaftar.
4. Halaman publik (profil, produk, Kebutuhan, verifikasi segel) tidak disaring, karena tidak membuat hubungan dagang.

## Akibat

- Kemitraan campuran tidak bisa terjadi, sehingga setel ulang malam selalu bisa menghapus seluruh data dagang demo.
- Juri yang mendaftar akun sendiri tidak melihat organisasi demo di Cari Mitra. Untuk mencoba alur dagang, juri memakai tombol "Coba akun demo".
