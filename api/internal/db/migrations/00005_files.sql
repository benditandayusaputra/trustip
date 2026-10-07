-- +goose Up
CREATE TABLE files (
    id uuid PRIMARY KEY,
    organization_id uuid NOT NULL REFERENCES organizations (id),
    purpose text NOT NULL,
    visibility text NOT NULL,
    content_type text NOT NULL,
    size_bytes integer NOT NULL,
    storage_key text NOT NULL,
    sha256 bytea NOT NULL,
    width integer NOT NULL,
    height integer NOT NULL,
    created_by uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT files_storage_key_key UNIQUE (storage_key),
    CONSTRAINT files_purpose_known CHECK (purpose IN ('LOGO', 'STORE_PHOTO', 'PAYMENT_QR')),
    CONSTRAINT files_visibility_known CHECK (visibility IN ('PUBLIC', 'PRIVATE')),
    CONSTRAINT files_content_type_image CHECK (content_type IN ('image/jpeg', 'image/png')),
    CONSTRAINT files_size_limit CHECK (size_bytes BETWEEN 1 AND 1048576),
    CONSTRAINT files_dimension_limit CHECK (width BETWEEN 1 AND 4096 AND height BETWEEN 1 AND 4096),
    CONSTRAINT files_sha256_length CHECK (octet_length(sha256) = 32)
);

CREATE INDEX files_organization_id_idx ON files (organization_id, created_at);

-- +goose Down
DROP TABLE files;
