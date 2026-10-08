# 0036. Kebutuhan Toko

## Konteks

Bagian 5 spesifikasi meminta toko memasang Kebutuhan dengan masa tayang 30 hari yang bisa diperpanjang, paling banyak 5 berstatus `OPEN` (BR-NED-01), tanpa nomor telepon dan tautan di teks (BR-NED-02), dan otomatis `EXPIRED` setelah masa tayang lewat (BR-NED-03). Model data menulis batas 5 ditegakkan di layanan, sedangkan R-SEC-44 meminta batas bisnis ditegakkan di basis data. Spesifikasi tidak menetapkan apa yang boleh diperpanjang, bentuk penolakan kontak di teks, dan kode galatnya.

## Keputusan

1. Modul baru `need` dengan tabel `needs`. Batas 5 Kebutuhan terbuka dijaga pemicu basis data yang mengunci baris organisasi lalu menghitung, sehingga dua pemasangan bersamaan tetap tidak melewati batas. Layanan menerjemahkan penolakan pemicu menjadi `NEED_LIMIT_REACHED` (409), memenuhi model data dan R-SEC-44 sekaligus.
2. Teks judul dan deskripsi ditolak dengan galat isian `CONTACT_NOT_ALLOWED` bila memuat deret 9 angka atau lebih (dengan spasi, titik, tanda hubung, atau kurung) atau tautan (`http`, `www.`, `wa.me`, atau nama domain umum). Angka biasa seperti jumlah dan harga tetap boleh.
3. Perpanjangan berlaku untuk Kebutuhan `OPEN` atau `EXPIRED` dan menjadikan masa tayang 30 hari dari waktu perpanjangan. Membuka lagi Kebutuhan yang kedaluwarsa ikut diperiksa batas 5. Terpenuhi dan tutup hanya untuk Kebutuhan yang masih terbuka. Selain itu dibalas `NEED_NOT_OPEN` (409) dengan pesan bagian 5.
4. Status yang tampil dihitung dari masa tayang, sehingga Kebutuhan yang sudah lewat tampil `EXPIRED` walau tugas harian 00.05 (spesifikasi 21.5) belum menandainya.
5. `proposals.need_id` ditambahkan supaya rincian Kebutuhan bisa menampilkan tanggapan penitip. Tanggapan dibuat lewat Ajuan di 9.7. Menerima Ajuan tidak mengubah status Kebutuhan (BR-NED-04).

## Akibat

- Pengingat 3 hari sebelum berakhir dan pemberitahuan ke penitip yang cocok ditambahkan bersama Papan Kebutuhan di 9.7.
