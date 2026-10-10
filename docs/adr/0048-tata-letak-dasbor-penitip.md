# ADR 0048: tata letak dasbor penitip

Tanggal: 10 Oktober 2026. Status: diterima.

## Konteks

Pengguna meminta perapian `/app/penitip`. Saran pada kolom samping yang sempit membentuk daftar panjang dan mendorong peringkat ke bawah, meninggalkan ruang kosong di bawah grafik. Peluang pada baris dua kolom hanya mengisi satu kolom.

## Keputusan

Empat angka kunci tetap mengikuti DSH 15.1. Pada desktop, grafik menempati delapan kolom dan Perlu tindakan empat kolom. Pada HP, Perlu tindakan mendahului grafik. Saran memakai seluruh lebar isi dengan tiga kartu awal, diikuti dua peringkat yang berdampingan hanya ketika ruang cukup, lalu Peluang selebar isi. Peringkat dan Peluang menampilkan lima baris awal. Tombol buka/ringkas menyediakan sisanya tanpa mengubah urutan dari server.

Filter toko/produk disisipkan melalui snippet kontrol di dalam kartu Penjualan harian. Komponen grafik mempertahankan instansinya ketika filter berubah, sehingga fokus pada pilihan tetap terjaga dan data dimuat ulang melalui efek yang sudah ada. Tampilan tabel tetap tersedia sebagai alternatif grafik.

Komponen daftar dan token desain yang sudah tersedia dipakai kembali. Tema terang/gelap, keadaan kosong/gagal, serta pembatasan angka dan grafik untuk OWNER tetap berlaku. Dasbor staf menggunakan satu kolom penuh tanpa menyisakan kolom grafik yang kosong.

## Konsekuensi

Tidak ada perubahan kontrak API, perhitungan keuangan, atau akses data. Perubahan filter dan buka/ringkas diuji di browser; tampilan produksi diperiksa pada HP, tablet, dan desktop dalam kedua tema.
