# 0025. Laporan penjualan

## Konteks

FR-RPT-02 sampai 04 dan BR-RPT-01 sampai 04 meminta laporan penjualan untuk OWNER toko dan penitip, dikelompokkan per mitra, per produk, atau per hari, dengan ringkasan yang angkanya berasal dari alokasi penjualan dan mutasi yang sama dengan Setoran. Spesifikasi bagian 14 mendefinisikan unit terjual, omzet, untung toko, hak penitip, unit dikembalikan, dan persen dikembalikan, tetapi tidak menetapkan arti "unit masuk" saat Titipan Masuk dibatalkan, pembulatan persen, dan bentuk rentang.

## Keputusan

1. Rentang adalah tanggal Asia/Jakarta dari `from` sampai `to`, keduanya termasuk, paling panjang 366 hari. Server menerjemahkannya menjadi batas setengah terbuka.
2. Angka uang dan unit terjual dijumlah dari `sale_allocations` menurut waktu hitung, sehingga total laporan sama dengan total Setoran pada rentang yang sama (BR-RPT-03). Unit dikembalikan dijumlah dari mutasi `RETURN`. Unit masuk adalah jumlah mutasi bersumber Titipan Masuk, yaitu `RECEIVE` dikurangi `ADJUSTMENT` pembatalannya.
3. Persen dikembalikan dikirim sebagai basis poin (R-RES-03): unit dikembalikan kali 10.000 dibagi unit masuk, dibulatkan setengah ke atas, dan 0 bila tidak ada unit masuk.
4. Tiga pengelompokan dipetakan ke tiga kueri bernama (R-SEC-58). Pengelompokan per hari mengisi setiap tanggal di rentang, termasuk yang tanpa penjualan, supaya grafik harian tidak bolong.
5. Lingkup data mengikuti pihak pelaku pada kemitraan: toko melihat semua titipan di tokonya dan penitip hanya produknya sendiri di tiap toko (BR-RPT-02). Hanya OWNER (`VIEW_REPORTS`).

6. Rekap harian (FR-RPT-01) hanya untuk OWNER toko, satu tanggal Asia/Jakarta dengan bawaan hari ini, satu baris per penitip. "Sudah dibayar" adalah hak penitip dari penjualan hari itu yang Setorannya sudah `PAID`.
7. Laporan Setoran (FR-RPT-06) menghitung Setoran yang terbit pada rentang per mitra: jumlah dan nilai yang terbit, lunas, belum lunas, dan lewat tempo. Lewat tempo memakai tanggal hari ini di Asia/Jakarta, sama dengan status turunan di ADR 0021.

## Akibat

- Rekap harian, laporan Setoran, grafik, dan ekspor CSV memakai definisi angka yang sama.
