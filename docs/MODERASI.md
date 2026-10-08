# Panduan moderasi

Sebelum halaman admin tersedia (FR-TRS-05, P1), laporan dari tombol Laporkan ditindak tim langsung di basis data aplikasi (BR-TRS-05). Jalankan perintah di bawah dengan koneksi pemilik skema dari panel Neon, satu per satu, dan catat setiap tindakan di catatan tim.

## 1. Melihat laporan terbuka

```sql
SELECT r.id, r.created_at, r.target_type, r.target_id, r.reason, r.note,
       u.full_name AS pelapor, o.name AS organisasi_pelapor
FROM listing_reports AS r
JOIN users AS u ON u.id = r.reporter_user_id
JOIN organizations AS o ON o.id = r.reporter_org_id
WHERE r.status = 'OPEN'
ORDER BY r.created_at;
```

Sasaran `ORGANIZATION` mengacu ke `organizations.id`, `PRODUCT` ke `products.id`, dan `NEED` ke `needs.id`.

## 2. Menindak sasaran

| Temuan | Tindakan |
|---|---|
| Produk menyesatkan atau terlarang | `UPDATE products SET is_archived = true, updated_at = now() WHERE id = '<id>';` |
| Kebutuhan tidak pantas | `UPDATE needs SET status = 'CLOSED', updated_at = now() WHERE id = '<id>' AND status = 'OPEN';` |
| Organisasi melanggar berulang | `UPDATE organizations SET status = 'SUSPENDED', updated_at = now() WHERE id = '<id>';` lalu cabut sesinya dengan `UPDATE sessions SET revoked_at = now() WHERE revoked_at IS NULL AND user_id IN (SELECT user_id FROM memberships WHERE organization_id = '<id>');` |

Organisasi yang ditangguhkan tidak tampil di Cari Mitra dan tidak bisa mengirim Ajuan, tetapi tetap bisa menyelesaikan titipan, hitungan, dan Setoran yang berjalan (BR-AUTHZ-06). Untuk mengaktifkan kembali, ubah `status` menjadi `ACTIVE`.

## 3. Menutup laporan

```sql
UPDATE listing_reports
SET status = 'RESOLVED', resolved_by = '<id pengguna tim>', resolved_at = now(), updated_at = now()
WHERE id = '<id laporan>';
```

Pakai `DISMISSED` bila laporan tidak terbukti. Laporan tidak pernah dihapus.
