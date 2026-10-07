# 0005. Service worker memanggil jaringan langsung

## Konteks

R-CC-50 menetapkan `fetch` hanya dipanggil di `app/src/lib/api/client.ts`, dan ESLint menegakkannya dengan `no-restricted-globals`. Service worker berjalan di konteks terpisah tanpa akses ke `apiClient`, dan tugasnya justru menjadi perantara jaringan untuk cangkang dan aset statis (FR-PWA-02).

## Keputusan

Larangan `fetch` tidak berlaku untuk `src/service-worker/**` dan `src/lib/server/**`. Pengecualian ditulis di satu tempat, `app/eslint.config.js`.

Service worker hanya meneruskan permintaan GET satu asal di luar `/api/`, melayani berkas cangkang dari simpanan, dan mengganti halaman navigasi yang gagal dengan `/luring`. Ia tidak pernah menjawab permintaan tulis dan tidak menyimpan jawaban API (BR-PWA-01, R-OFF-13).

## Akibat

- Kode aplikasi di peramban tetap hanya memanggil jaringan lewat `apiClient`.
- Simpanan baca untuk data API (fitur 8.1) ditambahkan di service worker dengan aturan yang sama, tanpa menyentuh permintaan tulis.
