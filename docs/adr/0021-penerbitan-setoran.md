# 0021. Penerbitan Setoran

## Konteks

FR-STL-01 sampai 04, BR-STL-01 sampai 03, dan BR-CNT-04 meminta Setoran diterbitkan sistem: bersama Hitung Sisa untuk jadwal `ON_COUNT`, dan setelah periode berakhir untuk `WEEKLY` dan `MONTHLY`. Periode tidak boleh beririsan untuk satu Kemitraan (BR-STL-02), dan Hitung Sisa tidak boleh masuk periode yang sudah disetor (BR-CNT-06). Beberapa hal tidak ditetapkan spesifikasi: batas periode untuk `ON_COUNT`, perlakuan saat jadwal berganti di tengah periode, pelaku pada segel dan log audit, dan bentuk baris Setoran saat satu produk punya syarat berbeda di beberapa batch.

## Keputusan

1. Periode disimpan sebagai `period_start` dan `period_end` bertipe `timestamptz` dengan batas setengah terbuka `[start, end)`. Batasan `EXCLUDE USING gist` dengan `btree_gist` menolak dua periode yang beririsan pada Kemitraan yang sama.
2. Periode mingguan dimulai Senin 00.00 dan bulanan tanggal 1 pukul 00.00, keduanya di Asia/Jakarta.
3. Setoran `ON_COUNT` mencakup semua alokasi penjualan yang belum tertaut sampai waktu hitung. `period_end` adalah waktu hitung ditambah satu detik, karena waktu hitung disimpan dalam detik penuh. `period_start` adalah akhir Setoran terakhir, atau waktu penjualan pertama bila belum ada Setoran.
4. Setoran berkala dimulai dari yang lebih akhir antara awal periode kalender dan akhir Setoran terakhir. Dengan begitu pergantian jadwal di tengah periode tidak membuat periode beririsan.
5. Hitung Sisa ditolak dengan galat isian `PERIOD_ALREADY_SETTLED` pada `counted_at` bila waktunya sebelum akhir Setoran terakhir. Pesannya menyebut nomor Setoran itu.
6. Penerbitan adalah peristiwa sistem. Entri segel `SETTLEMENT_ISSUED` dan log audit tidak punya pengguna dan organisasi pelaku, sehingga kolom `seals.actor_user_id`, `seals.actor_org_id`, dan `audit_logs.actor_id` boleh kosong. Bentuk kanonik entri segel memakai string kosong untuk keduanya.
7. Baris Setoran dikelompokkan per produk dan syarat salinan batch. Produk yang batch-nya punya syarat berbeda tampil sebagai beberapa baris dengan harga jual dan modelnya masing-masing. Total adalah penjumlahan alokasi, tidak dihitung ulang dari baris.
8. Isi yang disegel memuat nomor, kedua pihak, jadwal, periode, jatuh tempo, total, dan setiap alokasi beserta nilainya, diurutkan menurut ID. Status tidak ikut, sehingga pembayaran dicatat sebagai entri segel baru atas isi yang sama.
9. Modul `count` memanggil penerbit lewat port `SettlementIssuer` di dalam transaksi Hitung Sisa (R-API-02). Modul `settlement` yang memutuskan apakah jadwalnya `ON_COUNT`.
10. Pemberitahuan `SETTLEMENT_ISSUED` dikirim ke OWNER kedua pihak karena menyangkut uang.

## Akibat

- Dua Hitung Sisa dalam detik yang sama pada Kemitraan `ON_COUNT` tidak mungkin. Yang kedua ditolak `PERIOD_ALREADY_SETTLED`.
- Penjadwal di 7.3 memakai fungsi periode yang sama dan menerbitkan Setoran untuk setiap periode yang sudah lewat, sehingga tugas yang terlewat bisa menyusul.
