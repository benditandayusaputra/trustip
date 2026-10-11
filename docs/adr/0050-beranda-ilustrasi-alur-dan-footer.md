# ADR 0050: beranda dengan ilustrasi, alur, dan footer

Tanggal: 11 Oktober 2026. Status: diterima.

## Konteks

Pengguna meminta beranda yang langsung menjelaskan kegunaan Trustip, ilustrasi, alur penggunaan yang jelas, dan footer yang lebih lengkap. Penjelasan harus mengikuti perilaku aplikasi dan kedua tema.

## Keputusan

Layar pertama menjelaskan platform titip jual untuk UMKM dan menampilkan ilustrasi penitip bersama toko. Ilustrasi vektor akun dijadikan komponen bersama; ukuran beranda lebih besar, sedangkan halaman akun tetap ringkas pada HP.

Kebutuhan penitip dan toko ditampilkan terpisah. Alur dijelaskan dalam lima langkah bernomor: mencari mitra, menyepakati harga, mencatat titipan, menghitung sisa, dan mengonfirmasi setoran. Contoh hitungan ditempatkan setelah penjelasan alur, dengan label contoh dan keterangan bahwa pembayaran dilakukan langsung antar mitra.

Footer publik memuat identitas aplikasi serta kelompok tautan direktori, cara kerja, demo, pendaftaran, privasi, dan ketentuan. Tidak ada tautan kosong, kontak rekaan, atau klaim bisnis yang belum terbukti. Footer dipakai pada kerangka publik untuk beranda, profil, kebijakan, dan halaman luring. Direktori dengan hasil yang dimuat setelah halaman terbuka mempertahankan footer ringkas agar perluasan footer tidak menambah pergeseran tata letak selama pemuatan.

## Konsekuensi

Tidak ada perubahan API, dependensi, aturan dagang, atau autentikasi. Komponen memakai katalog teks dan token desain. Uji tautan, navigasi keyboard, tema, ukuran layar, aksesibilitas, dan anggaran halaman publik tetap diwajibkan.
