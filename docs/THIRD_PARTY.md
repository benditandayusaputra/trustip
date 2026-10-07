# Pustaka dan Layanan Pihak Ketiga

Setiap pustaka atau layanan pihak ketiga dicatat di sini dalam commit yang sama dengan pemakaian pertamanya (`03-ATURAN-REKAYASA.md` R-CC-52). Versi mengikuti berkas kunci (`api/go.sum`, `app/package-lock.json`).

## API (Go)

| Pustaka | Versi | Lisensi | Ukuran | Alasan dipakai |
|---|---|---|---|---|
| `github.com/gofiber/fiber/v3` | 3.5.0 | MIT | Biner server 8,3 MB setelah `-s -w` | Kerangka HTTP yang cepat dengan uji tanpa jaringan lewat `app.Test` |

## Alat pengembangan

Dipasang `make setup` ke `.bin/` dengan versi terkunci di `Makefile`. Tidak ikut ke biner atau bundel.

| Alat | Versi | Lisensi | Alasan dipakai |
|---|---|---|---|
| `golangci-lint` | 2.14.0 | GPL-3.0 | Menjalankan semua pemeriksa Go di tabel 3.13 dan format `gofumpt` |
| `gitleaks` | 8.30.1 | MIT | Pemindai rahasia di kait pra-commit dan di CI |
| `lefthook` | 2.1.17 | MIT | Pengelola kait git |
| `govulncheck` | 1.8.0 | BSD-3-Clause | Pemindai kerentanan dependensi Go |
| `deadcode` | 0.51.0 | BSD-3-Clause | Pencari kode mati |

## App (SvelteKit)

Ukuran untuk pustaka yang ikut ke peramban diukur dari bundel terkompresi. Alat pengembangan tidak ikut ke bundel.

| Pustaka | Versi | Lisensi | Ukuran | Alasan dipakai |
|---|---|---|---|---|
| `svelte` | 5.57 | MIT | Runtime kecil, terkompilasi per komponen | Kerangka antarmuka dengan rune untuk keadaan |
| `@sveltejs/kit` | 3.0 | MIT | Ikut bundel per rute | Rute, render server, proxy, service worker |
| `@sveltejs/adapter-vercel` | 7.0 | MIT | Alat bangun | Deploy app ke Vercel |
| `vite` | 8.3 | MIT | Alat bangun | Pembangun dan server pengembangan |
| `tailwindcss`, `@tailwindcss/vite` | 4.3 | MIT | Hanya kelas yang dipakai | Gaya dari token desain |
| `typescript`, `svelte-check` | 6.0, 4.6 | Apache-2.0, MIT | Alat pengembangan | Pemeriksaan tipe ketat |
| `eslint`, `typescript-eslint`, `eslint-plugin-svelte`, `svelte-eslint-parser`, `@eslint/js`, `eslint-config-prettier`, `globals` | 10.12, 8.60, 3.23, 1.8, 10.0, 10.1, 17.6 | MIT | Alat pengembangan | Lint mutu, keamanan, dan batas lapisan |
| `prettier`, `prettier-plugin-svelte`, `prettier-plugin-tailwindcss` | 3.9, 4.1, 0.8 | MIT | Alat pengembangan | Format otomatis |
| `vitest`, `@vitest/browser-playwright` | 4.1 | MIT | Alat pengembangan | Uji unit dan komponen |
| `@playwright/test`, `playwright` | 1.63 | Apache-2.0 | Alat pengembangan | Uji ujung ke ujung di HP dan desktop |
| `knip` | 6.40 | ISC | Alat pengembangan | Pencari ekspor dan dependensi tak terpakai |
| `@types/node` | 24 | MIT | Alat pengembangan | Tipe Node untuk berkas konfigurasi |
