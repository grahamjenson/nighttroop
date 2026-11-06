-- name: GetSchemaCompatibility :one
SELECT oldest_version FROM schema_compatibility WHERE version = ?;

-- name: SetSchemaCompatibility :exec
INSERT OR REPLACE INTO schema_compatibility (version, oldest_version) VALUES (?, ?);
