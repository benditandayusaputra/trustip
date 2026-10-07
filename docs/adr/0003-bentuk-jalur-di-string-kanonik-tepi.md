# 0003. Bentuk jalur di string kanonik tanda tangan tepi

## Konteks

String kanonik di `03-ATURAN-REKAYASA.md` bagian 7.6 menulis baris ketiga sebagai `{JALUR}?{KUERI}`. Aturan itu tidak menyebut bentuknya saat permintaan tidak punya kueri, padahal proxy di app dan verifikator di API harus menghasilkan string yang sama persis sampai karakter terakhir.

## Keputusan

1. Baris ketiga adalah URI permintaan apa adanya seperti yang diteruskan proxy: jalur, lalu `?` dan kueri mentah hanya bila kueri tidak kosong. Permintaan tanpa kueri tidak diberi `?` di ujung.
2. Kueri tidak diurutkan dan tidak dinormalkan. Proxy meneruskan URI yang sama dengan yang ditandatanganinya.
3. Contoh resmi ada di `api/testdata/vectors/edge_signature.json`. Uji Go dan uji TypeScript sama-sama membaca berkas itu (R-REPO-02, R-TEST-04).

## Akibat

- Proxy tidak boleh mengubah URI setelah menandatangani, termasuk menambah atau membuang `?` kosong.
- Perubahan bentuk string kanonik wajib menaikkan versi baris pertama dari `v1`, supaya proxy dan API yang berbeda versi gagal secara jelas.
