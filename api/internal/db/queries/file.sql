-- name: CreateFile :exec
INSERT INTO files (
    id, organization_id, purpose, visibility, content_type, size_bytes, storage_key, sha256,
    width, height, created_by, created_at
) VALUES (
    @id, @organization_id, @purpose, @visibility, @content_type, @size_bytes, @storage_key, @sha256,
    @width, @height, @created_by, @created_at
);

-- name: GetOrganizationFile :one
SELECT f.id, f.organization_id, f.purpose, f.visibility, f.content_type, f.size_bytes, f.storage_key, f.sha256,
    f.width, f.height, f.created_by, f.created_at
FROM files AS f
WHERE f.id = @id::uuid AND (
    f.organization_id = @organization_id::uuid
    OR EXISTS (
        SELECT 1
        FROM payments AS p
        JOIN settlements AS st ON st.id = p.settlement_id
        WHERE p.proof_file_id = f.id
            AND (st.consignor_org_id = @organization_id::uuid OR st.store_org_id = @organization_id::uuid)
    )
    OR EXISTS (
        SELECT 1
        FROM organizations AS o
        JOIN partnerships AS pa ON pa.consignor_org_id = o.id AND pa.status = 'ACTIVE'
        WHERE o.payment_qr_file_id = f.id AND pa.store_org_id = @organization_id::uuid
    )
);
