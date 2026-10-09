# 0044. Data contoh dan akun demo

## Konteks

Spesifikasi bagian 25 meminta data contoh 30 hari di satu kecamatan, dua akun demo, tombol "Coba akun demo", dan setel ulang tiap malam (FR-AUTH-07, tugas 03.00 di bagian 21.5). Tim memilih Kecamatan Lowokwaru, Kota Malang (PRD bagian 18 butir 5). Data contoh harus lolos aturan yang sama dengan data asli: segel sah, stok cocok, Setoran cocok dengan Hitung Sisa.

## Keputusan

1. Dunia demo ditulis sebagai data di `api/cmd/server/demo_world.json`: 28 organisasi, 15 produk, 9 kemitraan (8 aktif, 1 berakhir), 6 Kebutuhan, 3 Ajuan, dan irama titip, hitung, serta bayar tiap kemitraan.
2. Riwayat dibangun dengan menjalankan kasus pakai yang sama dengan API, di atas jam manual yang dimajukan dari 35 hari lalu sampai kemarin. Penjadwal Setoran mingguan dan bulanan, tugas harian, dan ringkasan reputasi ikut dijalankan pada waktunya. Tidak ada baris dagang yang ditulis langsung dengan SQL, sehingga segel, stok, dan Setoran sah dengan sendirinya.
3. Irama dihitung mundur dari hari ini, bukan dari hari dalam pekan, supaya keadaan akhir selalu memicu saran yang sama: batch kedaluwarsa dan hampir kedaluwarsa, stok segera habis, belum dihitung, Setoran lewat tempo, saran jumlah titip, dan mitra yang cocok.
4. Setel ulang memanggil fungsi basis data `reset_demo_data()` (`SECURITY DEFINER`) yang hanya menghapus baris milik organisasi `is_demo`. Peran aplikasi tetap tidak punya hak hapus langsung pada `seals` dan `audit_logs` (R-SEC-60), dan pemicu segel tetap hanya mengizinkan hapus pada kemitraan antar organisasi demo (R-SEC-31).
5. Tugas `demo-reset` berjalan di penjadwal API setiap hari pukul 03.00 Asia/Jakarta. Satu kali isi butuh beberapa menit, dan diulang sampai tiga kali bila koneksi basis data putus. `make demo` dan perintah `trustip-api demo` menjalankan hal yang sama sekali jalan.
6. Tombol "Coba akun demo" di halaman masuk menampilkan email dan kata sandi kedua akun dan mengisinya ke formulir masuk. Masuk tetap lewat rute masuk biasa dengan captcha dan batas lajunya, sehingga tidak ada jalur masuk tanpa kata sandi.

## Akibat

- Semua pemberitahuan sebelum kemarin ditandai sudah dibaca, supaya kotak pemberitahuan demo tidak penuh.
- Selama setel ulang berjalan (beberapa menit setelah 03.00), akun demo tidak bisa dipakai.
- Batas laju akun demo di tabel 7.9 tercakup batas laju rute masuk (5 per menit dan 20 per jam per IP).
