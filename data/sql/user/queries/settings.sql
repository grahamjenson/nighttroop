-- name: getSettingFromDB :one
SELECT value
FROM settings
WHERE id = ?;

-- name: upsertSettingToDB :exec
INSERT INTO settings(id, value, created_at)
VALUES (?, json(?), CAST(unixepoch('subsec') * 1000 AS INTEGER))
ON CONFLICT(id) DO UPDATE SET
    value = excluded.value,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)