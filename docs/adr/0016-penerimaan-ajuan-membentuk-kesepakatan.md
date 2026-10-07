# 0016. Penerimaan Ajuan membentuk Kesepakatan lewat port

## Konteks

BR-PRP-03 meminta Kemitraan `ACTIVE` dan Kesepakatan versi 1 dibuat dalam satu transaksi saat Ajuan diterima, dan R-SEC-27 meminta peristiwa itu disegel. Ajuan, Kesepakatan, dan segel adalah modul berbeda yang tidak boleh saling mengimpor (R-API-02). Isi yang disegel harus bisa dihitung ulang saat halaman Kesepakatan dibuka (R-SEC-33). Kesepakatan tidak punya kolom nomor, sedangkan entri segel butuh nomor dokumen. FR-PRP-03 meminta alasan saat menolak, tetapi model data belum punya kolomnya. Pemberitahuan `PROPOSAL_ACCEPTED` dan `PROPOSAL_DECLINED` ditujukan ke OWNER penulis revisi terakhir.

## Keputusan

1. Layanan Ajuan mendefinisikan port `AgreementFounder` dan `Announcer`. Akar komposisi menyambungkan `AgreementFounder` ke `agreement/service.Founder`, yang membuat baris Kemitraan, Kesepakatan versi 1 beserta produknya, lalu menyegelnya lewat port penyegel miliknya. Semua berjalan di transaksi yang dibuka layanan Ajuan.
2. Isi Kesepakatan yang disegel disusun di `agreement/domain` (`Agreement.Content`), satu-satunya tempat yang juga dipakai saat memeriksa segel. Produk diurutkan menurut ID supaya urutan simpan tidak mengubah hash. Status produk ditulis sebagai teks `ACTIVE` atau `STOPPED` karena bentuk kanonik hanya memuat string dan bilangan bulat.
3. Nomor Kesepakatan untuk segel dan tampilan adalah nomor Ajuan sumbernya ditambah versi, misalnya `AJ-202610-0001/V1`.
4. Ajuan dikunci dengan `FOR UPDATE` sebelum diputuskan, dan keputusan disimpan dengan `UPDATE` bersyarat status `OPEN` dan nomor revisi yang sama. Keputusan kedua yang datang bersamaan mendapat `PROPOSAL_ALREADY_DECIDED`.
5. Kolom `decision_reason` ditambahkan ke `proposals` untuk alasan penolakan.
6. `Announcer` di akar komposisi mencari anggota aktif organisasi tujuan lalu memanggil `Notifier`, dengan penyaring OWNER saja. Penarikan Ajuan tidak memberi tahu siapa pun karena tidak ada di katalog peristiwa.

## Akibat

- Penolakan oleh basis data atas kemitraan aktif ganda (BR-PRT-01) muncul sebagai `ALREADY_PARTNERS` walau pemeriksaan saat klaim sudah lolos.
- Modul `partnership` nanti cukup membaca tabel `partnerships`. Pembuatannya tetap satu jalur lewat penerimaan Ajuan.
