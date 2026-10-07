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

-- name: LockLoginAttempts :one
SELECT failed_login_count, locked_until
FROM users
WHERE id = @id
FOR UPDATE;

-- name: SaveLoginAttempts :exec
UPDATE users
SET failed_login_count = @failed_login_count, locked_until = @locked_until, updated_at = now()
WHERE id = @id;

-- name: SavePassword :exec
UPDATE users
SET password_hash = @password_hash, pepper_version = @pepper_version, updated_at = now()
WHERE id = @id;

-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, family_id, token_hash, expires_at, created_at)
VALUES (@id, @user_id, @family_id, @token_hash, @expires_at, @created_at);
