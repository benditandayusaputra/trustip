# 0022. Penjadwal di dalam proses

## Konteks

Spesifikasi bagian 21.5 meminta tugas terjadwal di dalam proses API: penetapan "dianggap sesuai" dan penanda kedaluwarsa setiap hari pukul 00.05, Setoran mingguan setiap Senin pukul 00.10, dan Setoran bulanan setiap tanggal 1 pukul 00.10, semuanya WIB. Basis data harus bisa tidur di luar waktu itu, tugas harus aman diulang, dan R-CC-72 meminta setiap goroutine berhenti saat server berhenti. Gerbang Tahap 7 meminta Setoran berkala dibuktikan lewat penjadwal dengan jam tiruan.

## Keputusan

1. Paket `platform/scheduler` menghitung waktu jalan berikutnya dari jam di memori (`NextRun`), menunggu lewat port `Waits`, lalu menjalankan semua tugas yang jatuh pada menit yang sama. Basis data hanya disentuh oleh tugasnya sendiri. Uji memakai `Waits` tiruan yang memajukan jam manual.
2. Penjadwal tidak menyusul waktu yang terlewat saat server mati. Yang menyusul adalah tugasnya: Setoran berkala menerbitkan semua periode kalender yang sudah lewat dan masih punya penjualan tanpa Setoran, dan penetapan "dianggap sesuai" memakai batas waktu, bukan tanggal tertentu.
3. Penetapan "dianggap sesuai" untuk Titipan Masuk dan Hitung Sisa mengubah status menjadi `CONFIRMED` dengan `auto_confirmed = true` dan `confirmed_at` sama dengan waktu terima ditambah 48 jam, sama dengan yang sudah ditampilkan saat dibaca. Tidak ada entri segel baru, sesuai ADR 0018 butir 10.
4. Ajuan `OPEN` yang `expires_at`-nya lewat ditandai `EXPIRED`. Penanda Kebutuhan Toko menyusul bersama tabelnya di Tahap 9.
5. Setoran berkala diterbitkan per Kemitraan dalam transaksinya sendiri dengan mengunci baris Kemitraan lebih dulu, sehingga tidak berjalan bersamaan dengan Hitung Sisa pada Kemitraan yang sama. Kemitraan yang jadwal aktifnya sudah berganti dilewati. Kegagalan satu Kemitraan tidak menghentikan yang lain dan dicatat di log.
6. Kasus pakai sistem tidak menerima `Actor`, sehingga tidak memanggil `policy.Require`. Uji arsitektur R-SEC-36 hanya memeriksa kasus pakai yang bertindak atas nama pengguna.

## Akibat

- Satu proses API menjalankan penjadwal. Bila kelak ada lebih dari satu proses, tugas tetap aman karena penerbitan dikunci per Kemitraan dan dijaga batasan periode, tetapi log akan mencatat kegagalan dari proses yang kalah.
