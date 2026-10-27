# 0061: Ubah akun sendiri dari halaman Profil

Tanggal: 27 Oktober 2026

## Konteks

Desain Profil dari Figma (ADR 0060) memuat foto profil, nomor HP, dan tombol Ubah sandi. API akun hanya mengembalikan nama dan email, sehingga ketiganya belum bisa dibuat. Tim meminta backend-nya dibuat sekaligus.

## Keputusan

1. `GET /api/v1/me` mengembalikan `whatsapp` dan `avatar` (`ImageRef` atau `null`). Login dan daftar tetap memakai skema `Account` yang lama.
2. `PATCH /api/v1/me` mengubah nama, nomor WhatsApp, dan foto. Aturan nama dan nomor sama dengan pendaftaran. Foto harus berkas bertujuan `AVATAR` yang diunggah pengguna itu sendiri di organisasinya; selain itu ditolak `avatar_file_id` `UNKNOWN_CHOICE`. `null` menghapus foto.
3. Tujuan unggah baru `AVATAR`: privat, boleh untuk pemilik dan staf di kedua tipe usaha. Aksi izin baru `EDIT_OWN_ACCOUNT` diberikan ke semua peran.
4. `POST /api/v1/me/password` wajib kata sandi saat ini. Kata sandi baru memakai kebijakan yang sama dengan pendaftaran dan harus berbeda dari yang lama. Setelah berhasil, semua keluarga sesi lain dicabut dan token aksesnya diblokir; perangkat yang dipakai tetap masuk. Batas laju 5 per menit dan 20 per jam per pengguna.
5. Kedua perubahan dicatat di log audit (`ACCOUNT_UPDATED`, `PASSWORD_CHANGED`).
6. Migrasi 00027 menambah `users.avatar_file_id`, memperluas `files_purpose_known`, dan memperbarui `reset_demo_data()` supaya mengosongkan foto pengguna demo sebelum menghapus berkasnya.

## Akibat

Rute baru masuk kontrak `openapi.yaml` dan tipe app dibangkitkan ulang. Migrasi hanya menambah, sehingga API versi sebelumnya tetap berjalan bila perlu mundur.
