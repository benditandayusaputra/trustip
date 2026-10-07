-- name: CreateOrganization :exec
INSERT INTO organizations (id, type, name, slug, whatsapp, created_at, updated_at)
VALUES (@id, @type, @name, @slug, @whatsapp, @created_at, @created_at);

-- name: CreateMembership :exec
INSERT INTO memberships (id, organization_id, user_id, role, created_at, updated_at)
VALUES (@id, @organization_id, @user_id, @role, @created_at, @created_at);

-- name: FindActiveMembership :one
SELECT m.organization_id, o.type AS organization_type, m.role
FROM memberships AS m
JOIN organizations AS o ON o.id = m.organization_id
WHERE m.user_id = @user_id AND m.is_active;
