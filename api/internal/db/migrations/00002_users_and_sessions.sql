-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY,
    email text NOT NULL,
    password_hash text NOT NULL,
    pepper_version text NOT NULL,
    full_name text NOT NULL,
    whatsapp text NOT NULL,
    is_platform_admin boolean NOT NULL DEFAULT false,
    failed_login_count integer NOT NULL DEFAULT 0,
    locked_until timestamptz,
    privacy_accepted_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_email_key UNIQUE (email),
    CONSTRAINT users_email_lowercase CHECK (email = lower(email)),
    CONSTRAINT users_email_length CHECK (char_length(email) BETWEEN 3 AND 254),
    CONSTRAINT users_password_hash_argon2id CHECK (password_hash LIKE '$argon2id$%'),
    CONSTRAINT users_pepper_version_format CHECK (pepper_version ~ '^v[1-9][0-9]*$'),
    CONSTRAINT users_full_name_length CHECK (char_length(full_name) BETWEEN 2 AND 80),
    CONSTRAINT users_whatsapp_format CHECK (whatsapp ~ '^62[0-9]{8,13}$'),
    CONSTRAINT users_failed_login_count_nonnegative CHECK (failed_login_count >= 0)
);

CREATE TABLE sessions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    family_id uuid NOT NULL,
    token_hash bytea NOT NULL,
    expires_at timestamptz NOT NULL,
    rotated_at timestamptz,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT sessions_token_hash_key UNIQUE (token_hash),
    CONSTRAINT sessions_token_hash_sha256 CHECK (octet_length(token_hash) = 32),
    CONSTRAINT sessions_expires_after_creation CHECK (expires_at > created_at)
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);
CREATE INDEX sessions_family_id_idx ON sessions (family_id);

-- +goose Down
DROP TABLE sessions;
DROP TABLE users;
