# 0008. Pemeriksa asal dipasang global sebelum autentikasi

## Konteks

Tabel 4.4 menempatkan autentikasi di urutan 9 dan pemeriksa asal di urutan 10, keduanya dipasang per kelompok rute (R-API-07). R-SEC-21 mewajibkan pemeriksaan kepala `Origin` dan `X-Trustip-Client: web` untuk **setiap** permintaan selain GET, termasuk rute publik seperti daftar, masuk, segarkan, dan keluar yang tidak punya perantara autentikasi.

Di Fiber, perantara yang dipasang pada kelompok berawalan sama berlaku untuk semua rute di bawah awalan itu, termasuk rute publik. Memasang pemeriksa asal per rute membuka peluang rute tulis baru lupa dipasangi.

## Keputusan

1. `RequireSameOrigin` dipasang sekali di rantai global, tepat setelah batas waktu (urutan 8), sehingga berjalan untuk semua rute sebelum autentikasi per rute.
2. GET dan HEAD dilewati. Metode lain wajib membawa `Origin` yang sama persis dengan `PUBLIC_APP_ORIGIN` dan `X-Trustip-Client: web`, atau dibalas 403 `CROSS_SITE_REJECTED` dan dicatat sebagai peristiwa keamanan.
3. Permintaan tulis yang membawa isi wajib berjenis `application/json` atau `multipart/form-data` (R-SEC-22).
4. Autentikasi tetap dipasang per rute. Uji `TestRoutes_RAPI07_EveryRouteIsPublicOrAuthenticated` memanggil setiap rute di luar daftar publik tanpa cookie dan gagal bila jawabannya bukan 401.

## Akibat

- Permintaan lintas situs ditolak sebelum token diperiksa, jadi tidak ada kerja autentikasi untuk permintaan yang pasti ditolak.
- Rute tulis baru otomatis terlindungi tanpa langkah tambahan.
- Proxy app mengulang pemeriksaan yang sama terhadap asalnya sendiri, dan SvelteKit menolak kiriman formulir lintas situs lebih dulu, sehingga ada tiga lapis bersama `SameSite=Lax`.
