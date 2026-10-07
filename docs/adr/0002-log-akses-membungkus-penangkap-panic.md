# 0002. Log akses membungkus penangkap panic

## Konteks

Tabel rantai perantara di `03-ATURAN-REKAYASA.md` bagian 4.4 menempatkan penangkap `panic` di urutan 2 dan log akses di urutan 3. Dengan urutan itu, `panic` di penangan melompati log akses, sehingga permintaan yang paling perlu ditelusuri justru tidak punya baris log akses.

Pustaka standar Go dan aturan `forbidigo` melarang `panic` di luar `cmd`, jadi log akses tidak bisa menangkap lalu melempar ulang `panic`.

## Keputusan

Urutan di server menjadi: nomor permintaan, log akses, penangkap `panic`, lalu sisanya sesuai tabel 4.4.

Log akses memanggil penangan galat untuk galat yang naik dari dalam rantai, sehingga status yang dicatat selalu status yang benar-benar dikirim.

## Akibat

- Setiap permintaan, termasuk yang berakhir dengan `panic`, menghasilkan tepat satu baris log akses dengan status 500.
- Galat tak terduga tetap dicatat sekali di aliran `app` oleh penangan galat (R-CC-49).
- Galat dari perantara nomor permintaan, yang berada di luar log akses, tetap ditangani penangan galat bawaan Fiber yang sama.
