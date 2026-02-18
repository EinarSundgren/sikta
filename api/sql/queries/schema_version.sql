-- name: ListSchemaVersions :many
SELECT id, version, applied_at
FROM schema_versions
ORDER BY id;
