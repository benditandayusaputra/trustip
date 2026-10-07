# 0011. Daftar kabupaten dan kota resmi Kemendagri

## Konteks

Tabel `organizations` di spesifikasi bagian 22.3 punya kolom `city_code` dan `city_name`, dan FR-DIR-04 meminta titik asal Cari Mitra bisa diganti ke kota tertentu. Spesifikasi tidak menyebut sumber daftar kota. Tim memutuskan pada 7 Oktober 2026 memakai daftar resmi Kemendagri.

## Keputusan

1. Tabel `cities` (kode, nama, provinsi, titik tengah) diisi lewat migrasi `00004_cities.sql` dari dataset `cahyadsn/wilayah` (MIT) yang mengikuti Kepmendagri Nomor 300.2.2-2430 Tahun 2025: 514 kabupaten dan kota di 38 provinsi.
2. `organizations.city_code` menjadi kunci asing ke `cities.code` dan boleh kosong sampai profil diisi. `city_name` disalin dari `cities` saat organisasi dibuat.
3. App memilih kota lewat dua pilihan bawaan: provinsi (`GET /api/v1/provinces`, 38 baris) lalu kabupaten atau kota (`GET /api/v1/cities?province_code=`, paling banyak 38 baris). Keduanya di bawah batas 50 baris per daftar (R-RES-04), jadi tidak perlu kursor atau isian pencarian.
4. Koreksi terhadap sumber dicatat di `docs/THIRD_PARTY.md`.

## Akibat

- Kode wilayah stabil dan bisa dicocokkan dengan data pemerintah.
- Pemekaran wilayah baru butuh migrasi baru yang menambah baris.
