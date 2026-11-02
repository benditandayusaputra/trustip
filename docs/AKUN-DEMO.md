# Akun demo

Semua toko, penitip, dan produk yang tampil di Cari Mitra (`/cari/toko`, `/cari/penitip`) dan Kebutuhan Toko adalah organisasi demo yang punya akun sungguhan. Masuk lewat [halaman masuk](https://trustip.vercel.app/masuk) memakai email di bawah.

**Kata sandi semua akun: `Coba-Demo-2026`**

Data demo disetel ulang setiap hari pukul 03.00 WIB, jadi perubahan saat mencoba akan kembali seperti semula keesokan harinya. Organisasi demo hanya bisa bermitra dengan organisasi demo lain (ADR 0043).

## Akun dengan riwayat transaksi

Delapan usaha ini sudah saling bermitra dan punya riwayat 30 hari: titipan, hitung sisa, setoran, dan pembayaran.

| Pemilik | Usaha | Jenis | Email | Yang bisa dicoba |
|---|---|---|---|---|
| Bendi | Roti Mas Bendi | Penitip | `penitip@demo.trustip.id` | Mitra dengan Warung Mbak Zara, Kantin Kampus, Kedai Kopi Sore. Terima Ajuan dari Minimarket Berkah, konfirmasi pembayaran |
| Rina Kusumawati | Keripik Bu Rina | Penitip | `keripik.rina@demo.trustip.id` | Mitra dengan empat toko, setoran mingguan |
| Hidayah Lestari | Dapur Sambal Mak Ijah | Penitip | `sambal.ijah@demo.trustip.id` | Mitra dengan Warung Mbak Zara dan Kedai Kopi Sore, setoran bulanan |
| Zara | Warung Mbak Zara | Toko | `toko@demo.trustip.id` | Catat titipan, Hitung Sisa titipan kemarin, catat bayar setoran |
| Bambang Sutrisno | Kantin Kampus | Toko | `kantin.kampus@demo.trustip.id` | Mitra Roti Mas Bendi dan Keripik Bu Rina, Kebutuhan lauk siap saji |
| Dedi Kurniawan | Toko Oleh-oleh Pak Dedi | Toko | `oleh.dedi@demo.trustip.id` | Mitra Keripik Bu Rina, contoh Ajuan yang ditolak |
| Yoga Pratama | Kedai Kopi Sore | Toko | `kedai.sore@demo.trustip.id` | Mitra Roti Mas Bendi dan Dapur Sambal Mak Ijah, Kebutuhan kue kering |
| Jessica | Minimarket Berkah | Toko | `minimarket@demo.trustip.id` | Mitra Keripik Bu Rina, Ajuan ke Roti Mas Bendi menunggu jawaban |

## Akun tanpa mitra

Usaha ini tampil di Cari Mitra tetapi belum bermitra dengan siapa pun. Cocok untuk mencoba alur dari awal: mencari mitra, mengajukan kerja sama, lalu menyepakati harga.

### Penitip

| Pemilik | Usaha | Email | Produk |
|---|---|---|---|
| Bayu Setiawan | Kopi Bubuk Arjuno | `d.arjuno@demo.trustip.id` | Kopi Bubuk Robusta (Ajuan ke Warung Mbak Zara menunggu jawaban) |
| Lastri Handayani | Kue Kering Bu Lastri | `d.lastri@demo.trustip.id` | Nastar Nanas |
| Yuni Astuti | Tahu Bakso Mbak Yuni | `d.yuni@demo.trustip.id` | Tahu Bakso Beku isi 10 |
| Joko Santoso | Keripik Apel Malang Jaya | `d.apel@demo.trustip.id` | Keripik Apel |
| Dewi Anggraini | Sambal Roa Bu Dewi | `d.roa@demo.trustip.id` | Sambal Roa |
| Sumiati | Lauk Rumahan Bu Sum | `d.sum@demo.trustip.id` | Ayam Bumbu Bali |
| Endang Purwati | Rajutan Ibu-ibu Merjosari | `d.rajut@demo.trustip.id` | Gantungan Kunci Rajut |
| Sunarti | Jamu Segar Mbok Narti | `d.jamu@demo.trustip.id` | Jamu Kunyit Asam |

### Toko

| Pemilik | Usaha | Email | Catatan |
|---|---|---|---|
| Harini Puspita | Warung Makan Bu Har | `d.har@demo.trustip.id` | |
| Agus Salim | Kantin Gedung Serbaguna | `d.gedung@demo.trustip.id` | |
| Rizky Ananda | Kafe Senja Dinoyo | `d.senja@demo.trustip.id` | |
| Nur Hidayati | Koperasi Mahasiswa Sejahtera | `d.kopma@demo.trustip.id` | Kebutuhan oleh-oleh khas Malang |
| Slamet Riyadi | Toko Kelontong Barokah | `d.barokah@demo.trustip.id` | Kebutuhan makanan beku siap goreng |
| Lina Marlina | Toko Roti dan Kue Sumbersari | `d.sumbersari@demo.trustip.id` | |
| Darmaji | Kios Pasar Dinoyo | `d.dinoyo@demo.trustip.id` | |
| Eko Wahyudi | Minimarket Sahabat | `d.sahabat@demo.trustip.id` | |
| Wulan Sari | Pusat Oleh-oleh Batu Asri | `d.batu@demo.trustip.id` | Kota Batu |
| Yuliani | Kantin Rumah Sakit Sehat | `d.rssehat@demo.trustip.id` | |
| Fajar Nugroho | Warung Kopi Tlogomas | `d.tlogomas@demo.trustip.id` | |
| Hendra Saputra | Toko Serba Ada Singosari | `d.singosari@demo.trustip.id` | Kabupaten Malang |

Daftar ini mengikuti `api/cmd/server/demo_world.json`. Bila data demo diubah, perbarui juga berkas ini.
