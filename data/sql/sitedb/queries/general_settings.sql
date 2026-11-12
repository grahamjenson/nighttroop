-- name: UpsertGeneralSetting :one
INSERT INTO general_settings (key, type, value_json, created_by_identity_name) VALUES (?, ?, ?, ?)
ON CONFLICT(key) DO UPDATE SET
    type = excluded.type,
    value_json = excluded.value_json,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
RETURNING *;

-- name: GetGeneralSetting :one
SELECT * FROM general_settings WHERE key = ?;

-- name: SearchGeneralSettings :many
SELECT * FROM general_settings WHERE updated_at >= sqlc.arg('UpdatedAtGTE') ORDER BY updated_at DESC LIMIT CAST(sqlc.arg('Limit') AS INTEGER);

-- name: CountGeneralSettings :one
SELECT COUNT(*) FROM general_settings;

-- name: DeleteGeneralSetting :one
DELETE FROM general_settings WHERE key = ? RETURNING *;

-- name: DeleteGeneralSettingsBefore :exec
DELETE FROM general_settings WHERE type = ? AND updated_at < ?;
