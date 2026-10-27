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

-- name: LockSessionByTokenHash :one
SELECT id, user_id, family_id, expires_at, rotated_at, revoked_at
FROM sessions
WHERE token_hash = @token_hash
FOR UPDATE;

-- name: MarkSessionRotated :exec
UPDATE sessions
SET rotated_at = @rotated_at
WHERE id = @id;

-- name: RevokeSessionFamily :exec
UPDATE sessions
SET revoked_at = @revoked_at
WHERE family_id = @family_id AND revoked_at IS NULL;

-- name: RevokeUserSessions :many
UPDATE sessions
SET revoked_at = @revoked_at
WHERE user_id = @user_id AND revoked_at IS NULL
RETURNING family_id;

-- name: GetAccount :one
SELECT u.id, u.email, u.full_name, u.whatsapp, COALESCE(f.id::text, '')::text AS avatar_file_id
FROM users AS u
LEFT JOIN files AS f ON f.id = u.avatar_file_id
WHERE u.id = @id;

-- name: CountOwnAvatarFiles :one
SELECT count(*)::integer
FROM files
WHERE id = @id::uuid AND organization_id = @organization_id::uuid AND created_by = @created_by::uuid
    AND purpose = 'AVATAR';

-- name: UpdateAccountProfile :execrows
UPDATE users
SET
    full_name = @full_name::text,
    whatsapp = @whatsapp::text,
    avatar_file_id = NULLIF(@avatar_file_id::text, '')::uuid,
    updated_at = @updated_at::timestamptz
WHERE id = @id::uuid;

-- name: LockPasswordOwner :one
SELECT email, full_name, password_hash, pepper_version
FROM users
WHERE id = @id
FOR UPDATE;

-- name: RevokeOtherUserSessions :many
UPDATE sessions
SET revoked_at = @revoked_at
WHERE user_id = @user_id AND family_id <> @keep_family_id::uuid AND revoked_at IS NULL
RETURNING family_id;
