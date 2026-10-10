# Tema terang, gelap, dan biru

## Konteks

Pada 9 Oktober 2026 tim meminta redesain seluruh halaman dengan warna utama biru biasa serta tema terang dan gelap. Rencana antarmuka telah direview sebelum implementasi. Desain lama memakai satu tema gelap dengan cyan, sehingga tidak memenuhi keputusan baru. Pengguna menggunakan HP di toko dan ruang produksi dengan kondisi pencahayaan yang berbeda.

## Keputusan

Gunakan primary `#2563EB`, permukaan terang putih/abu, dan permukaan gelap navy netral. Pisahkan token warna tombol utama dari token tautan agar teks memenuhi kontras pada kedua tema. Pertahankan Inter, JetBrains Mono, seluruh aturan bisnis, hak akses, dan jalur catat luring.

Preferensi tampilan adalah Sistem, Terang, atau Gelap. Simpan hanya preferensi tersebut di penyimpanan lokal perangkat. Skrip kecil dari asal sendiri menerapkannya sebelum hidrasi, memperbarui warna bilah peramban, mengikuti perubahan sistem bila pilihan Sistem aktif, dan menyinkronkan perubahan antar tab. Tidak ada pelonggaran CSP atau penyimpanan token autentikasi.

Gunakan shell publik untuk pengunjung/onboarding dan navigasi aplikasi menurut tipe organisasi bagi anggota yang telah masuk, termasuk direktori dan papan kebutuhan. Kerangka dengan sesi yang sedang dimuat mempertahankan instance isi, sehingga perubahan navigasi tidak mereset input pengguna. Navigasi bawah mengikuti spesifikasi; sidebar hanya tampil di desktop. Tema berlaku pada seluruh rute, komponen umpan balik, grafik, peta, halaman galat, ikon PWA, dan formulir. Ringkasan pencatatan ditempatkan dekat tindakan simpan, di atas navigasi dan safe area perangkat.

## Akibat

Dokumen desain serta ringkasan visual PRD dan spesifikasi diperbarui. Pustaka tambahan tidak diperlukan. Pemeriksaan mencakup preferensi tersimpan, perubahan tema sistem, lint, tipe, aksesibilitas kedua tema, tata letak responsif, alur pencatatan dan sinkronisasi, serta anggaran bundel/performa. Tema merupakan preferensi perangkat dan tidak bergantung pada hak ubah profil usaha.

Pemilih tema juga bekerja pada halaman yang menonaktifkan hidrasi, termasuk halaman luring. Skrip asal sendiri menangani perubahan kontrol secara native dan menyelaraskan semua pemilih saat halaman dimuat atau komponen dipasang. Uji luring menemukan jalur manifest relatif tidak cocok dengan pathname permintaan; service worker menormalkan jalur terhadap scope dan mengarahkan halaman yang belum tersimpan ke `/luring`. Pola jalur mengikuti [dokumentasi service worker SvelteKit](https://svelte.dev/docs/kit/service-workers).
