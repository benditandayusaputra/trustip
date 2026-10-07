-- name: DeliverNotifications :exec
INSERT INTO notifications (id, user_id, type, title, body, link, created_at)
SELECT
    unnest(@ids::uuid[]), unnest(@user_ids::uuid[]),
    @type::text, @title::text, @body::text, @link::text, @created_at::timestamptz;

-- name: ListNotifications :many
SELECT
    id, type, title, body, link, (read_at IS NOT NULL)::boolean AS is_read,
    COALESCE(read_at, created_at)::timestamptz AS read_at, created_at
FROM notifications
WHERE user_id = @user_id::uuid
    AND (@from_start::boolean OR (created_at, id) < (@after_at::timestamptz, @after_id::uuid))
ORDER BY created_at DESC, id DESC
LIMIT @row_limit::bigint;

-- name: CountUnreadNotifications :one
SELECT count(*)::bigint
FROM notifications
WHERE user_id = @user_id AND read_at IS NULL;

-- name: MarkNotificationRead :execrows
UPDATE notifications
SET read_at = COALESCE(read_at, @read_at::timestamptz)
WHERE id = @id::uuid AND user_id = @user_id::uuid;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET read_at = @read_at::timestamptz
WHERE user_id = @user_id::uuid AND read_at IS NULL;

-- name: ListOrganizationRecipients :many
SELECT user_id, role
FROM memberships
WHERE organization_id = @organization_id::uuid AND is_active
ORDER BY created_at, id;
