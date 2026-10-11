# Verifikasi halaman akun dan logo aplikasi

Tanggal: 11 Oktober 2026.

## Cakupan

Masuk dan registrasi memakai ilustrasi titip jual dengan biru utama dan aksen jingga, kartu formulir, pilihan tema, serta tautan kebijakan. Desktop memakai dua kolom; pada HP ilustrasi tampil ringkas di atas formulir. Akun contoh hanya ada pada `/demo`, yang dibuka dari ajakan demo beranda.

Logo dua mitra biru dan paket jingga yang diberikan pengguna dipakai pada merek, favicon, ikon PWA, dan ikon Apple. Ekspor tersedia pada 64, 180, 192, dan 512 px. Varian maskable memiliki latar dan ruang aman. Bentuk dan warna logo tidak digambar ulang.

## Pemeriksaan

- `make check`: format, tipe, lint, batas kode, arsitektur, token desain, rahasia, dan keamanan dependensi lulus.
- `make test`: seluruh paket Go dengan deteksi race lulus; 84 berkas dan 502 uji frontend lulus.
- `make e2e`: 73 uji browser lulus, termasuk alur transaksi, keamanan, performa, audit aksesibilitas seluruh halaman pada kedua tema, serta navigasi keyboard.
- Uji formulir mempertahankan validasi, fokus kesalahan, captcha, keadaan luring, pesan kegagalan, serta pengiriman dan keberhasilan registrasi/masuk.
- Uji kata sandi memastikan tampil/sembunyi mempertahankan nilai, autocomplete, fokus, dan hubungan aksesibilitas kontrol.
- Empat uji browser khusus lulus pada HP dan desktop: akun contoh tidak tampil pada masuk/registrasi, navigasi menjaga tujuan lanjutan, kontrol sandi dapat dipakai lewat Enter/Space, dan demo dari beranda mengisi hanya formulirnya sendiri.
- Uji alur keyboard lulus: dari `/demo`, memilih toko, masuk, dan membuka Hitung Sisa tanpa mouse.
- Audit kedua tema menunggu transisi warna selesai sebelum mengukur kontras. Pemeriksaan tidak mengabaikan aturan aksesibilitas.

## Tampilan produksi

Galeri mencakup 30 tangkapan layar: masuk, registrasi, dan demo pada lebar 320, 390, 768, 1024, dan 1440 px, dalam tema terang dan gelap. Tidak ditemukan overflow horizontal halaman atau galat JavaScript. Lihat [galeri halaman akun](../../output/auth-layout/index.html) dan [hasil pemeriksaan](../../output/auth-layout/verification.json).

Tidak ada perubahan kontrak API, aturan autentikasi server, atau dependensi. Pemeriksaan otomatis aksesibilitas tidak menggantikan evaluasi langsung dengan pengguna pembaca layar.
