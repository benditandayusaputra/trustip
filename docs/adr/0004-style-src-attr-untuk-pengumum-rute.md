# 0004. `style-src-attr` untuk pengumum rute kerangka

## Konteks

R-SEC-67 menetapkan `style-src 'self'` dan R-SEC-68 melarang `'unsafe-inline'` di `script-src`. Uji ujung ke ujung menemukan satu pelanggaran CSP di setiap halaman: SvelteKit memasang elemen `#svelte-announcer` untuk mengumumkan pergantian halaman ke pembaca layar, dan elemen itu memakai atribut `style` sebaris untuk menyembunyikannya secara visual (`node_modules/@sveltejs/kit/src/runtime/components/root.svelte`). Atribut itu tidak bisa diberi hash.

Mematikan pengumum rute akan merusak aksesibilitas (NFR-11), dan mengganti komponen akar kerangka berarti memelihara salinan kode pihak ketiga.

## Keputusan

Sesuai R-SEC-68, yang dibuka hanya `style-src-attr 'unsafe-inline'`. `script-src` tetap `'self'` ditambah hash yang dihitung kerangka, dan `style-src` untuk elemen `<style>` serta berkas gaya tetap `'self'`.

Komponen milik sendiri tetap tidak boleh memakai atribut `style`. Semua gaya lewat kelas token (R-APP-08), dan pemeriksa token menolak nilai mentah.

## Akibat

- Atribut `style` sebaris dari kode mana pun diizinkan peramban. Risiko yang tersisa adalah penyuntikan CSS lewat atribut, bukan eksekusi skrip.
- Uji `tests/security-headers.test.ts` memastikan `script-src` tidak memuat `unsafe` dan tidak ada pelanggaran CSP di konsol.
- Bila versi SvelteKit berikutnya memindahkan gaya pengumum ke kelas, kelonggaran ini dicabut.
