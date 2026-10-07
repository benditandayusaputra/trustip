# 0006. Hasil kunci uji Turnstile tidak diperiksa host dan aksinya

## Konteks

R-SEC-10 mewajibkan server memeriksa tiga hal pada jawaban verifikasi Turnstile: berhasil, nama host cocok, dan nama aksi cocok dengan rute. R-SEC-12 mewajibkan uji ujung ke ujung memakai kunci uji resmi Turnstile.

Jawaban layanan verifikasi untuk kunci rahasia uji resmi (`1x0000000000000000000000000000000AA`) selalu berbentuk:

```json
{"success": true, "hostname": "example.com", "error-codes": [], "metadata": {"result_with_testing_key": true}}
```

Nama host selalu `example.com` dan aksi tidak ada. Bila ketiga pemeriksaan dijalankan apa adanya, uji ujung ke ujung tidak pernah bisa lulus dengan kunci uji resmi.

## Keputusan

1. Bila jawaban memuat `metadata.result_with_testing_key: true`, server tetap mewajibkan `success: true` tetapi tidak membandingkan nama host dan aksi.
2. Tidak ada sakelar konfigurasi untuk ini. Penanda itu hanya diberikan layanan Turnstile untuk kunci rahasia uji, dan isinya tidak bisa diatur oleh peramban.
3. Pemuat konfigurasi API sudah menolak mulai bila `APP_ENV=production` memakai kunci uji Turnstile, sehingga cabang ini tidak pernah terjadi di produksi.

## Akibat

- Uji ujung ke ujung berjalan dengan widget dan layanan verifikasi Turnstile yang asli, memakai kunci uji resmi.
- Di produksi ketiga pemeriksaan R-SEC-10 tetap berlaku penuh.
- Uji unit `platform/captcha` mengunci kedua perilaku: hasil kunci uji yang berhasil diterima, dan hasil kunci uji yang gagal tetap ditolak.
