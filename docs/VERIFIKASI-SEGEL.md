# Verifikasi Segel

Setiap peristiwa dagang di Trustip (Kesepakatan disetujui, Titipan Masuk, Hitung Sisa, Setoran, pembayaran, dan kemitraan berakhir) disegel. Segel membuktikan dua hal: isi dokumen tidak berubah sejak dicatat, dan akun mana yang menyetujuinya. Segel bukan tanda tangan elektronik tersertifikasi dari penyelenggara sertifikasi elektronik.

## Segel di belakang layar

Sejak ADR 0058, segel tidak tampil di antarmuka. Detail dokumen tidak lagi memuat tanda "Segel sah", jejak segel, atau kode verifikasi, dan halaman `/verifikasi/<kode>` sudah dihapus. Penyegelan dan pemeriksaannya tetap berjalan di server.

## Memeriksa lewat rute publik API

Tim pengelola mengambil kode verifikasi dari tabel `seals`, lalu memanggil:

```
GET /api/v1/public/seals/<kode>
```

Contoh kode: `V7KA-DYAY-2Y4Z-3FGX`. Huruf kecil dan tanpa tanda hubung juga diterima.

Jawabannya memuat jenis dan nomor dokumen, nama kedua pihak, jenis peristiwa, waktu, dan status segel. Nilai uang, jumlah barang, dan syarat tidak pernah ikut. Kode yang tidak dikenal dijawab 404.

Status sah berarti tiga pemeriksaan lulus:

1. Hash entri yang dihitung ulang dari isi entri sama dengan hash yang tersimpan.
2. Tanda tangan Ed25519 atas hash entri cocok dengan kunci publik platform.
3. `prev_hash` entri sama dengan hash entri sebelumnya pada kemitraan yang sama, sehingga rantai tidak putus.

Pemeriksaan isi dokumen terhadap `content_hash` dilakukan server setiap kali salah satu pihak kemitraan membuka dokumen. Hasilnya tidak tampil di antarmuka. Bila tidak cocok, server mencatat peristiwa keamanan `seal_invalid`.

## Kunci publik

Kunci publik platform diterbitkan di:

```
GET /api/v1/public/seal-keys
```

Jawabannya daftar kunci dengan `key_id`, `algorithm` (`Ed25519`), dan `public_key` (32 byte dalam base64). Setiap entri segel mencatat `key_id` yang dipakai saat menandatangani, sehingga kunci lama tetap bisa dipakai memeriksa segel lama setelah kunci diganti.

## Format kanonik

Isi dokumen dan entri segel diubah ke bentuk kanonik sebelum di-hash:

- JSON tanpa spasi.
- Kunci objek diurutkan menurut urutan unit kode UTF-16 (RFC 8785).
- Nilai hanya objek, larik, string, dan bilangan bulat. Uang selalu bilangan bulat rupiah.
- String di-escape seperti JSON baku.

Entri segel berisi kunci berikut, dengan `v` bernilai `1`:

| Kunci | Isi |
|---|---|
| `v` | Versi format |
| `partnership_id` | ID kemitraan |
| `seq` | Nomor urut entri pada kemitraan, mulai dari 1 |
| `event_type` | Jenis peristiwa, misalnya `DROP_RECORDED` |
| `entity_type`, `entity_id`, `entity_number` | Jenis, ID, dan nomor dokumen |
| `content_hash` | SHA-256 isi dokumen kanonik, heksadesimal huruf kecil |
| `actor_user_id`, `actor_org_id` | Pengguna dan organisasi pelaku, kosong untuk peristiwa sistem |
| `occurred_at` | Waktu UTC dengan enam digit mikrodetik, misalnya `2026-10-19T08:00:00.123456Z` |
| `prev_hash` | Hash entri sebelumnya dalam heksadesimal, kosong untuk entri pertama |

`entry_hash` adalah SHA-256 dari entri kanonik. Tanda tangan adalah Ed25519 atas `entry_hash`. Kode verifikasi adalah 16 karakter pertama base32 (RFC 4648, tanpa padding) dari `entry_hash`, dibagi empat kelompok dengan tanda hubung.

## Memeriksa secara mandiri

Contoh lengkap ada di `api/testdata/vectors/seal.json`, yang juga dipakai uji Go dan TypeScript. Langkahnya:

1. Susun isi dokumen dalam bentuk kanonik, lalu hitung SHA-256-nya. Hasilnya harus sama dengan `content_hash`.
2. Susun entri dalam bentuk kanonik dengan `content_hash` tadi, lalu hitung SHA-256-nya. Hasilnya harus sama dengan `entry_hash`.
3. Ambil kunci publik dengan `key_id` entri dari rute kunci publik, lalu verifikasi tanda tangan Ed25519 atas `entry_hash`.
4. Untuk rantai, pastikan `prev_hash` setiap entri sama dengan `entry_hash` entri sebelumnya pada kemitraan yang sama.

Contoh dengan Python dan pustaka `cryptography`:

```python
import base64, hashlib, json
from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PrivateKey

vector = json.load(open("api/testdata/vectors/seal.json"))
entry_hash = hashlib.sha256(vector["entry_canonical"].encode()).hexdigest()
assert entry_hash == vector["entry_hash"]

code = base64.b32encode(bytes.fromhex(entry_hash)).decode()[:16]
assert "-".join(code[i:i + 4] for i in range(0, 16, 4)) == vector["code"]

seed = bytes.fromhex(vector["seed_hex"])
public_key = Ed25519PrivateKey.from_private_bytes(seed).public_key()
public_key.verify(bytes.fromhex(vector["signature"]), bytes.fromhex(entry_hash))
```

Pada segel nyata, kunci publik diambil dari rute kunci publik, bukan dibuat dari benih. Benih di vektor uji hanya kunci uji dari RFC 8032.

## Pemeriksaan harian

Tugas `seal-chain-check` berjalan setiap hari pukul 00.20 Asia/Jakarta. Tugas ini memeriksa seluruh rantai segel tiap kemitraan dengan tiga pemeriksaan di atas dan mencatat peristiwa keamanan `seal_chain_broken` untuk setiap kemitraan yang rantainya putus.
