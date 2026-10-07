-- name: CreateFile :exec
INSERT INTO files (
    id, organization_id, purpose, visibility, content_type, size_bytes, storage_key, sha256,
    width, height, created_by, created_at
) VALUES (
    @id, @organization_id, @purpose, @visibility, @content_type, @size_bytes, @storage_key, @sha256,
    @width, @height, @created_by, @created_at
);

-- name: GetOrganizationFile :one
SELECT id, organization_id, purpose, visibility, content_type, size_bytes, storage_key, sha256,
    width, height, created_by, created_at
FROM files
WHERE id = @id AND organization_id = @organization_id;
