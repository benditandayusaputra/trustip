-- name: CreateUser :exec
INSERT INTO users (
    id, email, password_hash, pepper_version, full_name, whatsapp,
    privacy_accepted_at, created_at, updated_at
) VALUES (
    @id, @email, @password_hash, @pepper_version, @full_name, @whatsapp,
    @privacy_accepted_at, @privacy_accepted_at, @privacy_accepted_at
);

-- name: FindUserByEmail :one
SELECT id, email, password_hash, pepper_version, full_name, whatsapp,
    failed_login_count, locked_until, privacy_accepted_at
FROM users
WHERE email = @email;
