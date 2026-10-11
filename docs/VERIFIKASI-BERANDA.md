# Verifikasi beranda, alur, dan footer

Tanggal: 11 Oktober 2026.

## Cakupan

Beranda menjelaskan platform titip jual untuk UMKM pada bagian pembuka, disertai ilustrasi penitip bersama toko, ajakan daftar/demo, dan tautan menuju cara kerja. Kartu peran menjelaskan manfaat bagi pemilik produk dan toko. Lima langkah bernomor menjelaskan siapa yang mencari mitra, menyepakati harga, mencatat titipan, menghitung sisa, dan mengonfirmasi setoran.

Contoh 20 roti, sisa 2, dan terjual 18 ditempatkan setelah alur. Harga jual Rp2.000 dan harga setor Rp1.750 per unit menjelaskan asal omzet Rp36.000, untung toko Rp4.500, dan setoran Rp31.500. Nilai tersebut diberi label contoh. Pembayaran dilakukan langsung antar mitra dan dicatat di aplikasi.

Footer publik memakai identitas aplikasi serta kelompok Jelajahi, Mulai, dan Informasi, dengan tautan direktori toko/penitip, Kebutuhan, cara kerja, pendaftaran, demo, privasi, dan ketentuan. Tidak ada kontak atau tautan sosial rekaan. Footer digunakan pada kerangka publik untuk beranda, profil, kebijakan, dan halaman luring. Direktori mempertahankan footer ringkas selama hasil dimuat setelah halaman terbuka.

Ilustrasi beranda dan akun memakai satu komponen vektor dengan token tema dan pilihan ukuran. Tampilan akun tetap ringkas pada HP. Tidak ada perubahan kontrak API, aturan bisnis, autentikasi, atau dependensi.

## Pemeriksaan

- `make check`: pemeriksaan format, tipe, lint, arsitektur, batas kode, token desain, rahasia, dan keamanan dependensi lulus.
- `make test`: seluruh paket Go dengan deteksi race lulus; 84 berkas dan 502 uji frontend lulus.
- `make e2e`: seluruh 73 uji browser lulus, mencakup desktop/HP, transaksi, layanan uji terpisah, keamanan, anggaran performa, kedua tema, dan navigasi keyboard.
- Uji komponen beranda, kerangka, serta formulir masuk dan registrasi lulus.
- Uji keyboard khusus lulus: menunggu formulir demo aktif, memilih peran toko, memastikan email dan sandi terisi, masuk, lalu membuka Hitung Sisa tanpa mouse. Captcha tetap wajib; uji tidak mengabaikan pemeriksaan keamanan.
- Audit beranda pada lima ukuran layar dalam kedua tema tidak menemukan pelanggaran aturan aksesibilitas otomatis.
- Pengukuran Lighthouse HP khusus lulus: beranda mendapat skor 100, LCP 1,45 detik, dan CLS 0,00015. Cari Toko mendapat skor 96 dengan CLS 0,08782 setelah mempertahankan footer ringkas. Angka tersebut adalah hasil pengukuran lokal dengan batas proyek, bukan data penggunaan produksi.

## Tampilan responsif

Galeri mencakup 30 tangkapan layar beranda, Cari Toko pengunjung, dan masuk pada lebar 320, 390, 768, 1024, dan 1440 px dalam tema terang/gelap. Tidak ditemukan overflow horizontal atau galat JavaScript. Bagian pembuka, alur, dan footer HP juga diperiksa secara langsung.

Lihat [galeri beranda](../../output/home-layout/index.html) dan [hasil pemeriksaan](../../output/home-layout/verification.json). Audit otomatis tidak menggantikan evaluasi langsung dengan pengguna.
