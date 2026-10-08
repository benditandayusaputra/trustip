# 0038. Mesin saran dan dasbor

## Konteks

Spesifikasi bagian 15 meminta dasbor penitip dan toko serta mesin saran sebagai fungsi murni dengan tingkat, penjelasan, dan satu aksi. Beberapa hal tidak ditetapkan: aksi INS-01 untuk penitip yang tidak boleh mengisi Hitung Sisa, cara menghitung perkiraan unit tak laku dan batch mana yang dinilai, asal kandidat INS-10, pembagian angka antara OWNER dan STAFF, serta jendela 14 hari untuk rata-rata harian.

## Keputusan

1. Mesin saran ada di modul `insight`. `Generate(snapshot, now)` murni, terurut menurut tingkat lalu nilai rupiah terdampak, dan rute `GET /api/v1/insights` mengembalikan lima teratas. Teks judul dan penjelasan disusun app dari jenis dan angka saran supaya tetap di katalog teks.
2. Aksi bergantung pada sisi pelaku. Untuk INS-01 dan INS-02, toko membuka Hitung Sisa sedangkan penitip diarahkan mengingatkan toko, karena tabel peran bagian 1 melarang penitip mengisi Hitung Sisa. Aksi INS-01 membuka Hitung Sisa dengan pengembalian produk itu terisi sebagai `EXPIRED`, paling banyak sebesar sisa.
3. INS-01 dan INS-02 dinilai per batch bersaldo. Perkiraan unit tak laku memakai bilangan rasional `sisa batch − terjual 14 hari × hari tersisa / hari data`, dibulatkan ke atas, sehingga tidak ada angka pecahan. Batch milik produk yang sudah tidak aktif di Kesepakatan tetap dinilai, dengan harga jual dari batch terakhir.
4. Jendela 14 hari dihitung sebagai 14 hari kalender Asia/Jakarta termasuk hari ini. "Belum dihitung" (INS-07) dihitung dari Hitung Sisa terakhir, atau dari titipan pertama bila belum pernah dihitung.
5. Kandidat INS-10 diambil dari Cari Mitra dengan urutan kecocokan terbaik lewat port `CandidateFinder`, lalu ditandai bila ada Ajuan terbuka dengan organisasi itu. Aksinya membuka Cari Mitra dengan nama organisasi itu sampai profil publik tersedia di 11.3.
6. Saran jumlah titip (INS-09) memakai hari yang sama dalam pekan dengan hari ini. Satu kejadian adalah satu tanggal Hitung Sisa; bila sehari ada beberapa hitungan, terjualnya dijumlah dan dianggap habis bila salah satunya bersisa 0.
7. `GET /api/v1/dashboard` boleh dibaca semua anggota karena isinya nilai dokumen (stok, Setoran terbuka, daftar perlu tindakan, Hitung cepat). Omzet, untung toko, hak penitip 30 hari, grafik, dan peringkat diambil dari laporan penjualan yang hanya untuk OWNER (BR-RPT-01). Grafik dasbor penitip memakai saringan `partner_id` dan `product_id` yang ditambahkan ke laporan penjualan.

## Akibat

- Dasbor penitip memuat lima permintaan paralel untuk OWNER dan dua untuk STAFF. Ukuran LCP pada data contoh dicatat di 12.3.
- Kebutuhan yang habis masa tayang tiga hari lagi memakai ambang yang sama dengan pengingat BR-NED-03.
