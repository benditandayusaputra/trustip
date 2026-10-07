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
