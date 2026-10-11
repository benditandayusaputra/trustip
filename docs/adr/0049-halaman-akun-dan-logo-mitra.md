# ADR 0049: halaman akun dan logo mitra

Tanggal: 10 Oktober 2026. Status: diterima.

## Konteks

Pengguna meminta halaman masuk dan registrasi yang lebih rapi, memiliki ilustrasi, responsif, mengikuti tema aplikasi, dan tidak menampilkan akun contoh. Pengguna juga memberikan logo dua mitra biru dengan paket jingga untuk mengganti logo aplikasi.

## Keputusan

Kedua halaman memakai kerangka akun bersama dengan ilustrasi vektor titip jual yang mengikuti token tema. Desktop memakai dua kolom; HP menggunakan ilustrasi ringkas di atas formulir. Formulir mempertahankan validasi, captcha, keadaan luring, pesan kegagalan, persetujuan kebijakan dan tujuan lanjutan. Kontrol kata sandi menyediakan pilihan tampil/sembunyi yang dapat diakses melalui keyboard.

Akun contoh dipindahkan dari `/masuk` ke `/demo`. Ajakan demo di beranda menunjuk halaman tersebut, sehingga fungsi demo lomba tetap tersedia dan kredensial contoh tidak tampil pada halaman masuk atau registrasi. Captcha tetap wajib sebelum masuk dengan akun demo.

Logo yang diberikan pengguna menjadi sumber identitas untuk merek, favicon, ikon PWA dan ikon Apple. Ekspor mengubah ukuran untuk kebutuhan aplikasi, tanpa menggambar ulang atau mengubah warna logo. Ilustrasi memakai biru utama dan jingga sebagai aksen, bukan mengganti warna utama aplikasi.

## Konsekuensi

Tidak ada perubahan kontrak API atau autentikasi server. `/demo` perlu ikut pemeriksaan responsif, aksesibilitas dan anggaran JavaScript halaman publik. Aset logo tersedia dari server aplikasi dan tidak memerlukan layanan eksternal.
