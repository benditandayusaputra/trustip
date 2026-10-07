# 0010. Tabel izin peran di `kernel/policy`

## Konteks

Aturan 7.8 meminta tabel izin di spesifikasi bagian 1 ditulis sekali sebagai data di `service/policy`, dan R-SEC-36 meminta setiap kasus pakai memanggil `policy.Require(actor, action)` di baris pertamanya. Tabel 4.2 hanya mengizinkan `service` sebuah modul mengimpor `domain` modulnya sendiri dan `kernel`. Tidak ada satu paket `service` bersama yang bisa diimpor semua modul, sehingga satu tabel di `modules/<x>/service/policy` tidak terjangkau dari modul lain.

Tabel spesifikasi juga memuat dua baris P1 (kelola anggota, umumkan kiriman) dan satu sel "Bila diizinkan OWNER" untuk STAFF toko yang mencatat pembayaran, yang baru bisa diberikan setelah P1.2.

Modul `auth` (daftar, masuk, segarkan, keluar, baca akun) berjalan sebelum pengguna punya organisasi, sehingga tidak ada peran yang bisa diperiksa.

## Keputusan

1. Tabel izin dan `Require` tinggal di `api/internal/kernel/policy`. Semua lapisan boleh mengimpor `kernel`, jadi tabel tetap ditulis sekali.
2. Pelaku (`kernel/actor`) membawa keanggotaan: organisasi, tipe organisasi, dan peran. Nilainya berasal dari klaim `org`, `org_type`, dan `role` di token akses yang diisi saat masuk dan saat sesi disegarkan (R-SEC-16, BR-AUTHZ-05).
3. Tabel memuat semua baris spesifikasi, termasuk baris P1, supaya uji `TestTable_RSEC37_*` bisa membandingkannya baris demi baris dengan membaca tabel di `docs/02-SPESIFIKASI-FITUR.md`.
4. Sel "Bila diizinkan OWNER" disimpan sebagai `DelegatedByOwner`. `Require` memperlakukannya sebagai ditolak sampai P1.2 menambah izin dari OWNER.
5. Pengguna tanpa organisasi dan aksi yang tidak ada di tabel dibalas `FORBIDDEN`.
6. Uji `TestArchitecture_RSEC36_EveryUseCaseRequiresAPolicyFirst` memeriksa setiap metode `Execute` di lapisan `service`. Modul `auth` dikecualikan karena kasus pakainya membentuk sesi itu sendiri.

## Akibat

- Perubahan tabel di spesifikasi langsung membuat uji gagal sampai kode disamakan.
- Perubahan keanggotaan baru terlihat di token setelah sesi disegarkan, paling lama 15 menit. Perubahan peran dan penonaktifan anggota (P1) wajib mencabut sesi yang terdampak (R-SEC-20).
