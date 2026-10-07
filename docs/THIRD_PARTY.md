# Pustaka dan Layanan Pihak Ketiga

Setiap pustaka atau layanan pihak ketiga dicatat di sini dalam commit yang sama dengan pemakaian pertamanya (`03-ATURAN-REKAYASA.md` R-CC-52). Versi mengikuti berkas kunci (`api/go.sum`, `app/package-lock.json`).

## API (Go)

| Pustaka | Versi | Lisensi | Ukuran | Alasan dipakai |
|---|---|---|---|---|
| `github.com/gofiber/fiber/v3` | 3.5.0 | MIT | Biner server 8,3 MB setelah `-s -w` | Kerangka HTTP yang cepat dengan uji tanpa jaringan lewat `app.Test` |
