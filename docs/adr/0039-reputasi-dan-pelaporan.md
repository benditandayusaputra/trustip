# 0039. Reputasi dan pelaporan

## Konteks

BR-TRS-01 dan BR-TRS-02 meminta ketepatan bayar toko dari Setoran yang jatuh tempo dalam 180 hari, dihitung ulang tugas harian dan disimpan sebagai ringkasan. FR-TRS-04 meminta tombol Laporkan dengan alasan, dan BR-TRS-05 meminta laporan tersimpan untuk ditindak tim langsung di basis data sebelum halaman admin ada. BR-AUTHZ-06 melarang organisasi `SUSPENDED` mengirim Ajuan. Spesifikasi tidak menetapkan batas hari "sudah jatuh tempo", tanggal pembayaran yang dipakai, daftar alasan laporan, perlakuan laporan ganda, dan titik penegakan penangguhan.

## Keputusan

1. Modul `trust` memegang ringkasan reputasi dan laporan. Tugas `reputation-summary` berjalan setiap hari pukul 00.15 Asia/Jakarta, setelah penerbitan Setoran berkala, dan menulis seluruh organisasi ke `org_stats` dalam satu pernyataan.
2. Setoran dihitung "sudah jatuh tempo" bila tanggal jatuh temponya sebelum hari ini dan paling lama 180 hari lalu. Setoran tepat waktu bila punya pembayaran `CONFIRMED` dengan tanggal bayar tidak melewati jatuh tempo. Persentase disimpan dalam basis poin dan dibulatkan setengah ke atas.
3. Kartu Cari Mitra menampilkan jumlah mitra aktif dan tanggal bergabung. Tanda "Baru" tetap mengikuti aturan direktori (ADR 0032).
4. Alasan laporan adalah `MISLEADING`, `PROHIBITED_GOODS`, `OFFENSIVE`, `SPAM`, dan `OTHER`. `OTHER` wajib disertai catatan. Laporan ditujukan ke organisasi, produk, atau Kebutuhan milik organisasi lain; sasaran yang tidak ada atau milik sendiri dibalas 404.
5. Satu pengguna punya paling banyak satu laporan terbuka per sasaran. Laporan ulang memperbarui alasan dan catatannya. Rute dibatasi 5 laporan per jam per pengguna (tabel 7.9).
6. Penangguhan ditegakkan di kueri kuota Ajuan harian: kuota hanya bertambah untuk organisasi `ACTIVE`, sehingga Ajuan langsung dan undangan dari organisasi `SUSPENDED` dibalas `ORGANIZATION_SUSPENDED`. Kewajiban yang berjalan tidak tersentuh. Cari Mitra sudah hanya menampilkan organisasi `ACTIVE`.

## Akibat

- Panduan menindak laporan dan menangguhkan organisasi ada di `docs/MODERASI.md`.
- Halaman admin moderasi tetap P1 (FR-TRS-05).
