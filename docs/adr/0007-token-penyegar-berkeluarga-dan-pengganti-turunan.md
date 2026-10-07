# 0007. Token penyegar memuat keluarga sesi dan penggantinya diturunkan

## Konteks

R-SEC-17 mewajibkan token penyegar acak yang hanya disimpan sebagai hash SHA-256 dan diputar setiap dipakai. R-SEC-18 mewajibkan server menerima token lama selama 10 detik setelah diputar dengan mengembalikan pasangan baru yang sama, supaya dua tab yang menyegarkan bersamaan tidak dianggap pencurian. Tabel 7.9 membatasi segarkan sesi 30 kali per menit per keluarga sesi.

Dua hal tidak bisa dipenuhi bila token hanya berupa 32 byte acak:

1. Server tidak menyimpan token pengganti dalam bentuk mentah, sehingga tidak bisa mengembalikan "pasangan yang sama" kepada tab kedua.
2. Pembatas laju berjalan sebelum basis data dibaca, sehingga keluarga sesi tidak diketahui saat kunci batas laju dibentuk.

## Keputusan

1. Token penyegar berbentuk `<id keluarga>.<rahasia base64url>`. ID keluarga sama dengan klaim `sid` di token akses, jadi tidak menambah informasi yang belum ada di sesi itu.
2. Token pertama memakai 32 byte acak. Token pengganti diturunkan: `rahasia = HMAC-SHA256(kunci turunan, token sebelumnya)`, dengan `kunci turunan = HMAC-SHA256(kunci JWT aktif, "trustip refresh token successor v1")`. Hanya server yang bisa menghitungnya.
3. Basis data tetap hanya menyimpan hash SHA-256 dari token utuh.
4. Saat token lama datang lagi dalam 10 detik setelah diputar, server menghitung ulang pengganti yang sama dan menandatangani token akses dengan waktu terbit sama dengan waktu putar, sehingga kedua tab menerima pasangan yang identik. Tidak ada baris baru yang ditulis.
5. Lewat dari 10 detik, token lama dianggap dipakai ulang: seluruh keluarga dicabut di basis data, ID keluarga masuk daftar cabut di memori sampai token aksesnya kedaluwarsa, dan peristiwa keamanan `refresh_token_reused` dicatat.
6. Pembatas laju segarkan memakai awalan ID keluarga dari cookie. Cookie yang bentuknya rusak memakai kunci IP.

## Akibat

- Rotasi kunci JWT mengubah kunci turunan. Pemutaran ulang yang kebetulan melintasi rotasi kunci dalam jendela 10 detik akan ditolak sebagai sesi berakhir, bukan dicabut, karena hash pengganti tidak cocok dengan baris mana pun.
- Pihak yang memalsukan awalan keluarga di cookie hanya bisa menghabiskan kuota segarkan keluarga itu bila tahu ID keluarganya, yang hanya ada di cookie HttpOnly korban.
- Uji integrasi dengan PostgreSQL nyata membuktikan dua segarkan bersamaan atas token yang sama menghasilkan satu pasangan.
