# 0054: QA navigasi dan formulir

Tanggal: 16 Oktober 2026

## Konteks

Pengguna meminta pemeriksaan satu per satu pada seluruh halaman dan revisi langsung terhadap desain serta UX yang masih bermasalah. Pemeriksaan memakai build produksi lokal dan data demo, termasuk desktop, HP, tema terang, dan tema gelap.

## Keputusan

Prioritaskan hasil pencarian pada HP, dengan filter lanjutan yang dapat dibuka. Pertahankan pencarian sebagai kontrol yang selalu terlihat. Sediakan menu publik pada HP, penanda Lainnya aktif, nama aksesibel untuk kontrol ikon, jejak halaman yang dapat diklik, serta tautan kembali pada detail dan formulir.

Kelompokkan formulir produk menjadi tiga bagian; gunakan dua kolom untuk pasangan isian yang sesuai dan area teks untuk konten panjang. Dahulukan formulir pada halaman ubah produk dan letakkan QR dalam disclosure. Pertahankan nilai serta fokus validasi yang sudah ada.

Kelompokkan metadata dokumen ke dalam kartu, dengan label dan nilai satu kolom pada HP. Sediakan navigasi antarbagian pengaturan tanpa membongkar komponen formulir. Pisahkan pengakhiran kemitraan dalam disclosure dan pertahankan konfirmasi dua langkah.

Tetapkan satu kolom eksplisit pada grid akun dan izinkan kartu formulir menyusut agar captcha dapat memilih ukuran compact pada layar 320 px. Jangan memotong widget atau mengubah verifikasinya.

## Akibat

Sidebar tetap menggunakan primary biru. Tidak ada perubahan kontrak API, aturan bisnis, dan hak akses. QA interaksi tidak menyimpan perubahan dokumen atau profil. Halaman onboarding memakai sesi uji di browser; tinjauan antrean memakai catatan lokal NEEDS_REVIEW yang tidak dikirim ke server. Hasil cakupan dan batas pemeriksaan dijelaskan pada `docs/QA-ANTARMUKA.md`.
