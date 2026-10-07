-- name: RecordAudit :exec
INSERT INTO audit_logs (id, organization_id, actor_id, action, entity_type, entity_id, metadata, ip, created_at)
VALUES (
    @id::uuid, NULLIF(@organization_id::text, '')::uuid, @actor_id::uuid, @action::text, @entity_type::text,
    @entity_id::uuid, @metadata::jsonb, @ip::text, @created_at::timestamptz
);

-- name: ListAuditActions :many
SELECT action, entity_type, COALESCE(organization_id::text, '')::text AS organization_id, ip
FROM audit_logs
WHERE actor_id = @actor_id
ORDER BY created_at, id;
