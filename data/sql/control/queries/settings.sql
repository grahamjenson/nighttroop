-- name: getSettingFromDB :one
SELECT * FROM settings WHERE subdomain = ? AND key = ?;

-- name: upsertSettingToDB :one
INSERT INTO settings(subdomain, key, value, created_at)
VALUES (?, ?, json(?), CAST(unixepoch('subsec') * 1000 AS INTEGER))
ON CONFLICT(subdomain, key) DO UPDATE SET
    value = excluded.value,
    updated_at = CAST(unixepoch('subsec') * 1000 AS INTEGER)
RETURNING *;