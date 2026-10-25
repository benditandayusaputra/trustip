# 0056: Grafik laporan tanpa pustaka

Tanggal: 25 Oktober 2026

## Konteks

Laporan dan dasbor hanya punya satu grafik area omzet harian. Pengguna meminta grafik laporan yang lebih lengkap untuk video lomba. Grafik area juga terlihat anjlok ke nol pada hari ini yang belum ada penjualannya. Aturan R-CC-52 meminta pustaka terkecil yang memenuhi kebutuhan, ADR 0004 melarang atribut `style` milik sendiri, dan lapisan `components/ui` tidak boleh mengimpor `domain`.

## Keputusan

1. Grafik dibuat sebagai SVG di fitur laporan tanpa pustaka baru: kolom untuk omzet harian, donat untuk pembagian omzet, dan batang mendatar yang bisa bertumpuk untuk omzet per mitra serta terjual dibanding dikembalikan per produk.
2. Geometri dihitung fungsi murni di `lib/domain/chart.ts` (`columnChart`, `donutSlices`, `stackedBars`) dan diuji. Posisi memakai atribut presentasi SVG dalam persen, bukan atribut `style`, sehingga CSP tidak berubah.
3. Grafik kolom menggantikan grafik area, sehingga `AreaChart` dan `areaShape` dihapus. Hari tanpa penjualan tampil sebagai kolom kosong.
4. Tab Penjualan memuat laporan per mitra dan per produk untuk rentang yang sama dan menampilkan empat grafik. Dasbor memakai donat Pembagian omzet dari ringkasan 30 hari yang sudah dimuat. Dasbor toko juga menampilkan grafik penjualan harian.
5. Komponen grafik dimuat malas seperti grafik sebelumnya (R-APP-11). Nilai tetap tertulis sebagai teks di samping batang dan di legenda donat, sehingga warna bukan satu-satunya pembawa informasi.

## Akibat

Tab Penjualan membuat dua permintaan tambahan ke rute laporan yang sama. Grafik harian tetap punya alternatif tabel. Dasbor toko kini memuat satu permintaan laporan harian lagi untuk OWNER.
