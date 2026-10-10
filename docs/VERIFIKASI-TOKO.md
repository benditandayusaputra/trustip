# Verifikasi dasbor toko dan Cari Toko

Tanggal pemeriksaan: 10 Oktober 2026. Lingkup: tata letak `/app/toko`, saringan `/cari/toko`, dan FR-DIR-13 pada direktori Toko. Keputusan tercatat dalam ADR 0047.

## Perilaku yang diperiksa

- Tujuh saran tetap menampilkan tiga kartu awal; tombol membuka semua saran dan mengembalikannya ke keadaan ringkas.
- Hitung cepat, Produk terlaris, dan Peluang membatasi tampilan awal tanpa menghilangkan akses ke data lainnya.
- Pencarian dan urutan terlihat sejak awal; saringan tambahan dibuka sesuai kebutuhan, dengan label, penjelasan, keadaan tombol, dan indikator jumlah pilihan aktif.
- Perubahan saringan tetap memperbarui alamat dan hasil. Kesalahan jaringan menyediakan Coba lagi dan mempertahankan pilihan.
- Daftar dan Peta memakai hasil yang sama. Klik marker, Enter, atau Space membuka detail. Fokus tetap pada marker setelah penggunaan keyboard.
- Toko pada koordinat identik tidak saling menutupi: satu marker menyatakan jumlah toko, dan daftar nama menyediakan pilihan setiap toko.
- Toko tanpa koordinat tetap tersedia dalam Daftar. API hanya mengembalikan lokasi publik; uji memakai lokasi publik yang berbeda dari lokasi privat serta lokasi yang kosong.
- Nama toko berisi karakter HTML tetap menjadi teks. Tidak ada HTML pengguna yang dimasukkan ke marker.

## Hasil

| Pemeriksaan | Hasil |
|---|---|
| `make check` | Lulus; format, lint, tipe, arsitektur, komentar, dan pemeriksaan rahasia |
| Audit dependensi frontend | Tidak ada kerentanan |
| Uji frontend | 83 berkas, 500 uji lulus |
| `make e2e` | 67 uji lulus dalam seluruh kelompok wajib |
| Peta dan saringan | Uji HP dan desktop lulus; audit semua aturan axe tanpa pelanggaran pada kedua tema |
| Uji integrasi lokasi publik dan lokasi kosong | Lulus dengan race detector pada basis data uji Neon |
| Pemeriksaan visual | 40 tangkapan layar; 320, 390, 768, 1024, dan 1440 px; tema terang dan gelap |
| Overflow horizontal dan galat JavaScript | Tidak ditemukan pada 40 tampilan |
| Uji API penuh | Seluruh paket lulus dengan race detector dan basis data uji Neon |

Artefak lokal berada di `../output/store-layout/index.html` dan `../output/store-layout/verification.json` dari akar repositori. Pengujian memakai build produksi frontend, API lokal, dan Neon live sesuai aturan proyek.

## Batas bukti

Audit otomatis dan pengujian keyboard mencakup halaman serta keadaan yang disebut di atas. Hasil ini tidak menggantikan pengujian pengguna dengan pembaca layar dan tidak membuktikan performa deployment. Peta menampilkan toko dari hasil yang sudah dimuat; Muat lagi mengikuti pagination yang ada. Ubin OpenStreetMap membutuhkan koneksi, sedangkan daftar nama dan kartu detail menyediakan alternatif ketika peta gagal dimuat.
