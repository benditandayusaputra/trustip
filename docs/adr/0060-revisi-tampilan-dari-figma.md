# 0060: Revisi tampilan dari desain Figma

Tanggal: 27 Oktober 2026

## Konteks

Setelah revisi ADR 0059, tim mengekspor desain Figma untuk bagian yang belum punya desain: Beranda, Masuk (frame LOG), Demo, Laporan Setoran, Pengaturan, Profil, dan Pemberitahuan. Pengaturan lama mencampur profil usaha, publikasi, pembayaran, dan tampilan dalam satu halaman panjang.

## Keputusan

1. Profil usaha pindah ke halaman baru `/app/profil` bersama kartu akun. Isian usaha terkunci sampai tombol Edit profil ditekan. Menu header Akun berganti nama menjadi Profil dan Kelola profil menaut ke `/app/profil`.
2. Pengaturan hanya memuat Tampilan (sakelar terang/gelap), Tampil di Cari Mitra, Info pembayaran yang terkunci sampai Edit, dan Keluar dari semua perangkat.
3. Profil dan Pemberitahuan menjadi halaman fokus tanpa sidebar, dengan tombol kembali di header.
4. Pemberitahuan mendapat jumlah belum dibaca, pilihan Semua dan Belum dibaca, serta tombol Tandai dibaca per kartu memakai API yang sudah ada.
5. Laporan Setoran memakai satu panel saringan untuk ringkasan dan daftar, tabel Ringkasan per Mitra, dan pencarian mitra atau nomor Setoran di perangkat.
6. Masuk, Daftar, dan Demo memakai panel biru penuh setengah layar dengan tombol kembali dan merek; pilihan tema dan footer dihapus dari kerangka akun. Tombol akun demo berbunyi Coba sebagai.
7. Pilihan tema di header berlabel Mode, judul header berwarna teks biasa setinggi 64 px, dan latar halaman terang menjadi `#F0F9FF`.
8. Beranda mengikuti Figma: diagram Satu alur, dua mitra, panel bergradasi lembut, judul tanpa label kecil kecuali Begini hasilnya, dan paragraf rata kiri.

## Yang tidak diikuti

- Foto profil, nomor HP akun, dan ubah kata sandi di desain Profil belum dibuat karena API akun belum menyediakannya. Ketiganya menyusul di ADR 0061.
- Logo di desain Figma berbeda antarbagian (kubus dan tumpukan), jadi logo yang disetujui 10 Oktober 2026 tetap dipakai.
- Huruf tebal di isi pemberitahuan tidak dibuat karena teks berasal dari server sebagai teks biasa.

## Akibat

API dan aturan bisnis tidak berubah. Uji e2e yang membuka profil toko atau mengubah nama usaha sekarang melalui `/app/profil`.
