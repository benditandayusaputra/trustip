-- name: GetServerVersion :one
SELECT current_setting('server_version') AS server_version;

-- name: ListAvailableExtensions :many
SELECT name, default_version
FROM pg_catalog.pg_available_extensions
WHERE name = ANY(sqlc.arg(names)::text[])
ORDER BY name;

-- name: ListInstalledExtensions :many
SELECT extname::text AS name
FROM pg_catalog.pg_extension
WHERE extname = ANY(sqlc.arg(names)::text[])
ORDER BY extname;

-- name: IsSchemaOwner :one
SELECT pg_has_role(current_user, n.nspowner, 'MEMBER') AS is_owner
FROM pg_catalog.pg_namespace AS n
WHERE n.nspname = 'public';

-- name: GetCurrentTransactionID :one
SELECT pg_current_xact_id()::text AS transaction_id;
